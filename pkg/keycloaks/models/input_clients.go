package keycloaks

import (
	"encoding/json"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// InputCreateClients
/* example input
{
  "protocol": "openid-connect",
  "clientId": "test",
  "name": "adsad",
  "description": "dasdsadsas",
  "publicClient": false,
  "authorizationServicesEnabled": false,
  "serviceAccountsEnabled": false,
  "implicitFlowEnabled": false,
  "directAccessGrantsEnabled": false,
  "standardFlowEnabled": true,
  "frontchannelLogout": true,
  "attributes": {
    "saml_idp_initiated_sso_url_name": "",
    "oauth2.device.authorization.grant.enabled": false,
    "oidc.ciba.grant.enabled": false,
    "post.logout.redirect.uris": "/*"
  },
  "alwaysDisplayInConsole": false,
  "rootUrl": "https://test.com",
  "baseUrl": "https://test.com",
  "redirectUris": [
    "/*"
  ],
  "webOrigins": [
    "+"
  ]
}
*/
type InputCreateClients struct {
	Protocol                     string                `json:"protocol"`
	ClientId                     string                `json:"clientId"`
	Name                         string                `json:"name"`
	Description                  string                `json:"description"`
	PublicClient                 bool                  `json:"publicClient"`
	AuthorizationServicesEnabled bool                  `json:"authorizationServicesEnabled"`
	ServiceAccountsEnabled       bool                  `json:"serviceAccountsEnabled"`
	ImplicitFlowEnabled          bool                  `json:"implicitFlowEnabled"`
	DirectAccessGrantsEnabled    bool                  `json:"directAccessGrantsEnabled"`
	StandardFlowEnabled          bool                  `json:"standardFlowEnabled"`
	FrontchannelLogout           bool                  `json:"frontchannelLogout"`
	AlwaysDisplayInConsole       bool                  `json:"alwaysDisplayInConsole"`
	RootUrl                      string                `json:"rootUrl"`
	BaseUrl                      string                `json:"baseUrl"`
	RedirectUris                 []string              `json:"redirectUris"`
	WebOrigins                   []string              `json:"webOrigins"`
	Attributes                   InputClientAttributes `json:"attributes"`
}

func (model InputCreateClients) Validate() error {
	return validation.ValidateStruct(&model,
		validation.Field(&model.Protocol, validation.Required),
		validation.Field(&model.ClientId, validation.Required),
		validation.Field(&model.Name, validation.Required),
		validation.Field(&model.Description, validation.Required),
		validation.Field(&model.StandardFlowEnabled, validation.Required),
		validation.Field(&model.FrontchannelLogout, validation.Required),
		validation.Field(&model.RootUrl, validation.Required),
		validation.Field(&model.BaseUrl, validation.Required),
		validation.Field(&model.RedirectUris, validation.Required),
		validation.Field(&model.WebOrigins, validation.Required),
		validation.Field(&model.Attributes, validation.By(func(value interface{}) error {
			var clientAttributes InputClientAttributes
			valueMarshal, _ := json.Marshal(value)
			_ = json.Unmarshal(valueMarshal, &clientAttributes)

			return clientAttributes.Validate()
		})),
	)
}

type InputClientAttributes struct {
	SamlIdpInitiatedSsoUrlName            string `json:"saml_idp_initiated_sso_url_name"`
	Oauth2DeviceAuthorizationGrantEnabled bool   `json:"oauth2.device.authorization.grant.enabled"`
	OidcCibaGrantEnabled                  bool   `json:"oidc.ciba.grant.enabled"`
	PostLogoutRedirectUris                string `json:"post.logout.redirect.uris"`
}

func (model InputClientAttributes) Validate() error {
	return validation.ValidateStruct(&model,
		validation.Field(&model.PostLogoutRedirectUris, validation.Required),
	)
}
