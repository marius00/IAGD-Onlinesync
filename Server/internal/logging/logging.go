package logging

import (
	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
	"go.uber.org/zap"
	"log"
)

var logger *zap.Logger

const RequestID string = "requestId"

func init() {
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		log.Printf("Well this is embarrassing.. error initializing logging: %v", err)
	}
}

func Logger(ctx *gin.Context) zap.Logger {
	newLogger := logger
	if requestID, ok := ctx.Value(RequestID).(string); ok {
		newLogger = newLogger.With(zap.String("req", requestID)).With(zap.String("ip", ctx.ClientIP()))
	} else {
		requestID := uuid.NewV4().String()
		ctx.Set(RequestID, requestID)
		newLogger = newLogger.With(zap.String("req", requestID)).With(zap.String("ip", ctx.ClientIP()))
	}

	return *newLogger
}
