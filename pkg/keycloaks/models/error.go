package keycloaks

type ErrorResponse struct {
	Error            string `json:"error,omitempty"`
	ErrorMessage     string `json:"errorMessage,omitempty"`
	ErrorDescription string `json:"errorDescription,omitempty"`
}
