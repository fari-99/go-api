package token_generator

type TokenData struct {
	Origin      string       `json:"origin,omitempty"`
	Authorized  bool         `json:"authorized,omitempty"`
	UserDetails *UserDetails `json:"user_details,omitempty"`
	AppData     *AppData     `json:"app_data,omitempty"`
}

type UserDetails struct {
	ID       int64  `json:"id,omitempty"`
	Email    string `json:"email,omitempty"`
	Username string `json:"username,omitempty"`
}

type AppData struct {
	AppName string   `json:"app_name"`
	IPList  []string `json:"ip_list,omitempty"`
}

type TokenGenerator struct {
	Type string `json:"type"`
}

func NewBaseTokenGenerator() {

}
