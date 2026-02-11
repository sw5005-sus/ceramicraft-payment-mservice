package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/config"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/http/data"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/log"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/repository/dao/mocks"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/repository/model"
)

func initEnv() {
	config.Config = &config.Conf{
		LogConfig: &config.LogConfig{
			Level:    "debug",
			FilePath: "",
		},
	}
	log.InitLogger()
}

func TestGetRedeemCodeService(t *testing.T) {
	service1 := GetRedeemCodeService()
	service2 := GetRedeemCodeService()

	if service1 != service2 {
		t.Errorf("Expected the same instance of RedeemCodeService, got different instances")
	}
}
func TestGenerateRedeemCodes(t *testing.T) {
	initEnv()
	redeemCodeDao := new(mocks.RedeemCodeDao)
	service := &RedeemCodeServiceImpl{
		redeemCodeDao: redeemCodeDao,
	}
	ctx := context.Background()
	amount := 100
	quantity := 5
	redeemCodeDao.On("BatchInsert", ctx, mock.Anything).Return(nil)
	err := service.GenerateRedeemCodes(ctx, amount, quantity)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	redeemCodeDao.AssertNumberOfCalls(t, "BatchInsert", 1)
	// Additional checks can be added here to verify the generated codes
}

func TestGenerateRedeemCodesErr(t *testing.T) {
	initEnv()
	redeemCodeDao := new(mocks.RedeemCodeDao)
	service := &RedeemCodeServiceImpl{
		redeemCodeDao: redeemCodeDao,
	}
	ctx := context.Background()
	amount := 100
	quantity := 10
	redeemCodeDao.On("BatchInsert", ctx, mock.Anything).Return(assert.AnError)
	err := service.GenerateRedeemCodes(ctx, amount, quantity)
	if err != assert.AnError {
		t.Errorf("Expected error, got %v", err)
	}

}
func TestQueryRedeemCodes(t *testing.T) {
	initEnv()
	redeemCodeDao := new(mocks.RedeemCodeDao)
	service := &RedeemCodeServiceImpl{
		redeemCodeDao: redeemCodeDao,
	}
	ctx := context.Background()
	code := "test-code"
	used := false
	query := &data.RedeemCodeQuery{
		Limit: 10,
		Code:  &code,
		Used:  &used,
	}
	expectedRedeemCodes := []*model.RedeemCode{
		{ID: 1, Code: "test-code", Amount: 100, UsedUserId: 0, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	redeemCodeDao.On("QueryRedeemCodes", ctx, mock.Anything).Return(expectedRedeemCodes, nil)

	result, err := service.QueryRedeemCodes(ctx, query)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(result) != len(expectedRedeemCodes) {
		t.Errorf("Expected %d redeem codes, got %d", len(expectedRedeemCodes), len(result))
	}
	for i, rc := range result {
		if rc.Code != expectedRedeemCodes[i].Code {
			t.Errorf("Expected code %s, got %s", expectedRedeemCodes[i].Code, rc.Code)
		}
	}

	redeemCodeDao.AssertNumberOfCalls(t, "QueryRedeemCodes", 1)
}

func TestQueryRedeemCodesErr(t *testing.T) {
	initEnv()
	redeemCodeDao := new(mocks.RedeemCodeDao)
	service := &RedeemCodeServiceImpl{
		redeemCodeDao: redeemCodeDao,
	}
	ctx := context.Background()
	code := "test-code"
	used := false
	query := &data.RedeemCodeQuery{
		Limit: 10,
		Code:  &code,
		Used:  &used,
	}
	redeemCodeDao.On("QueryRedeemCodes", ctx, mock.Anything).Return(nil, assert.AnError)

	result, err := service.QueryRedeemCodes(ctx, query)
	if err == nil {
		t.Errorf("Expected error, got none")
	}
	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}

	redeemCodeDao.AssertNumberOfCalls(t, "QueryRedeemCodes", 1)
}
