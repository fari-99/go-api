package models

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
)

type Transactions struct {
	ID            int64  `gorm:"column:id" json:"id"`
	TransactionNo string `gorm:"column:transaction_no" json:"transaction_no"`
	Status        int8   `gorm:"column:status" json:"status"`

	CreatedAt time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"deleted_at" sql:"DEFAULT:NULL"`
}

func (model *Transactions) generateTransactionNo(tx *gorm.DB) string {
	modelID := strconv.FormatInt(model.ID, 10)
	referenceNo := strconv.FormatInt(rand.Int63n(100000), 10)
	dateFormat := time.Now().Format("060102")

	transactionNo := fmt.Sprintf("ORDER:%s-%s-%s", modelID, referenceNo, dateFormat)
	return transactionNo
}

func (model *Transactions) AfterCreate(tx *gorm.DB) {
	if model.TransactionNo == "" || len(model.TransactionNo) == 0 {
		tx.Model(model).Update(&Transactions{TransactionNo: model.generateTransactionNo(tx)})
	}
}
