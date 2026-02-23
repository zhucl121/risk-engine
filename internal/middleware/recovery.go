// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery returns a gin middleware that catches panics, logs the stack trace,
// and responds with HTTP 500 so that a single bad request cannot take down the
// whole server.
func Recovery(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				reqID, _ := c.Get(KeyRequestID)
				log.Error("panic recovered",
					zap.Any("error", fmt.Sprintf("%v", r)),
					zap.Any("request_id", reqID),
					zap.String("stack", string(stack)),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":       "INTERNAL_ERROR",
					"message":    "an unexpected error occurred",
					"request_id": reqID,
				})
			}
		}()
		c.Next()
	}
}
