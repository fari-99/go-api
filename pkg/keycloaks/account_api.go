package keycloaks

import (
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"

	keycloakModel "go-api/pkg/keycloaks/models"
)

type AccountApiKeycloak struct {
	Realms      string        `json:"realms"`
	AdminApi    bool          `json:"admin_api"`
	RedisClient *redis.Client `json:"redis_client"`

	AccessToken string `json:"access_token"`
}

func NewAccountApi(realms string) (*AccountApiKeycloak, error) {
	if exists, err := RealmExists(realms); err != nil {
		return nil, err
	} else if !exists {
		return nil, fmt.Errorf("realms not exists")
	}

	accountApiKeycloak := &AccountApiKeycloak{
		Realms:   realms,
		AdminApi: false,
	}

	return accountApiKeycloak, nil
}

func (accountApi *AccountApiKeycloak) SetUserAccessToken(accessToken string) *AccountApiKeycloak {
	accountApi.AccessToken = accessToken
	return accountApi
}

func (accountApi *AccountApiKeycloak) GetAccount() (*keycloakModel.AccountUserProfile, error) {
	url := getUrlRealms(accountApi.Realms, accountApi.AdminApi)
	url += "/account"

	resp, err := getClientResty().
		SetHeaders(map[string]string{
			"Authorization": accountApi.AccessToken,
			"accept":        "application/json",
		}).
		Get(url)
	if err != nil {
		return nil, err
	}

	var accountData keycloakModel.AccountUserProfile
	_ = json.Unmarshal(resp.Body(), &accountData)

	return &accountData, nil
}

func (accountApi *AccountApiKeycloak) SetAttributes(input keycloakModel.AccountUserProfile) (*keycloakModel.AccountUserProfile, error) {
	url := getUrlRealms(accountApi.Realms, accountApi.AdminApi)
	url += "/account"

	resp, err := getClientResty().
		SetHeaders(map[string]string{
			"Authorization": accountApi.AccessToken,
			"accept":        "application/json",
		}).
		SetBody(input).
		Post(url)
	if err != nil {
		return nil, err
	}

	var accountData keycloakModel.AccountUserProfile
	_ = json.Unmarshal(resp.Body(), &accountData)

	return &accountData, nil
}
