package models

import "time"

type UserCodes struct {
	Base
	UserID    string    `json:"user_id"`
	Via       string    `json:"via"`
	Code      string    `json:"code"`
	Params    string    `json:"params"`
	IsUsed    int8      `json:"is_used"`
	ExpiredAt time.Time `json:"expired_at"`
}

func (model UserCodes) TableName() string {
	return "user_codes"
}
