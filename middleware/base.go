package middleware

type BaseMiddleware struct {
	AllowedRoles   []string `json:"allowed_roles"`
	AllowedAppName []string `json:"allowed_app_name"`
	Whitelist      []string `json:"whitelist"`
	Blacklist      []string `json:"blacklist"`
}

func DefaultConfig() BaseMiddleware {
	config := BaseMiddleware{
		AllowedRoles:   make([]string, 0, 2),
		AllowedAppName: make([]string, 0, 2),
		Whitelist:      make([]string, 0, 2),
		Blacklist:      make([]string, 0, 2),
	}

	return config
}
