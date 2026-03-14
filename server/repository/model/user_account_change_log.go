package model

import "time"

const (
	OpTypeTopUp   = 1
	OpTypePayment = 2
)

type UserAccountChangeLog struct {
	ID            int       `gorm:"primaryKey"`
	AccountId     int       `gorm:"index;not null"`
	OpType        int       `gorm:"not null"` // 1: top-up, 2: payment
	Amount        int       `gorm:"not null"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	IdempotentKey string    `gorm:"type:varchar"` // JSON string for extra info
}

func (u *UserAccountChangeLog) TableName() string {
	return "user_account_change_logs"
}

type UserAccountChangeLogQuery struct {
	AccountId     *int
	OpType        int
	IdempotentKey *string
	Limit         int
}
