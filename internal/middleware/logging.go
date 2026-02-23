// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger returns a gin middleware that emits a structured access log entry
// for every request after it completes.
// Fields: method, path, status, latency_ms, client_ip, request_id.
func Logger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		if q := c.Request.URL.RawQuery; q != "" {
			path += "?" + q
		}

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		reqID, _ := c.Get(KeyRequestID)

		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Int64("latency_ms", latency.Milliseconds()),
			zap.String("client_ip", c.ClientIP()),
			zap.Any("request_id", reqID),
		}
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		switch {
		case status >= 500:
			log.Error("access", fields...)
		case status >= 400:
			log.Warn("access", fields...)
		default:
			log.Info("access", fields...)
		}
	}
}
