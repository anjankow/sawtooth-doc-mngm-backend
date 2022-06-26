package user

import (
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

type User struct {
	ID         string
	Name       string
	PrivateKey string
	PublicKey  string
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

func (m UserManager) GetUserByID(userID string) (User, error) {
	path := graphURL + m.tenantID + "/users/" + userID +
		"?$select=userPrincipalName," +
		"extension_" + m.extensionID + "_PrivateKey," +
		"extension_" + m.extensionID + "_PublicKey"

	r, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return User{}, err
	}
	r.Header.Add("Authorization", "Bearer "+m.tokenGuard.token)

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return User{}, err
	}

	defer resp.Body.Close()
	reponseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return User{}, errors.New("reading response error: " + err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		// if unauthorized, set the new token and try again
		if resp.StatusCode == http.StatusUnauthorized {
			if err := m.setNewAppToken(); err != nil {
				return User{}, errors.New("token not valid, failed to set a new one: " + err.Error())
			}
			return m.GetUserByID(userID)
		}

		return User{}, errors.New("status code: " + resp.Status + "; body: " + string(reponseBody))
	}

	modifiedBody := strings.ReplaceAll(string(reponseBody), "extension_66a58d4966864ccabd3e5f966101de97_", "")

	var unmarshalled struct {
		Name       string `json:"userPrincipalName"`
		PrivateKey string `json:"PrivateKey"`
		PublicKey  string `json:"PublicKey"`
	}
	if err := json.Unmarshal([]byte(modifiedBody), &unmarshalled); err != nil {
		return User{}, errors.New("failed to unmarshal the response: " + err.Error())
	}

	return User{
		ID:         userID,
		Name:       unmarshalled.Name,
		PrivateKey: unmarshalled.PrivateKey,
		PublicKey:  unmarshalled.PublicKey,
	}, nil
}
