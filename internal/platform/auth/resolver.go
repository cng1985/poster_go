package auth

import (
	"context"

	"poster.app/studio/internal/platform/domain"
)

// Resolver 把原始会话令牌解析为可信 Scope。实现由账号模块提供。
type Resolver interface {
	Resolve(ctx context.Context, rawToken string) (domain.Scope, error)
}
