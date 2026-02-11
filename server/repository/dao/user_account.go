package dao

import (
	"context"
	"errors"
	"sync"

	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/log"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/repository"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/repository/model"
	"gorm.io/gorm"
)

type UserAccountDao interface {
	CreateUserAccount(ctx context.Context, userAccount *model.UserAccount) error
	GetUserAccountByUserID(ctx context.Context, userID int) (*model.UserAccount, error)
	AddBalanceInTransaction(ctx context.Context, userID int, amount int, oldAmount int, tx *gorm.DB) error
	SubtractBalanceInTransaction(ctx context.Context, userID int, amount int, oldAmount int, tx *gorm.DB) (int, error)
}

var (
	userAccountDaoImpl     UserAccountDao
	userAccountDaoSyncOnce sync.Once
)

func GetUserAccountDao() UserAccountDao {
	userAccountDaoSyncOnce.Do(func() {
		userAccountDaoImpl = &UserAccountDaoImpl{
			db: repository.DB,
		}
	})
	return userAccountDaoImpl
}

type UserAccountDaoImpl struct {
	db *gorm.DB
}

// CreateUserAccount implements UserAccountDao.
func (u *UserAccountDaoImpl) CreateUserAccount(ctx context.Context, userAccount *model.UserAccount) error {
	ret := u.db.WithContext(ctx).Create(userAccount)
	if ret.Error != nil {
		if errors.Is(ret.Error, gorm.ErrDuplicatedKey) {
			log.Logger.Warnf("User account already exists for user ID %d", userAccount.UserId)
			return nil
		}
		log.Logger.Errorf("Failed to create user account for user ID %d: %v", userAccount.UserId, ret.Error)
		return ret.Error
	}
	return ret.Error
}

// GetUserAccountByUserID implements UserAccountDao.
func (u *UserAccountDaoImpl) GetUserAccountByUserID(ctx context.Context, userID int) (*model.UserAccount, error) {
	var userAccount model.UserAccount
	ret := u.db.WithContext(ctx).Where("user_id = ?", userID).First(&userAccount)
	if ret.Error != nil {
		if errors.Is(ret.Error, gorm.ErrRecordNotFound) {
			log.Logger.Warnf("User account not found for user ID %d", userID)
			return nil, nil
		}
		log.Logger.Errorf("Failed to get user account by user ID %d: %v", userID, ret.Error)
		return nil, ret.Error
	}
	return &userAccount, nil
}

// AddBalance implements UserAccountDao.
func (u *UserAccountDaoImpl) AddBalanceInTransaction(ctx context.Context, userID int, amount int, oldAmount int, tx *gorm.DB) error {
	ret := tx.WithContext(ctx).Model(&model.UserAccount{}).
		Where("user_id = ? and balance=?", userID, oldAmount).
		Update("balance", gorm.Expr("balance + ?", amount))
	if ret.Error != nil {
		log.Logger.Errorf("Failed to add balance for user ID %d: %v", userID, ret.Error)
		return ret.Error
	}
	if ret.RowsAffected == 0 {
		log.Logger.Warnf("No user account found to add balance for user ID %d", userID)
		return gorm.ErrCheckConstraintViolated
	}
	log.Logger.Infof("Successfully added %d to user ID %d", amount, userID)
	return nil
}

// SubtractBalance implements UserAccountDao.
func (u *UserAccountDaoImpl) SubtractBalanceInTransaction(ctx context.Context, userID int, amount int, oldAmount int, tx *gorm.DB) (int, error) {
	ret := tx.WithContext(ctx).Model(&model.UserAccount{}).
		Where("user_id = ? and balance=?", userID, oldAmount).
		Update("balance", gorm.Expr("balance - ?", amount))
	if ret.Error != nil {
		log.Logger.Errorf("Failed to subtract balance for user ID %d: %v", userID, ret.Error)
		return 0, ret.Error
	}
	if ret.RowsAffected == 0 {
		log.Logger.Warnf("No user account found to subtract balance for user ID %d", userID)
		return 0, gorm.ErrCheckConstraintViolated
	}
	log.Logger.Infof("Successfully subtracted %d from user ID %d", amount, userID)
	return int(ret.RowsAffected), nil
}
