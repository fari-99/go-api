package keycloaks

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-resty/resty/v2"

	keycloaks "go-api/pkg/keycloaks/models"
)

type OpenIDHelper struct {
	AccessToken string
}

func NewOpenIDHelper() *OpenIDHelper {
	return &OpenIDHelper{}
}

func (base *OpenIDHelper) SetAccessToken(accessToken string) *OpenIDHelper {
	base.AccessToken = accessToken
	return base
}

func (base *OpenIDHelper) GetAccessToken(input keycloaks.AuthenticateKeycloak) (*keycloaks.AuthenticateResponse, error) {
	accessToken := base.AccessToken

	// if accessToken == "" {
	// 	return fmt.Errorf("access token is empty")
	// }

	var formInput map[string]string
	dataMarshal, _ := json.Marshal(input)
	_ = json.Unmarshal(dataMarshal, &formInput)

	keycloakUrl := os.Getenv("KEYCLOAK_HOST")
	if os.Getenv("KEYCLOAK_PORT") != "" {
		keycloakUrl += ":" + os.Getenv("KEYCLOAK_PORT")
	}

	keycloakUrl = fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token",
		keycloakUrl, os.Getenv("USER_KEYCLOAK_REALM"))

	client := resty.New()
	resp, err := client.R().
		SetDebug(true).
		SetAuthToken(accessToken).
		SetFormData(formInput).
		Post(keycloakUrl)

	if err != nil {
		return nil, err
	}

	var resResp keycloaks.AuthenticateResponse
	_ = json.Unmarshal(resp.Body(), &resResp)

	if resResp.Error != "" {
		return nil, fmt.Errorf(resResp.ErrorDescription)
	}

	return &resResp, nil
}

func (base *OpenIDHelper) Logout(input keycloaks.LogoutKeycloak) (interface{}, error) {
	accessToken := base.AccessToken

	// if accessToken == "" {
	// 	return fmt.Errorf("access token is empty")
	// }

	var formInput map[string]string
	dataMarshal, _ := json.Marshal(input)
	_ = json.Unmarshal(dataMarshal, &formInput)

	keycloakUrl := os.Getenv("KEYCLOAK_HOST")
	if os.Getenv("KEYCLOAK_PORT") != "" {
		keycloakUrl += ":" + os.Getenv("KEYCLOAK_PORT")
	}

	keycloakUrl = fmt.Sprintf("%s/realms/%s/protocol/openid-connect/logout",
		keycloakUrl, os.Getenv("USER_KEYCLOAK_REALM"))

	client := resty.New()
	resp, err := client.R().
		SetDebug(true).
		SetAuthToken(accessToken).
		SetFormData(formInput).
		Post(keycloakUrl)

	if err != nil {
		return nil, err
	}

	var resResp keycloaks.AuthenticateResponse
	_ = json.Unmarshal(resp.Body(), &resResp)

	if resResp.Error != "" {
		return nil, fmt.Errorf(resResp.ErrorDescription)
	}

	return &resResp, nil
}
