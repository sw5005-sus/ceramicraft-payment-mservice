package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	bizerror "github.com/sw5005-sus/ceramicraft-payment-mservice/common/biz_error"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/common/paymentpb"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/log"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/repository"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/repository/dao"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/repository/model"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/utils"
	"gorm.io/gorm"
)

type UserAccountService interface {
	CreateUserAccount(ctx context.Context, userId int) (*model.UserAccount, error)
	GetUserAccountByUserID(ctx context.Context, userId int) (*model.UserAccount, error)
	UserAccountTopUp(ctx context.Context, userId int, redeemCode string) (*model.UserAccount, *model.RedeemCode, error)
	PayOrder(ctx context.Context, userId int, bizId string, amount int) (*model.UserAccountChangeLog, error)
	GetUserPayHistory(ctx context.Context, query *paymentpb.PayOrderQueryRequest) ([]*model.UserAccountChangeLog, error)
}

var (
	userAccountServiceInstance UserAccountService
	userAccountServiceOnce     sync.Once
)

func GetUserAccountService() UserAccountService {
	userAccountServiceOnce.Do(func() {
		userAccountServiceInstance = &UserAccountServiceImpl{
			userAccountDao:          dao.GetUserAccountDao(),
			userAccountChangeLogDao: dao.GetUserAccountChangeLogDAO(),
			redeemCodeDao:           dao.GetRedeemCodeDao(),
			txBeginner:              repository.DB,
		}
	})
	return userAccountServiceInstance
}

type UserAccountServiceImpl struct {
	userAccountDao          dao.UserAccountDao
	userAccountChangeLogDao dao.UserAccountChangeLogDAO
	redeemCodeDao           dao.RedeemCodeDao
	txBeginner              repository.TxBeginner
}

const userAccountNoSize = 12

// CreateUserAccount implements UserAccountService.
func (u *UserAccountServiceImpl) CreateUserAccount(ctx context.Context, userId int) (*model.UserAccount, error) {
	if userId <= 0 {
		log.Logger.Errorf("Invalid user ID: %d", userId)
		return nil, fmt.Errorf("invalid user ID")
	}
	account, err := u.userAccountDao.GetUserAccountByUserID(ctx, userId)
	if err != nil {
		log.Logger.Errorf("Failed to get user account for user ID %d: %v", userId, err)
		return nil, err
	}
	if account != nil {
		log.Logger.Warnf("User account already exists for user ID %d", userId)
		return account, nil
	}
	account = &model.UserAccount{
		UserId:    userId,
		AccountNo: utils.GenRedeemCode(userAccountNoSize),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = u.userAccountDao.CreateUserAccount(ctx, account)
	if err != nil {
		log.Logger.Errorf("Failed to create user account for user ID %d: %v", userId, err)
		return nil, err
	}
	log.Logger.Infof("Successfully created user account for user ID %d, accountNo: %s", userId, account.AccountNo)
	return account, nil
}

// GetUserAccountByUserID implements UserAccountService.
func (u *UserAccountServiceImpl) GetUserAccountByUserID(ctx context.Context, userId int) (*model.UserAccount, error) {
	account, err := u.userAccountDao.GetUserAccountByUserID(ctx, userId)
	if err != nil {
		log.Logger.Errorf("Failed to get user account for user ID %d: %v", userId, err)
		return nil, err
	}
	if account == nil {
		log.Logger.Warnf("User account not found for user ID %d", userId)
		return nil, nil
	}
	return account, nil
}

// PayOrder implements UserAccountService.
func (u *UserAccountServiceImpl) PayOrder(ctx context.Context, userId int, bizId string, amount int) (*model.UserAccountChangeLog, error) {
	account, err := u.userAccountDao.GetUserAccountByUserID(ctx, userId)
	if err != nil {
		log.Logger.Errorf("Failed to get user account for user ID %d: %v", userId, err)
		return nil, &bizerror.BizError{Code: int(paymentpb.RespCode_UNKNOWN_ERROR), Message: "failed to get user account", Err: err}
	}
	if account == nil {
		log.Logger.Warnf("User account not found for user ID %d", userId)
		return nil, &bizerror.BizError{Code: int(paymentpb.RespCode_ACCOUNT_NOT_EXIST), Message: "user account not found"}
	}
	if account.Balance < amount {
		log.Logger.Warnf("Insufficient balance for user ID %d: balance %d, required %d", userId, account.Balance, amount)
		return nil, &bizerror.BizError{Code: int(paymentpb.RespCode_INSUFFICIENT_BALANCE), Message: "insufficient balance"}
	}
	changeLog := &model.UserAccountChangeLog{
		AccountId:     account.ID,
		OpType:        model.OpTypePayment,
		Amount:        amount,
		IdempotentKey: bizId,
		CreatedAt:     time.Now(),
	}
	err = u.txBeginner.Transaction(func(tx *gorm.DB) error {
		err = u.userAccountChangeLogDao.CreateChangeLogInTransaction(ctx, changeLog, tx)
		if err != nil {
			log.Logger.Errorf("Failed to create user account change log for user ID %d: %v", userId, err)
			return err
		}
		rowsAffected, err := u.userAccountDao.SubtractBalanceInTransaction(ctx, userId, amount, account.Balance, tx)
		if err != nil {
			log.Logger.Errorf("Failed to subtract balance for user ID %d: %v", userId, err)
			return err
		}
		if rowsAffected == 0 {
			log.Logger.Errorf("No user account found to subtract balance for user ID %d", userId)
			return &bizerror.BizError{Code: int(paymentpb.RespCode_UNKNOWN_ERROR), Message: "failed to subtract balance"}
		}
		return nil
	})
	if err != nil {
		log.Logger.Errorf("Transaction failed for user ID %d: %v", userId, err)
		return nil, &bizerror.BizError{Code: int(paymentpb.RespCode_UNKNOWN_ERROR), Message: "transaction failed", Err: err}
	}
	log.Logger.Infof("Successfully paid order for user ID %d, amount %d, biz ID %s", userId, amount, bizId)
	return changeLog, nil
}

// UserAccountTopUp implements UserAccountService.
func (u *UserAccountServiceImpl) UserAccountTopUp(ctx context.Context, userId int, redeemCode string) (*model.UserAccount, *model.RedeemCode, error) {
	account, err := u.userAccountDao.GetUserAccountByUserID(ctx, userId)
	if err != nil {
		log.Logger.Errorf("Failed to get user account for user ID %d: %v", userId, err)
		return nil, nil, err
	}
	if account == nil {
		log.Logger.Warnf("User account not found for user ID %d", userId)
		return nil, nil, fmt.Errorf("user account not found")
	}
	redeemCodeRecord, err := u.redeemCodeDao.GetByCode(ctx, redeemCode)
	if err != nil {
		log.Logger.Errorf("Failed to get redeem code %s: %v", redeemCode, err)
		return nil, nil, fmt.Errorf("failed to get redeem code")
	}
	if redeemCodeRecord == nil {
		log.Logger.Warnf("Redeem code not found: %s", redeemCode)
		return nil, nil, fmt.Errorf("invalid redeem code")
	}
	if redeemCodeRecord.UsedUserId != 0 {
		log.Logger.Warnf("Redeem code already used: %s", redeemCode)
		return nil, nil, fmt.Errorf("redeem code already used")
	}
	changeLog := &model.UserAccountChangeLog{
		AccountId:     account.ID,
		OpType:        model.OpTypeTopUp,
		Amount:        redeemCodeRecord.Amount,
		IdempotentKey: redeemCode,
		CreatedAt:     time.Now(),
	}
	err = u.txBeginner.Transaction(func(tx *gorm.DB) error {
		redeemCodeRecord.UsedUserId = userId
		ret, err := u.redeemCodeDao.UseRedeemCodeInTransaction(ctx, redeemCodeRecord, tx)
		if err != nil {
			log.Logger.Errorf("Failed to mark redeem code %s as used: %v", redeemCode, err)
			return err
		}
		if ret == 0 {
			log.Logger.Errorf("Redeem code %s was already used by another user", redeemCode)
			return fmt.Errorf("redeem code was already used")
		}
		log.Logger.Infof("Redeem code %s marked as used by user ID %d", redeemCode, userId)
		err = u.userAccountChangeLogDao.CreateChangeLogInTransaction(ctx, changeLog, tx)
		if err != nil {
			log.Logger.Errorf("Failed to create user account change log for user ID %d: %v", userId, err)
			return err
		}
		log.Logger.Infof("User account change log created for user ID %d, amount %d, redeem code %s", userId, redeemCodeRecord.Amount, redeemCode)
		err = u.userAccountDao.AddBalanceInTransaction(ctx, userId, int(redeemCodeRecord.Amount), account.Balance, tx)
		if err != nil {
			log.Logger.Errorf("Failed to add balance for user ID %d: %v", userId, err)
			return err
		}
		log.Logger.Infof("Balance added for user ID %d, amount %d", userId, redeemCodeRecord.Amount)
		return nil
	})
	if err != nil {
		log.Logger.Errorf("Transaction failed for user ID %d: %v", userId, err)
		return nil, nil, err
	}
	log.Logger.Infof("Successfully topped up user account for user ID %d, amount %d, redeem code %s", userId, redeemCodeRecord.Amount, redeemCode)
	userAccount, err := u.userAccountDao.GetUserAccountByUserID(ctx, userId)
	return userAccount, redeemCodeRecord, err
}

func (u *UserAccountServiceImpl) GetUserPayHistory(ctx context.Context, query *paymentpb.PayOrderQueryRequest) ([]*model.UserAccountChangeLog, error) {
	var accountId *int
	if query.UserId > 0 {
		account, err := u.userAccountDao.GetUserAccountByUserID(ctx, int(query.UserId))
		if err != nil {
			log.Logger.Errorf("Failed to get user account for user ID %d: %v", query.UserId, err)
			return nil, &bizerror.BizError{Code: int(paymentpb.RespCode_UNKNOWN_ERROR), Message: "failed to get user account", Err: err}
		}
		if account == nil {
			log.Logger.Warnf("User account not found for user ID %d", query.UserId)
			return nil, &bizerror.BizError{Code: int(paymentpb.RespCode_ACCOUNT_NOT_EXIST), Message: "user account not found"}
		}
		accountId = &account.ID
	}
	changeLogs, err := u.userAccountChangeLogDao.QueryChangeLogs(ctx,
		&model.UserAccountChangeLogQuery{
			AccountId:     accountId,
			IdempotentKey: query.BizId,
			OpType:        model.OpTypePayment,
		})
	if err != nil {
		log.Logger.Errorf("Failed to query user account change logs: %v", err)
		return nil, &bizerror.BizError{Code: int(paymentpb.RespCode_UNKNOWN_ERROR), Message: "failed to query change logs", Err: err}
	}
	return changeLogs, nil
}
