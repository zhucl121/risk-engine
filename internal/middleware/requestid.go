// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// HeaderRequestID is the canonical HTTP header name for request tracing.
	HeaderRequestID = "X-Request-ID"
	// KeyRequestID is the gin context key under which the request ID is stored.
	KeyRequestID = "request_id"
)

// RequestID injects a unique request identifier into every request.
// If the client supplies X-Request-ID it is reused; otherwise a UUID v4 is
// generated. The value is:
//   - stored in the gin context under KeyRequestID
//   - echoed back to the caller in the X-Request-ID response header
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(HeaderRequestID)
		if id == "" {
			id = uuid.New().String()
		}
		c.Set(KeyRequestID, id)
		c.Header(HeaderRequestID, id)
		c.Next()
	}
}
