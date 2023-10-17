package models

import "time"

type UserSocials struct {
	Base
	UserID            IDType     `json:"user_id"`
	NotificationType  int8       `json:"notification_type"`
	Token             string     `json:"token"`
	Identifier        string     `json:"identifier"`
	ExpiredIdentifier *time.Time `json:"expired_identifier"`
	Status            int8       `json:"status"`

	LinkAuth string `json:"link_auth" db:"-"`
}

func (UserSocials) UserSocials() string {
	return "user_socials"
}
