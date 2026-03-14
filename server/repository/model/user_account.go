package model

import (
	"fmt"
	"time"

	ceramicraftsecure "github.com/sw5005-sus/ceramicraft-secure"
)

type UserAccount struct {
	ID            int       `gorm:"type:int;primaryKey"`
	UserId        int       `gorm:"type:int;uniqueIndex;not null"`
	AccountNo     string    `gorm:"type:varchar(32);uniqueIndex;not null"`
	Balance       int       `gorm:"type:int;not null;default:0"`
	IntegritySign string    `gorm:"type:varchar(128);not null"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

func (u *UserAccount) TableName() string {
	return "user_accounts"
}

func (u *UserAccount) GetHiddenAccountNo() string {
	if len(u.AccountNo) <= 8 {
		return u.AccountNo
	}
	return u.AccountNo[0:4] + "****" + u.AccountNo[len(u.AccountNo)-4:]
}

var (
	GenHmacF1unc     = ceramicraftsecure.GenHmacSha256
	VerifyHmacSha256 = ceramicraftsecure.VerifyHmacSha256
)

func (u *UserAccount) Sign(latestLog *UserAccountChangeLog) error {
	key := u.GetToSignData(latestLog)
	sign, err := GenHmacF1unc(key)
	if err != nil {
		return err
	}
	u.IntegritySign = sign
	return nil
}

func (u *UserAccount) GetToSignData(latestLog *UserAccountChangeLog) string {
	return fmt.Sprintf("%d:%d:%d", u.UserId, u.Balance, latestLog.ID)
}
