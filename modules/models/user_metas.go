package models

type UserMetas struct {
	Base
	UserID IDType `json:"user_id"`
	Groups string `json:"groups"`
	Key    string `json:"key"`
	Value  string `json:"value"`
}

func (UserMetas) TableName() string {
	return "user_metas"
}
