package middleware

type BaseMiddleware struct {
	AllowedAppName []string `json:"allowed_app_name"`
	Whitelist      []string `json:"whitelist"`
	Blacklist      []string `json:"blacklist"`
}

func DefaultConfig() BaseMiddleware {
	config := BaseMiddleware{
		AllowedAppName: make([]string, 0, 2),
		Whitelist:      make([]string, 0, 2),
		Blacklist:      make([]string, 0, 2),
	}

	return config
}
