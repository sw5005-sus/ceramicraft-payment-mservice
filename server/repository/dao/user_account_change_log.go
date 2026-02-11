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

type UserAccountChangeLogDAO interface {
	CreateChangeLogInTransaction(ctx context.Context, changeLog *model.UserAccountChangeLog, tx *gorm.DB) error
	QueryChangeLogs(ctx context.Context, query *model.UserAccountChangeLogQuery) ([]*model.UserAccountChangeLog, error)
}

var (
	userAccountChangeLogDAOImpl     UserAccountChangeLogDAO
	userAccountChangeLogDAOSyncOnce sync.Once
)

func GetUserAccountChangeLogDAO() UserAccountChangeLogDAO {
	userAccountChangeLogDAOSyncOnce.Do(func() {
		userAccountChangeLogDAOImpl = &UserAccountChangeLogDAOImpl{
			db: repository.DB,
		}
	})
	return userAccountChangeLogDAOImpl
}

type UserAccountChangeLogDAOImpl struct {
	db *gorm.DB
}

// CreateChangeLogInTransaction implements UserAccountChangeLogDAO.
func (u *UserAccountChangeLogDAOImpl) CreateChangeLogInTransaction(ctx context.Context, changeLog *model.UserAccountChangeLog, tx *gorm.DB) error {
	ret := tx.WithContext(ctx).Create(changeLog)
	if ret.Error != nil {
		if errors.Is(ret.Error, gorm.ErrDuplicatedKey) {
			log.Logger.Warnf("User account change log already exists for user ID %d, biz ID %s", changeLog.AccountId, changeLog.IdempotentKey)
		} else {
			log.Logger.Errorf("Failed to create user account change log for user ID %d, biz ID %s: %v", changeLog.AccountId, changeLog.IdempotentKey, ret.Error)
		}
		return ret.Error
	}
	return nil
}

func (u *UserAccountChangeLogDAOImpl) QueryChangeLogs(ctx context.Context, query *model.UserAccountChangeLogQuery) ([]*model.UserAccountChangeLog, error) {
	var changeLogs []*model.UserAccountChangeLog
	dbQuery := u.db.WithContext(ctx).Model(&model.UserAccountChangeLog{})
	if query.AccountId != nil {
		dbQuery = dbQuery.Where("account_id = ?", *query.AccountId)
	}
	if query.IdempotentKey != nil {
		dbQuery = dbQuery.Where("idempotent_key = ?", *query.IdempotentKey)
	}
	ret := dbQuery.Order("id desc").Limit(repository.DefaultQueryLimit).Find(&changeLogs)
	if ret.Error != nil {
		log.Logger.Errorf("Failed to query user account change logs: %v", ret.Error)
		return nil, ret.Error
	}
	return changeLogs, nil
}
