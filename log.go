package main

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type LoggerMiddleware struct {
	handler http.Handler
	logger  *zap.Logger
}

func NewLoggerMiddleware(handler http.Handler, logger *zap.Logger) *LoggerMiddleware {
	return &LoggerMiddleware{handler: handler, logger: logger}
}

func (l *LoggerMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	begin := time.Now()
	l.handler.ServeHTTP(w, r)
	l.logger.Info("served request",
		zap.Int64("duration", time.Since(begin).Milliseconds()),
		zap.String("path", r.URL.Path),
		zap.String("method", r.Method))
}
