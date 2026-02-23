// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yourorg/riskengine/internal/metrics"
)

// Metrics returns a gin middleware that records HTTP request duration and
// total count into Prometheus using the riskengine_http_* collectors.
//
// It uses the matched route pattern (c.FullPath()) as the `path` label so that
// high-cardinality paths (e.g. /api/v1/decision?requestId=…) do not explode
// the label space.
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		metrics.ActiveRequests.Inc()

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}
		elapsed := time.Since(start).Seconds()

		metrics.HTTPRequestDuration.
			WithLabelValues(c.Request.Method, path, status).
			Observe(elapsed)
		metrics.HTTPRequestTotal.
			WithLabelValues(c.Request.Method, path, status).
			Inc()
		metrics.ActiveRequests.Dec()
	}
}
