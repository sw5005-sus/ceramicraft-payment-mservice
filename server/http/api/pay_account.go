package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/http/data"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/log"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/service"
)

// GetUserPayAccountInfo godoc
// @Summary Get user pay account info
// @Description Get user pay account info
// @Tags PayAccount
// @Accept json
// @Produce json
// @Success 200 {object} data.BaseResponse{data=data.UserPayAccount}
// @Failure 400 {object} data.BaseResponse
// @Failure 404 {object} data.BaseResponse
// @Failure 500 {object} data.BaseResponse
// @Router /payment-ms/v1/customer/pay-accounts/self [get]
func GetUserPayAccountInfo(c *gin.Context) {
	var userId int
	if userIdInterface, exists := c.Get("userID"); exists {
		userId = userIdInterface.(int)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found in context"})
		return
	}
	userAccount, err := service.GetUserAccountService().GetUserAccountByUserID(c.Request.Context(), userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, data.BaseResponse{ErrMsg: err.Error()})
		return
	}
	if userAccount == nil {
		c.JSON(http.StatusNotFound, data.BaseResponse{ErrMsg: "User pay account not found"})
		return
	}

	c.JSON(http.StatusOK, data.BaseResponse{
		Data: &data.UserPayAccount{
			UserId:    userAccount.UserId,
			Balance:   userAccount.Balance,
			AccountNo: userAccount.GetHiddenAccountNo(),
			CreatedAt: userAccount.CreatedAt.Unix(),
			UpdatedAt: userAccount.UpdatedAt.Unix(),
		}})
}

// TopUpUserPayAccount godoc
// @Summary Top up user pay account
// @Description Top up user pay account using a redeem code
// @Tags PayAccount
// @Accept json
// @Produce json
// @Param topup body data.UserPayAccountTopUpRequest true "Top up request"
// @Success 200 {object} data.BaseResponse{data=data.UserPayAccountTopUpResult}
// @Failure 400 {object} data.BaseResponse
// @Router /payment-ms/v1/customer/pay-accounts/self/top-ups [post]
func TopUpUserPayAccount(c *gin.Context) {
	var userId int
	if userIdInterface, exists := c.Get("userID"); exists {
		userId = userIdInterface.(int)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found in context"})
		return
	}
	log.Logger.Infof("TopUpUserPayAccount called for user ID: %d", userId)
	var req data.UserPayAccountTopUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, data.BaseResponse{ErrMsg: err.Error()})
		return
	}

	account, redeemCode, err := service.GetUserAccountService().UserAccountTopUp(c.Request.Context(), userId, req.RedeemCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, data.BaseResponse{ErrMsg: err.Error()})
		return
	}

	c.JSON(http.StatusOK, data.BaseResponse{Data: &data.UserPayAccountTopUpResult{TopUpAmount: redeemCode.Amount, CurrentBalance: account.Balance}})
}
