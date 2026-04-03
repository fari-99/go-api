package keycloaks

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/redis/go-redis/v9"

	keycloakConstants "go-api/pkg/keycloaks/constants"
	keycloakModels "go-api/pkg/keycloaks/models"
)

type AdminApiKeycloak struct {
	Realms      string        `json:"realms"`
	AdminApi    bool          `json:"admin_api"`
	RedisClient *redis.Client `json:"redis_client"`

	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type TokenResponse struct {
	keycloakModels.ErrorResponse

	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	IdToken          string `json:"id_token"`
	NotBeforePolicy  int    `json:"not-before-policy"`
	SessionState     string `json:"session_state"`
	Scope            string `json:"scope"`
}

func getClientResty() *resty.Request {
	client := resty.New()
	isDebugMode := os.Getenv("APP_MODE") == "debug"
	return client.R().SetDebug(isDebugMode)
}

func getUrlRealms(realms string, adminApi bool) string {
	url := os.Getenv("KEYCLOAK_HOST")
	if os.Getenv("KEYCLOAK_PORT") != "" {
		url += ":" + os.Getenv("KEYCLOAK_PORT")
	}

	if adminApi {
		url += "/admin"
	}

	url += "/realms"

	switch realms {
	case os.Getenv("ADMIN_KEYCLOAK_REALM"):
		url += "/" + os.Getenv("ADMIN_KEYCLOAK_REALM")
	case os.Getenv("USER_KEYCLOAK_REALM"):
		url += "/" + os.Getenv("USER_KEYCLOAK_REALM")
	default:
		panic(realms + "not set yet")
	}

	return url
}

func NewAdminApi(realms string) (*AdminApiKeycloak, error) {
	if exists, err := RealmExists(realms); err != nil {
		return nil, err
	} else if !exists {
		return nil, fmt.Errorf("realms not exists")
	}

	respToken, err := getAccessToken(realms, false) // getting access token doesn't use /admin/
	if err != nil {
		return nil, err
	}

	adminApiKeycloak := &AdminApiKeycloak{
		Realms:       realms,
		AdminApi:     true, // this is admin api, so default must be true
		AccessToken:  respToken.AccessToken,
		RefreshToken: respToken.RefreshToken,
	}

	return adminApiKeycloak, nil
}

func getAccessToken(realms string, adminApi bool) (respToken *TokenResponse, err error) {
	url := getUrlRealms(realms, adminApi)
	url += "/protocol/openid-connect/token"

	// TODO: using redis cache to save token

	resp, err := getClientResty().
		SetFormData(map[string]string{
			"grant_type":    "client_credentials",
			"client_id":     os.Getenv("ROBIN_CLIENT_ID"),
			"client_secret": os.Getenv("ROBIN_CLIENT_SECRET"),
			"scope":         "openid",
		}).
		Post(url)
	if err != nil {
		return nil, err
	}

	var tokenResponse TokenResponse
	err = json.Unmarshal(resp.Body(), &tokenResponse)
	if err != nil {
		return nil, err
	}

	if tokenResponse.ErrorDescription != "" {
		return nil, fmt.Errorf(tokenResponse.ErrorDescription)
	} else if tokenResponse.ErrorMessage != "" {
		return nil, fmt.Errorf(tokenResponse.ErrorMessage)
	} else if tokenResponse.Error != "" {
		return nil, fmt.Errorf(tokenResponse.Error)
	}

	return &tokenResponse, nil
}

func (adminApi *AdminApiKeycloak) GetUserDetails(userID string) (*keycloakModels.UserDetails, error) {
	url := getUrlRealms(adminApi.Realms, adminApi.AdminApi)
	url += "/users/" + userID

	resp, err := getClientResty().
		SetHeaders(map[string]string{
			"Authorization": "Bearer " + adminApi.AccessToken,
			"Accept":        "application/json",
		}).
		Get(url)
	if err != nil {
		return nil, err
	}

	var userDetails keycloakModels.UserDetails
	_ = json.Unmarshal(resp.Body(), &userDetails)

	if userDetails.ErrorDescription != "" {
		return nil, fmt.Errorf(userDetails.ErrorDescription)
	} else if userDetails.ErrorMessage != "" {
		return nil, fmt.Errorf(userDetails.ErrorMessage)
	} else if userDetails.Error != "" {
		return nil, fmt.Errorf(userDetails.Error)
	}

	return &userDetails, nil
}

func (adminApi *AdminApiKeycloak) UpdateUser(userID string, input keycloakModels.InputUpdateUser) error {
	url := getUrlRealms(adminApi.Realms, adminApi.AdminApi)
	url += "/users/" + userID

	resp, err := getClientResty().
		SetHeaders(map[string]string{
			"Authorization": "Bearer " + adminApi.AccessToken,
			"Accept":        "application/json",
		}).
		SetBody(input).
		Put(url)
	if err != nil {
		return err
	}

	var checkError keycloakModels.ErrorResponse
	_ = json.Unmarshal(resp.Body(), &checkError)

	if checkError.ErrorDescription != "" {
		return fmt.Errorf(checkError.ErrorDescription)
	} else if checkError.ErrorMessage != "" {
		return fmt.Errorf(checkError.ErrorMessage)
	} else if checkError.Error != "" {
		return fmt.Errorf(checkError.Error)
	}

	return nil
}

func (adminApi *AdminApiKeycloak) CreateUser(input keycloakModels.InputCreateUser) (userID string, err error) {
	url := getUrlRealms(adminApi.Realms, adminApi.AdminApi)
	url += "/users"

	resp, err := getClientResty().
		SetHeaders(map[string]string{
			"Authorization": "Bearer " + adminApi.AccessToken,
			"Accept":        "application/json",
		}).
		SetBody(input).
		Post(url)
	if err != nil {
		return "", err
	}

	var checkError keycloakModels.ErrorResponse
	_ = json.Unmarshal(resp.Body(), &checkError)

	if checkError.ErrorDescription != "" {
		return "", fmt.Errorf(checkError.ErrorDescription)
	} else if checkError.ErrorMessage != "" {
		return "", fmt.Errorf(checkError.ErrorMessage)
	} else if checkError.Error != "" {
		return "", fmt.Errorf(checkError.Error)
	}

	// RESP : 201 Created with NO CONTENT
	// So we get it from Location response headers
	// ex response header from keycloak: http://keycloak:8080/admin/realms/master/users/<CLIENT SCOPES ID>
	locationUrl, err := resp.RawResponse.Location()
	if errors.Is(err, http.ErrNoLocation) {
		return "", fmt.Errorf("failed to get client scopes ID")
	}

	locationUrls := strings.Split(locationUrl.String(), "/")
	return locationUrls[len(locationUrls)-1], nil
}

func (adminApi *AdminApiKeycloak) ListUser(input keycloakModels.FilterListUser) (*keycloakModels.ListUser, error) {
	url := getUrlRealms(adminApi.Realms, adminApi.AdminApi)
	url += "/users"

	var queryParams map[string]string
	inputMarshal, _ := json.Marshal(input)
	_ = json.Unmarshal(inputMarshal, &queryParams)

	resp, err := getClientResty().
		SetHeaders(map[string]string{
			"Authorization": "Bearer " + adminApi.AccessToken,
			"Accept":        "application/json",
		}).
		SetQueryParams(queryParams).
		Get(url)
	if err != nil {
		return nil, err
	}

	responseBody := resp.Body()

	var checkError keycloakModels.ErrorResponse
	_ = json.Unmarshal(responseBody, &checkError)

	if checkError.ErrorDescription != "" {
		return nil, fmt.Errorf(checkError.ErrorDescription)
	} else if checkError.ErrorMessage != "" {
		return nil, fmt.Errorf(checkError.ErrorMessage)
	} else if checkError.Error != "" {
		return nil, fmt.Errorf(checkError.Error)
	}

	var listUsers []keycloakModels.ListUser
	_ = json.Unmarshal(responseBody, &listUsers)

	if len(listUsers) == 0 {
		return nil, fmt.Errorf("record not found")
	}

	return &listUsers[0], nil
}

func (adminApi *AdminApiKeycloak) SetPassword(userID string, input keycloakModels.InputSetPassword) error {
	url := getUrlRealms(adminApi.Realms, adminApi.AdminApi)
	url += "/users/" + userID + "/reset-password"

	resp, err := getClientResty().
		SetHeaders(map[string]string{
			"Authorization": "Bearer " + adminApi.AccessToken,
		}).
		SetBody(input).
		Put(url)
	if err != nil {
		return err
	}

	var checkError keycloakModels.ErrorResponse
	_ = json.Unmarshal(resp.Body(), &checkError)

	if checkError.ErrorDescription != "" {
		return fmt.Errorf(checkError.ErrorDescription)
	} else if checkError.ErrorMessage != "" {
		return fmt.Errorf(checkError.ErrorMessage)
	} else if checkError.Error != "" {
		return fmt.Errorf(checkError.Error)
	}

	// set password temp response is empty (204 no response)
	return nil
}

func (adminApi *AdminApiKeycloak) SendVerifyEmail(userID string, input *keycloakModels.VerifyEmail) error {
	url := getUrlRealms(adminApi.Realms, adminApi.AdminApi)
	url += "/users/" + userID + "/send-verify-email"

	var queryParams map[string]string
	if input != nil {
		inputMarshal, _ := json.Marshal(input)
		_ = json.Unmarshal(inputMarshal, &queryParams)
	}

	resp, err := getClientResty().
		SetHeaders(map[string]string{
			"Authorization": "Bearer " + adminApi.AccessToken,
		}).
		SetQueryParams(queryParams).
		Put(url)
	if err != nil {
		return err
	}

	var checkError keycloakModels.ErrorResponse
	_ = json.Unmarshal(resp.Body(), &checkError)

	if checkError.ErrorDescription != "" {
		return fmt.Errorf(checkError.ErrorDescription)
	} else if checkError.ErrorMessage != "" {
		return fmt.Errorf(checkError.ErrorMessage)
	} else if checkError.Error != "" {
		return fmt.Errorf(checkError.Error)
	}

	// send verify email response is empty (204 no response)
	return nil

}

func (adminApi *AdminApiKeycloak) CreateClient(input keycloakModels.InputCreateClients) (string, error) {
	url := getUrlRealms(adminApi.Realms, adminApi.AdminApi)
	url += "/clients"

	resp, err := getClientResty().
		SetHeaders(map[string]string{
			"Authorization": "Bearer " + adminApi.AccessToken,
		}).
		SetBody(input).
		Post(url)
	if err != nil {
		return "", err
	}

	var checkError keycloakModels.ErrorResponse
	_ = json.Unmarshal(resp.Body(), &checkError)

	if checkError.ErrorDescription != "" {
		return "", fmt.Errorf(checkError.ErrorDescription)
	} else if checkError.ErrorMessage != "" {
		return "", fmt.Errorf(checkError.ErrorMessage)
	} else if checkError.Error != "" {
		return "", fmt.Errorf(checkError.Error)
	}

	// RESP : 201 Created with NO CONTENT
	// So we get it from Location response headers
	// ex response header from keycloak: http://keycloak:8080/admin/realms/master/client-scopes/<CLIENT SCOPES ID>
	locationUrl, err := resp.RawResponse.Location()
	if errors.Is(err, http.ErrNoLocation) {
		return "", fmt.Errorf("failed to get client scopes ID")
	}

	locationUrls := strings.Split(locationUrl.String(), "/")
	return locationUrls[len(locationUrls)-1], nil
}

func (adminApi *AdminApiKeycloak) CreateClientScopes(input keycloakModels.InputCreateClientScopes) (string, error) {
	url := getUrlRealms(adminApi.Realms, adminApi.AdminApi)
	url += "/client-scopes"

	resp, err := getClientResty().
		SetHeaders(map[string]string{
			"Authorization": "Bearer " + adminApi.AccessToken,
		}).
		SetBody(input).
		Post(url)
	if err != nil {
		return "", err
	}

	var checkError keycloakModels.ErrorResponse
	_ = json.Unmarshal(resp.Body(), &checkError)

	if checkError.ErrorDescription != "" {
		return "", fmt.Errorf(checkError.ErrorDescription)
	} else if checkError.ErrorMessage != "" {
		return "", fmt.Errorf(checkError.ErrorMessage)
	} else if checkError.Error != "" {
		return "", fmt.Errorf(checkError.Error)
	}

	// RESP : 201 Created with NO CONTENT
	// So we get it from Location response headers
	// ex response header from keycloak: http://keycloak:8080/admin/realms/master/client-scopes/<CLIENT SCOPES ID>
	locationUrl, err := resp.RawResponse.Location()
	if errors.Is(err, http.ErrNoLocation) {
		return "", fmt.Errorf("failed to get client scopes ID")
	}

	locationUrls := strings.Split(locationUrl.String(), "/")
	return locationUrls[len(locationUrls)-1], nil
}

func (adminApi *AdminApiKeycloak) AddClientScopes(scopeType string, clientID string, clientScopeID string) error {
	url := getUrlRealms(adminApi.Realms, adminApi.AdminApi)

	switch scopeType {
	case keycloakConstants.ClientScopesDefault:
		url += fmt.Sprintf("/clients/%s/default-client-scopes/%s",
			clientID, clientScopeID)
	case keycloakConstants.ClientScopesOptional:
		url += fmt.Sprintf("/clients/%s/optional-client-scopes/%s",
			clientID, clientScopeID)
	}

	resp, err := getClientResty().
		SetHeaders(map[string]string{
			"Authorization": "Bearer " + adminApi.AccessToken,
		}).
		Put(url)
	if err != nil {
		return err
	}

	var checkError keycloakModels.ErrorResponse
	_ = json.Unmarshal(resp.Body(), &checkError)

	if checkError.ErrorDescription != "" {
		return fmt.Errorf(checkError.ErrorDescription)
	} else if checkError.ErrorMessage != "" {
		return fmt.Errorf(checkError.ErrorMessage)
	} else if checkError.Error != "" {
		return fmt.Errorf(checkError.Error)
	}

	return nil
}

func (adminApi *AdminApiKeycloak) DeleteClientScopes(scopeType string, clientID string, clientScopeID string) error {
	url := getUrlRealms(adminApi.Realms, adminApi.AdminApi)

	switch scopeType {
	case keycloakConstants.ClientScopesDefault:
		url += fmt.Sprintf("/clients/%s/default-client-scopes/%s",
			clientID, clientScopeID)
	case keycloakConstants.ClientScopesOptional:
		url += fmt.Sprintf("/clients/%s/optional-client-scopes/%s",
			clientID, clientScopeID)
	}

	resp, err := getClientResty().
		SetHeaders(map[string]string{
			"Authorization": "Bearer " + adminApi.AccessToken,
		}).
		Delete(url)
	if err != nil {
		return err
	}

	var checkError keycloakModels.ErrorResponse
	_ = json.Unmarshal(resp.Body(), &checkError)

	if checkError.ErrorDescription != "" {
		return fmt.Errorf(checkError.ErrorDescription)
	} else if checkError.ErrorMessage != "" {
		return fmt.Errorf(checkError.ErrorMessage)
	} else if checkError.Error != "" {
		return fmt.Errorf(checkError.Error)
	}

	return nil
}

func (adminApi *AdminApiKeycloak) ListClient(filter keycloakModels.ClientListFilter) ([]keycloakModels.Clients, error) {
	url := getUrlRealms(adminApi.Realms, adminApi.AdminApi)
	url += "/clients"

	var queryParams map[string]string
	filterMarshal, _ := json.Marshal(filter)
	_ = json.Unmarshal(filterMarshal, &queryParams)

	resp, err := getClientResty().
		SetHeaders(map[string]string{
			"Authorization": "Bearer " + adminApi.AccessToken,
		}).
		SetQueryParams(queryParams).
		Get(url)
	if err != nil {
		return nil, err
	}

	var checkError keycloakModels.ErrorResponse
	_ = json.Unmarshal(resp.Body(), &checkError)

	if checkError.ErrorDescription != "" {
		return nil, fmt.Errorf(checkError.ErrorDescription)
	} else if checkError.ErrorMessage != "" {
		return nil, fmt.Errorf(checkError.ErrorMessage)
	} else if checkError.Error != "" {
		return nil, fmt.Errorf(checkError.Error)
	}

	var clients []keycloakModels.Clients
	_ = json.Unmarshal(resp.Body(), &clients)

	return clients, nil
}

func (adminApi *AdminApiKeycloak) DetailClient(clientID string) error {
	url := getUrlRealms(adminApi.Realms, adminApi.AdminApi)
	url += fmt.Sprintf("/clients/%s", clientID)

	resp, err := getClientResty().
		SetHeaders(map[string]string{
			"Authorization": "Bearer " + adminApi.AccessToken,
		}).
		Get(url)
	if err != nil {
		return err
	}

	var checkError keycloakModels.ErrorResponse
	_ = json.Unmarshal(resp.Body(), &checkError)

	if checkError.ErrorDescription != "" {
		return fmt.Errorf(checkError.ErrorDescription)
	} else if checkError.ErrorMessage != "" {
		return fmt.Errorf(checkError.ErrorMessage)
	} else if checkError.Error != "" {
		return fmt.Errorf(checkError.Error)
	}

	return nil
}
