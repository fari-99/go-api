package keycloaks

import (
	"encoding/json"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// InputCreateClientScopes
/* example input
{
  "name": "sadadadas",
  "description": "dasdasdada",
  "attributes": {
    "consent.screen.text": "",
    "display.on.consent.screen": "false",
    "include.in.token.scope": "false",
    "gui.order": "100000000000"
  },
  "type": "optional",
  "protocol": "openid-connect"
}
*/
type InputCreateClientScopes struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Attributes  ClientScopeAttributes `json:"attributes"`
	Type        string                `json:"type"`
	Protocol    string                `json:"protocol"`
}

func (model InputCreateClientScopes) Validate() error {
	return validation.ValidateStruct(&model,
		validation.Field(&model.Name, validation.Required),
		validation.Field(&model.Description, validation.Required),
		validation.Field(&model.Type, validation.Required),
		validation.Field(&model.Protocol, validation.Required),
		validation.Field(&model.Attributes, validation.By(func(value interface{}) error {
			var clientAttributes ClientScopeAttributes
			valueMarshal, _ := json.Marshal(value)
			_ = json.Unmarshal(valueMarshal, &clientAttributes)

			return clientAttributes.Validate()
		})),
	)
}

type ClientScopeAttributes struct {
	ConsentScreenText      string `json:"consent.screen.text"`
	DisplayOnConsentScreen string `json:"display.on.consent.screen"`
	IncludeInTokenScope    string `json:"include.in.token.scope"`
	GuiOrder               string `json:"gui.order"`
}

func (model ClientScopeAttributes) Validate() error {
	return validation.ValidateStruct(&model,
		// validation.Field(&model.ConsentScreenText, validation.Required),
		// validation.Field(&model.DisplayOnConsentScreen, validation.Required),
		// validation.Field(&model.IncludeInTokenScope, validation.Required),
		validation.Field(&model.GuiOrder, validation.Required),
	)
}

type ClientListFilter struct {
	ClientID     string `json:"clientId,omitempty"`
	First        string `json:"first,omitempty"`
	Max          string `json:"max,omitempty"`
	Q            string `json:"q,omitempty"`
	Search       string `json:"search,omitempty"`
	ViewAbleOnly string `json:"viewableOnly,omitempty"`
}
