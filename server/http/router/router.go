package router

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	auditclient "github.com/sw5005-sus/ceramicraft-audit-client"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/config"
	_ "github.com/sw5005-sus/ceramicraft-payment-mservice/server/docs"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/http/api"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/http/data"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/log"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/metrics"
	"github.com/sw5005-sus/ceramicraft-user-mservice/common/middleware"
	swaggerFiles "github.com/swaggo/files"
	gs "github.com/swaggo/gin-swagger"
)

const (
	serviceURIPrefix = "/payment-ms/v1"
)

func NewRouter() *gin.Engine {
	r := gin.Default()
	auditMiddleware := auditclient.AuditMiddleware(
		"payment-ms",
		config.Config.AuditGrpcConfig.Host,
		config.Config.AuditGrpcConfig.Port)
	basicGroup := r.Group(serviceURIPrefix)
	{
		basicGroup.Use(metrics.MetricsMiddleware())
		basicGroup.GET("/metrics", gin.WrapH(promhttp.Handler()))
		basicGroup.GET("/swagger/*any", gs.WrapHandler(
			swaggerFiles.Handler,
			gs.URL("/payment-ms/v1/swagger/doc.json"),
		))
		basicGroup.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})
	}

	v1Authed := basicGroup.Group("")
	{
		v1Authed.Use(otelgin.Middleware(data.ServiceName), log.TraceLoggerMiddleware(), middleware.AuthMiddleware())
		v1Authed.GET("/merchant/redeem-codes", middleware.RequireRoles("merchant_admin"), auditMiddleware, api.QueryRedeemCodes)
		v1Authed.POST("/merchant/redeem-codes/generate", middleware.RequireRoles("merchant_admin"), auditMiddleware, api.GenerateRedeemCodes)
		v1Authed.POST("/customer/pay-accounts/self/top-ups", api.TopUpUserPayAccount)
		v1Authed.GET("/customer/pay-accounts/self", api.GetUserPayAccountInfo)
	}
	return r
}
