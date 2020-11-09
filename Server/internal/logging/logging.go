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
		log.Printf("Well this is embarassing.. error initializing logging: %v", err)
	}
}

func Logger(ctx *gin.Context) zap.Logger {
	newLogger := logger
	if requestId, ok := ctx.Value(RequestID).(string); ok {
		newLogger = newLogger.With(zap.String("req", requestId))
	} else {
		requestId := uuid.NewV4().String()
		ctx.Set(RequestID, requestId)
		newLogger = newLogger.With(zap.String("req", requestId))
	}

	return *newLogger
}
