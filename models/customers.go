package models

import "time"

type Customers struct {
	ID       int64  `gorm:"id" json:"id"`
	Username string `gorm:"username" json:"username"`
	Password string `gorm:"password" json:"password"`
	Email    string `gorm:"email" json:"email"`
	Status   int8   `gorm:"status" json:"status"`

	CreatedAt time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"deleted_at" sql:"DEFAULT:NULL"`
}

type CustomerResult struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Status   int8   `json:"status"`

	BearerToken string `json:"bearer_token"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}
