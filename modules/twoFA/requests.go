package twoFA

type RequestValidateRecoveryCode struct {
	RecoveryCode string `json:"recovery_code"`
}

type Request2FADisabled struct {
	Password string `json:"password"`
}

type RequestDisableTotp struct {
	Password string `json:"password" binding:"required"`
	OtpValue string `json:"otp_value" binding:"required"`
}

type RequestValidateOtp struct {
	OtpValue string `json:"otp_value" binding:"required"`
}

type RequestDisableOtp struct {
	Password string `json:"password" binding:"required"`
}
