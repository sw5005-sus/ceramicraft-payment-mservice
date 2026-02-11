package api

import (
	"net/http"

	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/http/data"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/log"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/service"

	"github.com/gin-gonic/gin"
)

const (
	maxGenCodeSize = 100
)

// GenerateRedeemCodes godoc
// @Summary Generate redeem codes
// @Description Generate redeem codes
// @Tags RedeemCodes
// @Accept json
// @Produce json
// @Param amount query int true "Amount for each redeem code"
// @Param count query int true "Number of redeem codes to generate"
// @Success 200 {object} data.BaseResponse{data=data.RedeemCodeGenResult}
// @Failure 400 {object} data.BaseResponse
// @Router /payment-ms/v1/merchant/redeem-codes/generate [post]
func GenerateRedeemCodes(c *gin.Context) {
	var req data.RedeemCodeGenRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Logger.Errorf("GenerateRedeemCodes bind error: %v", err)
		c.JSON(http.StatusBadRequest, data.BaseResponse{ErrMsg: err.Error()})
		return
	}
	if req.Amount <= 0 || req.Count <= 0 || req.Count > maxGenCodeSize {
		log.Logger.Error("GenerateRedeemCodes error: invalid amount or count")
		c.JSON(http.StatusBadRequest, data.BaseResponse{ErrMsg: "Amount must be positive and count must be between 1 and 100"})
		return
	}
	err := service.GetRedeemCodeService().GenerateRedeemCodes(c.Request.Context(), req.Amount, req.Count)
	if err != nil {
		log.Logger.Errorf("GenerateRedeemCodes service error: %v", err)
		c.JSON(http.StatusInternalServerError, data.BaseResponse{ErrMsg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data.BaseResponse{Data: map[string]interface{}{"gen_success_cnt": req.Count}})
}

// QueryRedeemCodes godoc
// @Summary Query redeem codes
// @Description Query redeem codes
// @Tags RedeemCodes
// @Accept json
// @Produce json
// @Param query query data.RedeemCodeQuery false "Redeem code to search for"
// Success 200 {object} data.BaseResponse{data=[]data.RedeemCodeVO}
// @Failure 400 {object} data.BaseResponse
// @Router /payment-ms/v1/merchant/redeem-codes [get]
func QueryRedeemCodes(c *gin.Context) {
	var query data.RedeemCodeQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		log.Logger.Errorf("QueryRedeemCodes bind error:", err)
		c.JSON(http.StatusBadRequest, data.BaseResponse{ErrMsg: err.Error()})
		return
	}
	if query.Code == nil && query.Used == nil {
		log.Logger.Error("QueryRedeemCodes error: at least one query parameter must be provided")
		c.JSON(http.StatusBadRequest, data.BaseResponse{ErrMsg: "At least one query parameter must be provided"})
		return
	}
	ret, err := service.GetRedeemCodeService().QueryRedeemCodes(c.Request.Context(), &query)
	if err != nil {
		log.Logger.Errorf("QueryRedeemCodes service error: $v", err)
		c.JSON(http.StatusInternalServerError, data.BaseResponse{ErrMsg: err.Error()})
		return
	}
	c.JSON(200, data.BaseResponse{Data: ret})
}
