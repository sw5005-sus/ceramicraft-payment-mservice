package service

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/repository/dao/mocks"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/repository/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGetUserAccountService(t *testing.T) {
	t.Run("should return the same instance on multiple calls", func(t *testing.T) {
		service1 := GetUserAccountService()
		service2 := GetUserAccountService()
		if service1 != service2 {
			t.Errorf("Expected the same instance, got different instances")
		}
	})

	t.Run("should initialize the service correctly", func(t *testing.T) {
		service := GetUserAccountService()
		if service == nil {
			t.Errorf("Expected service to be initialized, got nil")
		}
		if _, ok := service.(*UserAccountServiceImpl); !ok {
			t.Errorf("Expected service to be of type UserAccountServiceImpl, got %T", service)
		}
	})
}
func TestCreateUserAccount(t *testing.T) {
	ctx := context.Background()
	userId := 1
	initEnv()
	t.Run("should create a new user account", func(t *testing.T) {
		userAccountDao := new(mocks.UserAccountDao)
		service := &UserAccountServiceImpl{
			userAccountDao: userAccountDao,
		}
		userAccountDao.On("GetUserAccountByUserID", ctx, userId).Return(nil, nil).Once()
		userAccountDao.On("CreateUserAccount", ctx, mock.Anything).Return(nil).Once()
		account, err := service.CreateUserAccount(ctx, userId)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if account == nil {
			t.Errorf("Expected account to be created, got nil")
		}
		if account != nil && account.UserId != userId {
			t.Errorf("Expected user ID %d, got %d", userId, account.UserId)
		}
		userAccountDao.AssertCalled(t, "GetUserAccountByUserID", ctx, userId)
		userAccountDao.AssertCalled(t, "CreateUserAccount", ctx, mock.Anything)
	})

	t.Run("should return existing account if already created", func(t *testing.T) {
		userAccountDao := new(mocks.UserAccountDao)
		service := &UserAccountServiceImpl{
			userAccountDao: userAccountDao,
		}
		userAccountDao.On("GetUserAccountByUserID", ctx, userId).Return(&model.UserAccount{
			UserId: userId,
		}, nil).Once()
		account, err := service.CreateUserAccount(ctx, userId)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if account == nil {
			t.Errorf("Expected account to be returned, got nil")
		}
		if account != nil && account.UserId != userId {
			t.Errorf("Expected user ID %d, got %d", userId, account.UserId)
		}
		userAccountDao.AssertNotCalled(t, "CreateUserAccount", ctx, mock.Anything)
	})

	t.Run("should return error if user ID is invalid", func(t *testing.T) {
		userAccountDao := new(mocks.UserAccountDao)
		service := &UserAccountServiceImpl{
			userAccountDao: userAccountDao,
		}
		_, err := service.CreateUserAccount(ctx, -1)
		if err == nil {
			t.Errorf("Expected error for invalid user ID, got nil")
		}
		userAccountDao.AssertNotCalled(t, "GetUserAccountByUserID", ctx, mock.Anything)
	})

	t.Run("should return error if GetUserAccountByUserID fails", func(t *testing.T) {
		userAccountDao := new(mocks.UserAccountDao)
		service := &UserAccountServiceImpl{
			userAccountDao: userAccountDao,
		}
		userAccountDao.On("GetUserAccountByUserID", ctx, userId).Return(nil, assert.AnError).Once()
		_, err := service.CreateUserAccount(ctx, userId)
		if err == nil {
			t.Errorf("Expected error from GetUserAccountByUserID, got nil")
		}
		userAccountDao.AssertCalled(t, "GetUserAccountByUserID", ctx, userId)
	})

	t.Run("should return error if CreateUserAccount fails", func(t *testing.T) {
		userAccountDao := new(mocks.UserAccountDao)
		service := &UserAccountServiceImpl{
			userAccountDao: userAccountDao,
		}
		userAccountDao.On("GetUserAccountByUserID", ctx, userId).Return(nil, nil).Once()
		userAccountDao.On("CreateUserAccount", ctx, mock.Anything).Return(assert.AnError).Once()
		_, err := service.CreateUserAccount(ctx, userId)
		if err == nil {
			t.Errorf("Expected error from CreateUserAccount, got nil")
		}
		userAccountDao.AssertCalled(t, "GetUserAccountByUserID", ctx, userId)
		userAccountDao.AssertCalled(t, "CreateUserAccount", ctx, mock.Anything)
	})
}

func TestGetUserAccountByUserID(t *testing.T) {
	ctx := context.Background()
	userId := 1
	initEnv()
	t.Run("should return user account if found", func(t *testing.T) {
		userAccountDao := new(mocks.UserAccountDao)
		service := &UserAccountServiceImpl{
			userAccountDao: userAccountDao,
		}
		userAccount := &model.UserAccount{UserId: userId}
		userAccountDao.On("GetUserAccountByUserID", ctx, userId).Return(userAccount, nil).Once()

		account, err := service.GetUserAccountByUserID(ctx, userId)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if account == nil {
			t.Errorf("Expected account to be returned, got nil")
		}
		if account != nil && account.UserId != userId {
			t.Errorf("Expected user ID %d, got %d", userId, account.UserId)
		}
		userAccountDao.AssertCalled(t, "GetUserAccountByUserID", ctx, userId)
	})

	t.Run("should return nil if account not found", func(t *testing.T) {
		userAccountDao := new(mocks.UserAccountDao)
		service := &UserAccountServiceImpl{
			userAccountDao: userAccountDao,
		}
		userAccountDao.On("GetUserAccountByUserID", ctx, userId).Return(nil, nil).Once()

		account, err := service.GetUserAccountByUserID(ctx, userId)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if account != nil {
			t.Errorf("Expected account to be nil, got %v", account)
		}
		userAccountDao.AssertCalled(t, "GetUserAccountByUserID", ctx, userId)
	})

	t.Run("should return error if GetUserAccountByUserID fails", func(t *testing.T) {
		userAccountDao := new(mocks.UserAccountDao)
		service := &UserAccountServiceImpl{
			userAccountDao: userAccountDao,
		}
		userAccountDao.On("GetUserAccountByUserID", ctx, userId).Return(nil, assert.AnError).Once()

		_, err := service.GetUserAccountByUserID(ctx, userId)
		if err == nil {
			t.Errorf("Expected error from GetUserAccountByUserID, got nil")
		}
		userAccountDao.AssertCalled(t, "GetUserAccountByUserID", ctx, userId)
	})
}
func TestPayOrder(t *testing.T) {
	ctx := context.Background()
	userId := 1
	bizId := "test-biz-id"
	amount := 100
	initEnv()

	t.Run("should successfully pay order", func(t *testing.T) {
		userAccountDao := new(mocks.UserAccountDao)
		userAccountChangeLogDao := new(mocks.UserAccountChangeLogDAO)
		service := &UserAccountServiceImpl{
			userAccountDao:          userAccountDao,
			userAccountChangeLogDao: userAccountChangeLogDao,
			txBeginner:              &fakeTx{DB: initMemDb(t)},
		}

		userAccount := &model.UserAccount{ID: 1, UserId: userId, Balance: 200}
		userAccountDao.On("GetUserAccountByUserID", ctx, userId).Return(userAccount, nil).Once()
		userAccountChangeLogDao.On("CreateChangeLogInTransaction", ctx, mock.Anything, mock.Anything).Return(nil).Once()
		userAccountDao.On("SubtractBalanceInTransaction", ctx, userId, amount, userAccount.Balance, mock.Anything).Return(1, nil).Once()

		changeLog, err := service.PayOrder(ctx, userId, bizId, amount)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if changeLog == nil {
			t.Errorf("Expected change log to be created, got nil")
		}
		if changeLog != nil && (changeLog.Amount != amount || changeLog.OpType != model.OpTypePayment || changeLog.IdempotentKey != bizId) {
			t.Errorf("Change log fields do not match expected values")
		}
	})

	t.Run("should return error if user account not found", func(t *testing.T) {
		userAccountDao := new(mocks.UserAccountDao)
		service := &UserAccountServiceImpl{
			userAccountDao: userAccountDao,
		}

		userAccountDao.On("GetUserAccountByUserID", ctx, userId).Return(nil, nil).Once()

		_, err := service.PayOrder(ctx, userId, bizId, amount)
		if err == nil {
			t.Errorf("Expected error for non-existing user account, got nil")
		}
	})

	t.Run("should return error if insufficient balance", func(t *testing.T) {
		userAccountDao := new(mocks.UserAccountDao)
		service := &UserAccountServiceImpl{
			userAccountDao: userAccountDao,
		}

		userAccount := &model.UserAccount{ID: 1, UserId: userId, Balance: 50}
		userAccountDao.On("GetUserAccountByUserID", ctx, userId).Return(userAccount, nil).Once()

		_, err := service.PayOrder(ctx, userId, bizId, amount)
		if err == nil {
			t.Errorf("Expected error for insufficient balance, got nil")
		}
	})

	t.Run("should return error if transaction fails", func(t *testing.T) {
		userAccountDao := new(mocks.UserAccountDao)
		userAccountChangeLogDao := new(mocks.UserAccountChangeLogDAO)
		service := &UserAccountServiceImpl{
			userAccountDao:          userAccountDao,
			userAccountChangeLogDao: userAccountChangeLogDao,
			txBeginner:              &fakeTx{DB: initMemDb(t)},
		}

		userAccount := &model.UserAccount{ID: 1, UserId: userId, Balance: 200}
		userAccountDao.On("GetUserAccountByUserID", ctx, userId).Return(userAccount, nil).Once()
		userAccountChangeLogDao.On("CreateChangeLogInTransaction", ctx, mock.Anything, mock.Anything).Return(nil).Once()
		userAccountDao.On("SubtractBalanceInTransaction", ctx, userId, amount, userAccount.Balance, mock.Anything).Return(0, nil).Once() // Simulate failure

		_, err := service.PayOrder(ctx, userId, bizId, amount)
		if err == nil {
			t.Errorf("Expected error from transaction failure, got nil")
		}
	})
}
func TestUserAccountTopUp(t *testing.T) {
	ctx := context.Background()
	userId := 1
	redeemCode := "valid-redeem-code"
	initEnv()

	t.Run("should successfully top up user account", func(t *testing.T) {
		userAccountDao := new(mocks.UserAccountDao)
		redeemCodeDao := new(mocks.RedeemCodeDao)
		userAccountChangeLogDao := new(mocks.UserAccountChangeLogDAO)
		service := &UserAccountServiceImpl{
			userAccountDao:          userAccountDao,
			redeemCodeDao:           redeemCodeDao,
			userAccountChangeLogDao: userAccountChangeLogDao,
			txBeginner:              &fakeTx{DB: initMemDb(t)},
		}

		userAccount := &model.UserAccount{ID: 1, UserId: userId, Balance: 100}
		redeemCodeRecord := &model.RedeemCode{Amount: 50}

		userAccountDao.On("GetUserAccountByUserID", ctx, userId).Return(userAccount, nil).Twice()
		redeemCodeDao.On("GetByCode", ctx, redeemCode).Return(redeemCodeRecord, nil).Once()
		userAccountChangeLogDao.On("CreateChangeLogInTransaction", ctx, mock.Anything, mock.Anything).Return(nil).Once()
		userAccountDao.On("AddBalanceInTransaction", ctx, userId, int(redeemCodeRecord.Amount), userAccount.Balance, mock.Anything).Return(nil).Once()
		redeemCodeDao.On("UseRedeemCodeInTransaction", ctx, mock.Anything, mock.Anything).Return(1, nil).Once()

		account, code, err := service.UserAccountTopUp(ctx, userId, redeemCode)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if account == nil {
			t.Errorf("Expected account to be returned, got nil")
		}
		if code == nil {
			t.Errorf("Expected redeem code to be returned, got nil")
		}
	})

	t.Run("should return error if user account not found", func(t *testing.T) {
		userAccountDao := new(mocks.UserAccountDao)
		redeemCodeDao := new(mocks.RedeemCodeDao)
		service := &UserAccountServiceImpl{
			userAccountDao: userAccountDao,
			redeemCodeDao:  redeemCodeDao,
		}

		userAccountDao.On("GetUserAccountByUserID", ctx, userId).Return(nil, nil).Once()

		_, _, err := service.UserAccountTopUp(ctx, userId, redeemCode)
		if err == nil {
			t.Errorf("Expected error for non-existing user account, got nil")
		}
	})

	t.Run("should return error if redeem code not found", func(t *testing.T) {
		userAccountDao := new(mocks.UserAccountDao)
		redeemCodeDao := new(mocks.RedeemCodeDao)
		service := &UserAccountServiceImpl{
			userAccountDao: userAccountDao,
			redeemCodeDao:  redeemCodeDao,
		}

		userAccount := &model.UserAccount{ID: 1, UserId: userId, Balance: 100}
		userAccountDao.On("GetUserAccountByUserID", ctx, userId).Return(userAccount, nil).Once()
		redeemCodeDao.On("GetByCode", ctx, redeemCode).Return(nil, nil).Once()

		_, _, err := service.UserAccountTopUp(ctx, userId, redeemCode)
		if err == nil {
			t.Errorf("Expected error for non-existing redeem code, got nil")
		}
	})

	t.Run("should return error if redeem code already used", func(t *testing.T) {
		userAccountDao := new(mocks.UserAccountDao)
		redeemCodeDao := new(mocks.RedeemCodeDao)
		service := &UserAccountServiceImpl{
			userAccountDao: userAccountDao,
			redeemCodeDao:  redeemCodeDao,
		}

		userAccount := &model.UserAccount{ID: 1, UserId: userId, Balance: 100}
		redeemCodeRecord := &model.RedeemCode{UsedUserId: userId}

		userAccountDao.On("GetUserAccountByUserID", ctx, userId).Return(userAccount, nil).Once()
		redeemCodeDao.On("GetByCode", ctx, redeemCode).Return(redeemCodeRecord, nil).Once()

		_, _, err := service.UserAccountTopUp(ctx, userId, redeemCode)
		if err == nil {
			t.Errorf("Expected error for already used redeem code, got nil")
		}
	})

	t.Run("should return error if transaction fails", func(t *testing.T) {
		userAccountDao := new(mocks.UserAccountDao)
		redeemCodeDao := new(mocks.RedeemCodeDao)
		userAccountChangeLogDao := new(mocks.UserAccountChangeLogDAO)
		service := &UserAccountServiceImpl{
			userAccountDao:          userAccountDao,
			redeemCodeDao:           redeemCodeDao,
			userAccountChangeLogDao: userAccountChangeLogDao,
			txBeginner:              &fakeTx{DB: initMemDb(t)},
		}

		userAccount := &model.UserAccount{ID: 1, UserId: userId, Balance: 100}
		redeemCodeRecord := &model.RedeemCode{Amount: 50}

		userAccountDao.On("GetUserAccountByUserID", ctx, userId).Return(userAccount, nil).Once()
		redeemCodeDao.On("GetByCode", ctx, redeemCode).Return(redeemCodeRecord, nil).Once()
		redeemCodeDao.On("UseRedeemCodeInTransaction", ctx, mock.Anything, mock.Anything).Return(1, nil).Once()
		userAccountChangeLogDao.On("CreateChangeLogInTransaction", ctx, mock.Anything, mock.Anything).Return(assert.AnError).Once()

		_, _, err := service.UserAccountTopUp(ctx, userId, redeemCode)
		if err == nil {
			t.Errorf("Expected error from transaction failure, got nil")
		}
	})
}

type fakeTx struct{ *gorm.DB }

func (f *fakeTx) Transaction(fn func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	return fn(f.DB) // Pass through the same DB instance
}

func initMemDb(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	assert.NoError(t, db.AutoMigrate(&model.UserAccount{}, &model.UserAccountChangeLog{}, &model.RedeemCode{}))
	return db
}
