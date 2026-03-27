package otp_helper

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"go-api/modules/configs"

	"github.com/davecgh/go-spew/spew"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const OtpSenderSms = "sms"
const OtpSenderEmail = "email"
const OtpSenderWhatsapp = "whatsapp"

const otpKeyPrefix = "otp:%d:%s:%s" // otp:user_id:sender_type:action

var otpSender = map[string]bool{
	OtpSenderSms:      true,
	OtpSenderEmail:    true,
	OtpSenderWhatsapp: true,
}

func isOtpSender(senderType string) bool {
	senderType = strings.ToLower(senderType)
	_, ok := otpSender[senderType]
	return ok
}

type OtpSender struct {
	ctx         context.Context
	redisClient redis.UniversalClient
	db          *gorm.DB
	expireTime  time.Duration

	userID uint64
}

func NewOtpSender(ctx context.Context) *OtpSender {
	return &OtpSender{
		ctx:         ctx,
		redisClient: configs.GetRedis("OTP"),
		db:          configs.DatabaseBase(configs.MySQLType).GetMysqlConnection(true),
		expireTime:  5 * time.Minute, // default 5 minutes
	}
}

func (o *OtpSender) SetExpireTime(expireTime time.Duration) *OtpSender {
	o.expireTime = expireTime
	return o
}

func (o *OtpSender) SetUserID(userID uint64) *OtpSender {
	o.userID = userID
	return o
}

func (o *OtpSender) SendOtp(senderType string, action string) error {
	if !isOtpSender(senderType) {
		return fmt.Errorf("invalid sender type: %s", senderType)
	}

	otpKey := spew.Sprintf(otpKeyPrefix, o.userID, senderType, action)

	// generate otp
	otp, err := generateOtp(6)
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	// save otp to redis
	err = o.redisClient.Set(o.ctx, otpKey, otp, o.expireTime).Err()
	if err != nil {
		return fmt.Errorf("failed to save otp to redis: %w", err)
	}

	switch senderType {
	case OtpSenderSms:
		break
	case OtpSenderEmail:
		emailHelper := sendEmailOtp(o.db).setSendTo(o.userID)
		err = emailHelper.send(action, otp)
	case OtpSenderWhatsapp:
		break
	}

	return nil
}

func (o *OtpSender) VerifyOtp(senderType, action, otpInput string) error {
	if !isOtpSender(senderType) {
		return fmt.Errorf("invalid sender type: %s", senderType)
	}

	// get otp from redis
	otpKey := spew.Sprintf(otpKeyPrefix, o.userID, senderType, action)
	otp, err := o.redisClient.Get(o.ctx, otpKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get otp from redis: %w", err)
	}

	if otp == "" {
		return fmt.Errorf("otp not found in redis")
	}

	// validate otp
	if otpInput != otp {
		return fmt.Errorf("invalid OTP")
	}

	return nil
}

func generateOtp(length int) (string, error) {
	if length < 6 || length > 8 {
		length = 6
	}

	otp := ""
	for i := 0; i < length; i++ {
		digit, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", fmt.Errorf("failed to generate OTP: %w", err)
		}
		otp += digit.String()
	}

	return otp, nil
}
