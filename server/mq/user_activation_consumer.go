package mq

import (
	"context"
	"encoding/json"

	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/log"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/service"
)

type UserActivationMessage struct {
	UserID       int   `json:"user_id"`
	ActivateTime int64 `json:"activate_time"`
}

func userActivationProcess(msg []byte) error {
	var activationMsg UserActivationMessage
	err := json.Unmarshal(msg, &activationMsg)
	if err != nil {
		log.Logger.Warnf("Failed to unmarshal user activation message: %s", string(msg))
		return nil
	}
	userAccount, err := service.GetUserAccountService().CreateUserAccount(context.Background(), activationMsg.UserID)
	if err != nil {
		log.Logger.Errorf("Failed to create user account for user ID %d: %v", activationMsg.UserID, err)
		return err
	}
	log.Logger.Infof("User account created for user ID %d: %+v", activationMsg.UserID, userAccount)
	return nil
}
