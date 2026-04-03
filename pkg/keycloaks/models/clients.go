package keycloaks

type Clients struct {
	ErrorResponse

	Id                                 string                `json:"id"`
	ClientId                           string                `json:"clientId"`
	Name                               string                `json:"name"`
	Description                        string                `json:"description"`
	RootUrl                            string                `json:"rootUrl"`
	AdminUrl                           string                `json:"adminUrl"`
	BaseUrl                            string                `json:"baseUrl"`
	SurrogateAuthRequired              bool                  `json:"surrogateAuthRequired"`
	Enabled                            bool                  `json:"enabled"`
	AlwaysDisplayInConsole             bool                  `json:"alwaysDisplayInConsole"`
	ClientAuthenticatorType            string                `json:"clientAuthenticatorType"`
	Secret                             string                `json:"secret"`
	RedirectUris                       []string              `json:"redirectUris"`
	WebOrigins                         []string              `json:"webOrigins"`
	NotBefore                          int                   `json:"notBefore"`
	BearerOnly                         bool                  `json:"bearerOnly"`
	ConsentRequired                    bool                  `json:"consentRequired"`
	StandardFlowEnabled                bool                  `json:"standardFlowEnabled"`
	ImplicitFlowEnabled                bool                  `json:"implicitFlowEnabled"`
	DirectAccessGrantsEnabled          bool                  `json:"directAccessGrantsEnabled"`
	ServiceAccountsEnabled             bool                  `json:"serviceAccountsEnabled"`
	PublicClient                       bool                  `json:"publicClient"`
	FrontchannelLogout                 bool                  `json:"frontchannelLogout"`
	Protocol                           string                `json:"protocol"`
	Attributes                         ClientAttributes      `json:"attributes"`
	AuthenticationFlowBindingOverrides AuthFlowBindOverrides `json:"authenticationFlowBindingOverrides"`
	FullScopeAllowed                   bool                  `json:"fullScopeAllowed"`
	NodeReRegistrationTimeout          int                   `json:"nodeReRegistrationTimeout"`
	DefaultClientScopes                []string              `json:"defaultClientScopes"`
	OptionalClientScopes               []string              `json:"optionalClientScopes"`
	Access                             ClientAccess          `json:"access"`
}

type AuthFlowBindOverrides struct {
}

type ClientAttributes struct {
	OidcCibaGrantEnabled                  string `json:"oidc.ciba.grant.enabled"`
	ClientSecretCreationTime              string `json:"client.secret.creation.time"`
	BackchannelLogoutSessionRequired      string `json:"backchannel.logout.session.required"`
	PostLogoutRedirectUris                string `json:"post.logout.redirect.uris"`
	Oauth2DeviceAuthorizationGrantEnabled string `json:"oauth2.device.authorization.grant.enabled"`
	BackchannelLogoutRevokeOfflineTokens  string `json:"backchannel.logout.revoke.offline.tokens"`
}

type ClientAccess struct {
	View      bool `json:"view"`
	Configure bool `json:"configure"`
	Manage    bool `json:"manage"`
}
