package constant

// RedisSession Session

const QRCodeWhatsapp = "QR-CODE-WHATSAPP"

// RedisLock

const RedLockValidateTotp = "VALIDATE_TOTP:%d"
const RedLockValidateRecoveryCode = "VALIDATE_RECOVERY_CODE:%d"
const RedLockCreateOtp = "CREATE_OTP:%d:%s:%s"
const RedLockValidateOtp = "VALIDATE_OTP:%d:%s:%s"
