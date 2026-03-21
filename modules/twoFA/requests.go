package twoFA

type RequestValidateRecoveryCode struct {
	RecoveryCode string `json:"recovery_code"`
}

type Request2FADisabled struct {
	Password string `json:"password"`
}
