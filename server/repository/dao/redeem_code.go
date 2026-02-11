package dao

import (
	"context"
	"sync"

	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/log"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/repository"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/repository/model"
	"gorm.io/gorm"
)

type RedeemCodeDao interface {
	BatchInsert(ctx context.Context, redeemCodes []*model.RedeemCode) error
	GetByCode(ctx context.Context, code string) (*model.RedeemCode, error)
	QueryRedeemCodes(ctx context.Context, query *model.RedeemCodeQuery) ([]*model.RedeemCode, error)
	UseRedeemCodeInTransaction(ctx context.Context, redeemCode *model.RedeemCode, tx *gorm.DB) (int, error)
}

type RedeemCodeDaoImpl struct {
	db *gorm.DB
}

var (
	redeemCodeDaoImpl     RedeemCodeDao
	redeemCodeDaoSyncOnce sync.Once
)

func GetRedeemCodeDao() RedeemCodeDao {
	redeemCodeDaoSyncOnce.Do(func() {
		redeemCodeDaoImpl = &RedeemCodeDaoImpl{
			db: repository.DB,
		}
	})
	return redeemCodeDaoImpl
}

func (dao *RedeemCodeDaoImpl) BatchInsert(ctx context.Context, redeemCodes []*model.RedeemCode) error {
	ret := dao.db.WithContext(ctx).Create(&redeemCodes)
	if ret.Error != nil {
		log.Logger.Errorf("Failed to batch insert redeem codes: %v", ret.Error)
		return ret.Error
	}
	log.Logger.Infof("Successfully batch inserted %d redeem codes", ret.RowsAffected)
	return nil
}

func (dao *RedeemCodeDaoImpl) GetByCode(ctx context.Context, code string) (*model.RedeemCode, error) {
	var redeemCode model.RedeemCode
	ret := dao.db.WithContext(ctx).Where("code = ?", code).First(&redeemCode)
	if ret.Error != nil {
		log.Logger.Errorf("Failed to get redeem code by code %s: %v", code, ret.Error)
		return nil, ret.Error
	}
	return &redeemCode, nil
}

func (dao *RedeemCodeDaoImpl) QueryRedeemCodes(ctx context.Context, query *model.RedeemCodeQuery) ([]*model.RedeemCode, error) {
	var redeemCodes []*model.RedeemCode
	dbQquery := dao.db.WithContext(ctx).Model(&model.RedeemCode{})
	if query.Code != nil {
		dbQquery = dbQquery.Where("code = ?", *query.Code)
	}
	if query.Used != nil && *query.Used {
		dbQquery = dbQquery.Where("used_user_id != 0")
	} else if query.Used != nil && !*query.Used {
		dbQquery = dbQquery.Where("used_user_id = 0")
	}
	if query.Limit > 0 && query.Limit < repository.DefaultQueryLimit {
		dbQquery = dbQquery.Limit(query.Limit)
	} else {
		dbQquery = dbQquery.Limit(repository.DefaultQueryLimit)
	}
	ret := dbQquery.Find(&redeemCodes)
	if ret.Error != nil {
		log.Logger.Errorf("Failed to query redeem codes: %v", ret.Error)
		return nil, ret.Error
	}
	return redeemCodes, nil
}

func (dao *RedeemCodeDaoImpl) UseRedeemCodeInTransaction(ctx context.Context, redeemCode *model.RedeemCode, tx *gorm.DB) (int, error) {
	ret := tx.WithContext(ctx).Model(redeemCode).Where("used_user_id=0").Save(redeemCode)
	if ret.Error != nil {
		log.Logger.Errorf("Failed to update redeem code %s: %v", redeemCode.Code, ret.Error)
		return 0, ret.Error
	}
	log.Logger.Infof("Successfully updated redeem code %s", redeemCode.Code)
	return int(ret.RowsAffected), nil
}
