package models

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type Transactions struct {
	Base
	TransactionNo string `gorm:"column:transaction_no" json:"transaction_no"`
	Status        uint8  `gorm:"column:status" json:"status"`
	CreatedBy     IDType `gorm:"column:created_by" json:"created_by"`
}

func (Transactions) TableName() string {
	return "transactions"
}

func (model *Transactions) generateTransactionNo(tx *gorm.DB) string {
	modelID := model.ID.String()
	referenceNo := strconv.FormatInt(rand.Int63n(100000), 10)
	dateFormat := time.Now().Format("060102")

	transactionNo := fmt.Sprintf("ORDER:%s-%s-%s", dateFormat, modelID, referenceNo)
	return transactionNo
}

func (model *Transactions) AfterCreate(tx *gorm.DB) {
	if model.TransactionNo == "" || len(model.TransactionNo) == 0 {
		tx.Model(model).Updates(&Transactions{TransactionNo: model.generateTransactionNo(tx)})
	}
}
