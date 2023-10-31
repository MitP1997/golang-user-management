package utils

import (
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func AddKeyToContextLogger(ctx *gin.Context, key, value string) (logger *zap.Logger) {
	logger = GetContextLogger(ctx)
	key = toSnakeCase(key)
	coreField := zapcore.Field{Key: key, Type: zapcore.StringType, String: value}
	logger = logger.With(coreField)
	SetContextLogger(ctx, logger)
	return
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
