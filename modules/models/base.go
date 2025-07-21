package models

import (
	"encoding/json"
	"time"

	"github.com/fari-99/go-helper/rabbitmq"
	"github.com/google/uuid"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

type IDType string // for uuid

type Base struct {
	ID        IDType     `gorm:"column:id" json:"id" sql:"type:uuid;primary_key;default:uuid_generate_v4()" `
	CreatedAt time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"deleted_at" sql:"DEFAULT:NULL"`
}

func (base *Base) BeforeCreate(tx *gorm.DB) error {
	idUuid := uuid.New()
	base.ID = IDType(idUuid.String())
	return nil
}

func (base *Base) AfterCreate(tx *gorm.DB) error {
	publishEventDatabase(tx, "create")
	return nil
}

func (base *Base) AfterUpdate(tx *gorm.DB) error {
	publishEventDatabase(tx, "update")
	return nil
}

func (base *Base) AfterDelete(tx *gorm.DB) error {
	publishEventDatabase(tx, "delete")
	return nil
}

func publishEventDatabase(tx *gorm.DB, action string) {
	currentTable := tx.Statement.Table
	ctx := tx.Statement.Context
	data := tx.Statement.Model

	// default action
	actionBy := map[string]interface{}{
		"id":       0,
		"username": "System",
		"email":    "-",
	}

	if ctx != nil {
		currentUserMarshal := cast.ToString(ctx.Value("user_details"))
		var currentUser Users
		err := json.Unmarshal([]byte(currentUserMarshal), &currentUser)
		if err == nil {
			actionBy = map[string]interface{}{
				"id":       currentUser.ID,
				"username": currentUser.Username,
				"email":    currentUser.Email,
			}
		}
	}

	modelMarshal, _ := json.Marshal(data)

	eventData := map[string]interface{}{
		"event_type": action,
		"table":      currentTable,
		"action_by":  actionBy,
		"timestamp":  time.Now(),
		"data":       string(modelMarshal),
	}

	eventDataMarshal, _ := json.Marshal(eventData)

	queueSetup := rabbitmq.NewBaseQueue("", "database_action_logs")
	queueSetup.SetupQueue(nil, nil)
	queueSetup.AddPublisher(nil, nil)
	_ = queueSetup.Publish(string(eventDataMarshal))
}
