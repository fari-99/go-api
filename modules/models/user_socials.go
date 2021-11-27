package models

import "time"

type UserSocials struct {
	ID                uint64     `json:"id"`
	UserID            uint64     `json:"user_id"`
	NotificationType  int8       `json:"notification_type"`
	Token             string     `json:"token"`
	Identifier        string     `json:"identifier"`
	ExpiredIdentifier *time.Time `json:"expired_identifier"`
	Status            int8       `json:"status"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	DeletedAt         *time.Time `json:"deleted_at"`

	LinkAuth string `json:"link_auth" db:"-"`
}

func (UserSocials) UserSocials() string {
	return "user_socials"
}
