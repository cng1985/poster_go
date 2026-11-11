package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"poster.app/studio/internal/platform/auth"
	"poster.app/studio/internal/platform/domain"
	apperrors "poster.app/studio/internal/platform/errors"
	"poster.app/studio/internal/platform/response"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-Id")
		if !validRequestID(id) {
			id = newRequestID()
		}
		c.Header("X-Request-Id", id)
		c.Request = c.Request.WithContext(domain.WithRequestID(c.Request.Context(), id))
		c.Next()
	}
}

func ClientIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request = c.Request.WithContext(domain.WithClientIP(c.Request.Context(), c.ClientIP()))
		c.Next()
	}
}

func SecurityHeaders() gin.HandlerFunc {
	const policy = "default-src 'self'; img-src 'self' blob: data:; style-src 'self' 'unsafe-inline'; font-src 'self'; script-src 'self'; base-uri 'self'; frame-ancestors 'self'"
	return func(c *gin.Context) {
		header := c.Writer.Header()
		header.Set("X-Content-Type-Options", "nosniff")
		header.Set("X-Frame-Options", "SAMEORIGIN")
		header.Set("Referrer-Policy", "no-referrer")
		header.Set("Content-Security-Policy", policy)
		header.Set("X-Permitted-Cross-Domain-Policies", "none")
		c.Next()
	}
}

func RequestTimeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func BodyLimit(jsonMax, uploadMax int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if path != "/api" && !strings.HasPrefix(path, "/api/") {
			c.Next()
			return
		}
		limit := jsonMax
		if path == "/api/asset/upload" {
			limit = uploadMax
		}
		if c.Request.Body != nil && limit > 0 {
			limited := &response.LimitedBody{ReadCloser: c.Request.Body, Remain: limit}
			c.Request.Body = limited
			c.Set("limited_body", limited)
		}
		c.Next()
	}
}

func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			recovered := recover()
			if recovered == nil {
				return
			}
			logger.Error("panic",
				"request_id", domain.RequestID(c.Request.Context()),
				"route", c.FullPath(),
				"err", recovered,
				"stack", string(debug.Stack()),
			)
			response.HandleError(c, apperrors.Internal(errPanic(recovered)))
			c.Abort()
		}()
		c.Next()
	}
}

func AccessLog(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}
		actor := ""
		if scope, ok := domain.ScopeFrom(c.Request.Context()); ok {
			actor = strconv.FormatInt(scope.ActorID, 10)
		}
		code, _ := c.Get("app_code")
		class, _ := c.Get("error_class")
		level := slog.LevelInfo
		if strings.HasPrefix(c.Request.URL.Path, "/health/") {
			level = slog.LevelDebug
		}
		logger.Log(context.Background(), level, "http",
			"request_id", domain.RequestID(c.Request.Context()),
			"actor", actor,
			"method", c.Request.Method,
			"route", route,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"app_code", code,
			"error_class", class,
		)
	}
}

func NewAuth(resolver auth.Resolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		const prefix = "Bearer "
		if len(header) > 400 || !strings.HasPrefix(header, prefix) {
			response.HandleError(c, apperrors.Unauthorized("未登录"))
			c.Abort()
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
		scope, err := resolver.Resolve(c.Request.Context(), token)
		if err != nil {
			response.HandleError(c, err)
			c.Abort()
			return
		}
		c.Request = c.Request.WithContext(domain.WithScope(c.Request.Context(), scope))
		c.Next()
	}
}

func validRequestID(id string) bool {
	if len(id) < 8 || len(id) > 64 {
		return false
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '.' || r == '_' || r == '-':
		default:
			return false
		}
	}
	return true
}

func newRequestID() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

type panicError struct{ value any }

func errPanic(value any) error { return panicError{value: value} }

func (e panicError) Error() string {
	return "panic"
}

var _ = http.StatusOK
