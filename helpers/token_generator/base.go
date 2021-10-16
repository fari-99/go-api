package token_generator

import (
	"encoding/json"
	"go-api/helpers/crypts"
	"go-api/modules/models"
	"os"
)

type TokenData struct {
	Origin      string   `json:"origin,omitempty"`
	Authorized  bool     `json:"authorized,omitempty"`
	UserDetails string   `json:"user_details,omitempty"`
	AppData     *AppData `json:"app_data,omitempty"`
}

type UserDetails struct {
	ID        int64    `json:"id,omitempty"`
	Email     string   `json:"email,omitempty"`
	Username  string   `json:"username,omitempty"`
	UserRoles []string `json:"user_roles"`
}

type AppData struct {
	AppName   string   `json:"app_name"`
	IPList    []string `json:"ip_list,omitempty"`
	UserAgent string   `json:"user_agent"`
}

type TokenGenerator struct {
	Type string `json:"type"`
}

func EncryptUserDetails(customer models.Customers) (string, error) {
	userDetails := UserDetails{
		ID:       customer.ID,
		Email:    customer.Email,
		Username: customer.Username,
	}

	dataMarshal, _ := json.Marshal(userDetails)

	encryptionHelper := crypts.NewEncryptionBase().SetPassphrase(os.Getenv("USER_DETAILS_PASSPHRASE"))
	encryptedData, err := encryptionHelper.Encrypt(dataMarshal)

	return string(encryptedData), err
}

func DecryptUserDetails(secretMessage string) (UserDetails, error) {
	encryptionHelper := crypts.NewEncryptionBase().SetPassphrase(os.Getenv("USER_DETAILS_PASSPHRASE"))
	decryptedData, err := encryptionHelper.Decrypt([]byte(secretMessage))
	if err != nil {
		return UserDetails{}, err
	}

	var userDetails UserDetails
	err = json.Unmarshal(decryptedData, &userDetails)
	if err != nil {
		return UserDetails{}, err
	}

	return userDetails, nil
}
