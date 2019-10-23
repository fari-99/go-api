package token_generator

type AccessToken struct {
	Origin      string      `json:"origin,omitempty"`
	Authorized  bool        `json:"authorized,omitempty"`
	UserDetails UserDetails `json:"user_details,omitempty"`
}

