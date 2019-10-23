package token_generator

type TokenData struct {
	Origin      string      `json:"origin,omitempty"`
	Authorized  bool        `json:"authorized,omitempty"`
	UserDetails UserDetails `json:"user_details,omitempty"`
}

type UserDetails struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type TokenGenerator struct {
	Type string `json:"type"`
}

func NewBaseTokenGenerator() {

}
