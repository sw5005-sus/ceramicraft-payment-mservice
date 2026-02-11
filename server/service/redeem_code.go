package service

import (
	"context"
	"sync"
	"time"

	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/http/data"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/log"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/repository/dao"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/repository/model"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/utils"
)

type RedeemCodeService interface {
	GenerateRedeemCodes(ctx context.Context, amount int, quantity int) error
	QueryRedeemCodes(ctx context.Context, query *data.RedeemCodeQuery) ([]*data.RedeemCodeVO, error)
}

var (
	redeemCodeServiceInstance RedeemCodeService
	redeemCodeServiceOnce     sync.Once
)

func GetRedeemCodeService() RedeemCodeService {
	redeemCodeServiceOnce.Do(func() {
		redeemCodeServiceInstance = &RedeemCodeServiceImpl{
			redeemCodeDao: dao.GetRedeemCodeDao(),
		}
	})
	return redeemCodeServiceInstance
}

type RedeemCodeServiceImpl struct {
	redeemCodeDao dao.RedeemCodeDao
}

const redeemCodeSize = 16

// GenerateRedeemCodes implements RedeemCodeService.
func (r *RedeemCodeServiceImpl) GenerateRedeemCodes(ctx context.Context, amount int, quantity int) error {
	toInsert := make([]*model.RedeemCode, quantity)
	currentTime := time.Now()
	codeSet := make(map[string]struct{})
	for i := 0; i < quantity; i++ {
		var code string
		for {
			code = utils.GenRedeemCode(redeemCodeSize)
			if _, exists := codeSet[code]; !exists {
				log.Logger.Infof("Generated redeem code: %s", code)
				codeSet[code] = struct{}{}
				break
			}
			log.Logger.Warnf("Duplicate redeem code generated: %s, regenerating...", code)
		}
		toInsert[i] = &model.RedeemCode{
			Code:      code,
			Amount:    amount,
			CreatedAt: currentTime,
			UpdatedAt: currentTime,
		}
	}
	err := r.redeemCodeDao.BatchInsert(ctx, toInsert)
	if err != nil {
		log.Logger.Errorf("Failed to generate redeem codes: %v", err)
		return err
	}
	log.Logger.Infof("Successfully generated %d redeem codes with amount %d each", quantity, amount)
	return nil
}

// QueryRedeemCodes implements RedeemCodeService.
func (r *RedeemCodeServiceImpl) QueryRedeemCodes(ctx context.Context, query *data.RedeemCodeQuery) ([]*data.RedeemCodeVO, error) {
	dbQuery := &model.RedeemCodeQuery{
		Limit: query.Limit,
		Code:  query.Code,
		Used:  query.Used,
	}
	redeemCodes, err := r.redeemCodeDao.QueryRedeemCodes(ctx, dbQuery)
	if err != nil {
		log.Logger.Errorf("Failed to query redeem codes: %v", err)
		return nil, err
	}
	result := make([]*data.RedeemCodeVO, len(redeemCodes))
	for i, rc := range redeemCodes {
		result[i] = &data.RedeemCodeVO{
			Id:         int(rc.ID),
			Code:       rc.Code,
			Amount:     int(rc.Amount),
			UsedUserId: rc.UsedUserId,
			CreatedAt:  rc.CreatedAt.Unix(),
			UpdatedAt:  rc.UpdatedAt.Unix(),
		}
	}
	return result, nil
}
