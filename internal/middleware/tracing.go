// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
)

type contextKey string

const (
	// HeaderTraceID is the HTTP header that carries the distributed trace ID.
	HeaderTraceID = "X-Trace-ID"
	// HeaderSpanID carries the parent span ID for hierarchical tracing.
	HeaderSpanID = "X-Span-ID"

	ctxKeyTraceID contextKey = "trace_id"
	ctxKeySpanID  contextKey = "span_id"
)

// Tracing extracts distributed-tracing headers from the incoming request and
// injects them into the request context.  It is intentionally lightweight and
// compatible with OpenTelemetry W3C Trace Context propagation: callers that
// later wire in an otel SDK simply replace this middleware.
func Tracing() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(HeaderTraceID)
		spanID := c.GetHeader(HeaderSpanID)

		// Fall back to the request ID if no explicit trace ID is provided.
		if traceID == "" {
			if id, ok := c.Get(KeyRequestID); ok {
				traceID, _ = id.(string)
			}
		}

		ctx := context.WithValue(c.Request.Context(), ctxKeyTraceID, traceID)
		ctx = context.WithValue(ctx, ctxKeySpanID, spanID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// TraceIDFromContext retrieves the trace ID stored by the Tracing middleware.
func TraceIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(ctxKeyTraceID).(string)
	return v
}

// SpanIDFromContext retrieves the span ID stored by the Tracing middleware.
func SpanIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(ctxKeySpanID).(string)
	return v
}
