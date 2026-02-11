package grpc

import (
	"context"
	"fmt"

	"github.com/sw5005-sus/ceramicraft-payment-mservice/common/paymentpb"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/log"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/repository/model"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/service"
)

type PaymentService struct {
	paymentpb.UnimplementedPaymentServiceServer
}

func (s *PaymentService) PayOrder(ctx context.Context, req *paymentpb.PayOrderRequest) (*paymentpb.PayOrderResponse, error) {
	log.Logger.Infof("[grpc-svr] method=PayOrder, req=%v", req)
	resp := &paymentpb.PayOrderResponse{}
	errmsg := ""
	if req.UserId == 0 || req.Amount <= 0 || req.BizId == "" {
		log.Logger.Warnf("Invalid request: %v", req)
		resp.Code = int32(paymentpb.RespCode_BAD_REQUEST)
		errmsg = "UserId, Amount and BizId must be provided and valid"
		resp.ErrorMsg = &errmsg
		return resp, nil
	}
	changeLog, err := service.GetUserAccountService().PayOrder(ctx, int(req.UserId), req.BizId, int(req.Amount))
	if err != nil {
		resp.Code = int32(paymentpb.RespCode_UNKNOWN_ERROR)
		errmsg = err.Error()
		resp.ErrorMsg = &errmsg
		return resp, nil
	}
	log.Logger.Infof("Payment successful, change log: %+v", changeLog)
	resp.Code = int32(paymentpb.RespCode_SUCCESS)
	resp.PayOrderInfo = &paymentpb.PayOrderInfo{
		PayOrderId:  genPayOrderId(changeLog),
		Amount:      int32(changeLog.Amount),
		UserId:      req.UserId,
		CreatedTime: changeLog.CreatedAt.Unix(),
	}
	return resp, nil
}

func (s *PaymentService) QueryPayOrder(ctx context.Context, req *paymentpb.PayOrderQueryRequest) (*paymentpb.PayOrderQueryResponse, error) {
	log.Logger.Infof("[grpc-svr] method=QueryPayOrder, req=%v", req)
	resp := &paymentpb.PayOrderQueryResponse{}
	errmsg := ""
	if req.UserId == 0 && req.BizId == nil {
		log.Logger.Warnf("Either UserId or BizId must be provided")
		resp.Code = int32(paymentpb.RespCode_BAD_REQUEST)
		errmsg = "Either UserId or BizId must be provided"
		resp.ErrorMsg = &errmsg
		return resp, nil
	}
	changeLogs, err := service.GetUserAccountService().GetUserPayHistory(ctx, req)
	if err != nil {
		resp.Code = int32(paymentpb.RespCode_UNKNOWN_ERROR)
		errmsg = err.Error()
		resp.ErrorMsg = &errmsg
		return resp, nil
	}
	ret := make([]*paymentpb.PayOrderInfo, 0)
	for _, cLog := range changeLogs {
		ret = append(ret, &paymentpb.PayOrderInfo{
			PayOrderId:  genPayOrderId(cLog),
			Amount:      int32(cLog.Amount),
			UserId:      req.UserId,
			CreatedTime: cLog.CreatedAt.Unix(),
		})
	}
	resp.Code = int32(paymentpb.RespCode_SUCCESS)
	resp.PayOrderInfos = ret
	return resp, nil
}

func genPayOrderId(changeLog *model.UserAccountChangeLog) string {
	return fmt.Sprintf("%d_%s_%d", changeLog.AccountId, changeLog.IdempotentKey, changeLog.ID)
}
