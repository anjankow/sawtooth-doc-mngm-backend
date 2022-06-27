package usermanager

import (
	"context"
	"doc-management/internal/model"
	"doc-management/internal/signkeys"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

type TokenGuard struct {
	token string
}

type UserManager struct {
	tenantID    string
	clientID    string
	extensionID string

	secret     string
	tokenGuard *TokenGuard
}

const (
	graphURL = "https://graph.microsoft.com/v1.0/"
)

func NewUserManager(tenantID, clientID, extensionID, secret string) (UserManager, error) {

	manager := UserManager{
		tenantID:    tenantID,
		clientID:    clientID,
		extensionID: extensionID,
		secret:      secret,

		tokenGuard: &TokenGuard{token: ""},
	}

	err := manager.setNewAppToken()
	if err != nil {
		return UserManager{}, err
	}

	return manager, nil
}

// InitAndReadAppKeys reads the app keys from the AD, sets them if they don't exist yet
func (m UserManager) InitAndReadAppKeys(ctx context.Context, appUserID string) (keys signkeys.UserKeys, err error) {
	user, err := m.GetUserByID(ctx, appUserID)
	if err != nil {
		return
	}

	return user.Keys, nil
}

// GetUserByID gets the user info from the AD and updates the user's keys if not assigned yet
func (m UserManager) GetUserByID(ctx context.Context, userID string) (model.User, error) {
	user, err := m.getUserByID(ctx, userID)
	if err != nil {
		return model.User{}, err
	}

	if user.HasValidKeys() {
		return user, nil
	}

	// if the keys are not initialized, update the user's profile with new keys
	keys, err := signkeys.GenerateKeys()
	if err != nil {
		return model.User{}, errors.New("failed to generate user's keys: " + err.Error())
	}

	return m.updateUserKeys(ctx, user, keys)
}

func (m UserManager) updateUserKeys(ctx context.Context, user model.User, userKeys signkeys.UserKeys) (model.User, error) {
	path := graphURL + m.tenantID + "/users/" + user.ID

	var updateBody struct {
		PrivateKey string `json:"extension_EXTENSION_ID_PrivateKey"`
		PublicKey  string `json:"extension_EXTENSION_ID_PublicKey"`
	}

	updatedUser := user.SetKeys(userKeys)
	updateBody.PrivateKey = updatedUser.Keys.PrivateKey.AsHex()
	updateBody.PublicKey = updatedUser.Keys.PrivateKey.AsHex()

	marshalledReqBody, err := json.Marshal(updateBody)
	if err != nil {
		return user, err
	}

	modifiedBody := strings.ReplaceAll(string(marshalledReqBody), "EXTENSION_ID", m.extensionID)
	r, err := http.NewRequestWithContext(ctx, http.MethodPatch, path, strings.NewReader(modifiedBody))
	if err != nil {
		return user, err
	}
	r.Header.Add("Authorization", "Bearer "+m.tokenGuard.token)
	r.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return user, err
	}

	if isResponseSuccess(resp.StatusCode) {
		// on successful response, nothing more to do, return
		return updatedUser, nil
	}

	// if unauthorized, set the new token and try again
	if resp.StatusCode == http.StatusUnauthorized {
		if err := m.setNewAppToken(); err != nil {
			return user, errors.New("token not valid, failed to set a new one: " + err.Error())
		}
		return m.updateUserKeys(ctx, user, userKeys)
	}

	// on any other not successfull status code read the response body to get the error
	defer resp.Body.Close()
	reponseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		reponseBody = []byte("failed to read the response body: " + err.Error())
	}

	return user, errors.New("update user's keys, status code: " + resp.Status + "; " + string(reponseBody))

}

func (m UserManager) getUserByID(ctx context.Context, userID string) (model.User, error) {
	path := graphURL + m.tenantID + "/users/" + userID +
		"?$select=userPrincipalName," +
		"extension_" + m.extensionID + "_PrivateKey," +
		"extension_" + m.extensionID + "_PublicKey"

	r, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return model.User{}, err
	}
	r.Header.Add("Authorization", "Bearer "+m.tokenGuard.token)

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return model.User{}, err
	}

	defer resp.Body.Close()
	reponseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return model.User{}, errors.New("reading response error: " + err.Error())
	}

	if !isResponseSuccess(resp.StatusCode) {
		// if unauthorized, set the new token and try again
		if resp.StatusCode == http.StatusUnauthorized {
			if err := m.setNewAppToken(); err != nil {
				return model.User{}, errors.New("token not valid, failed to set a new one: " + err.Error())
			}
			return m.GetUserByID(ctx, userID)
		}

		return model.User{}, errors.New("status code: " + resp.Status + "; body: " + string(reponseBody))
	}

	modifiedBody := strings.ReplaceAll(string(reponseBody), "extension_"+m.extensionID+"_", "")

	var unmarshalled struct {
		Name       string `json:"userPrincipalName"`
		PrivateKey string `json:"PrivateKey"`
		PublicKey  string `json:"PublicKey"`
	}
	if err := json.Unmarshal([]byte(modifiedBody), &unmarshalled); err != nil {
		return model.User{}, errors.New("failed to unmarshal the response: " + err.Error())
	}

	keys, err := signkeys.NewUserKeys(unmarshalled.PrivateKey, unmarshalled.PublicKey)
	if err != nil {
		return model.User{}, errors.New("failed to parse the keys: " + err.Error())
	}
	return model.User{
		ID:   userID,
		Name: unmarshalled.Name,
		Keys: keys,
	}, nil
}

func isResponseSuccess(responseCode int) bool {
	return responseCode >= 200 && responseCode < 300
}
