package user

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	loginDomain = "https://login.microsoftonline.com/"
)

func (m UserManager) GetAppToken() (string, error) {

	path := loginDomain + m.tenantID + "/oauth2/v2.0/token"
	data := url.Values{}
	data.Set("client_id", m.clientID)
	data.Set("client_secret", m.secret)
	data.Set("scope", "https://graph.microsoft.com/.default")
	data.Set("grant_type", "client_credentials")

	r, err := http.NewRequest(http.MethodGet, path, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	reponseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New("reading response error: " + err.Error())
	}

	var unmarshalled struct {
		Token string `json:"access_token"`
	}
	if err := json.Unmarshal(reponseBody, &unmarshalled); err != nil {
		return "", errors.New("failed to unmarshal the response: " + err.Error())
	}

	return unmarshalled.Token, nil
}
