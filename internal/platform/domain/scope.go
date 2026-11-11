package domain

import "context"

type scopeKey struct{}
type requestIDKey struct{}
type clientIPKey struct{}

// Scope 是中间件写入的可信身份，不接受客户端自行声明用户。
// 本项目是单租户产品，不设置 TenantID。
type Scope struct {
	ActorID   int64
	SessionID int64
	RequestID string
}

func WithScope(ctx context.Context, scope Scope) context.Context {
	if scope.RequestID == "" {
		scope.RequestID = RequestID(ctx)
	}
	return context.WithValue(ctx, scopeKey{}, scope)
}

func ScopeFrom(ctx context.Context) (Scope, bool) {
	scope, ok := ctx.Value(scopeKey{}).(Scope)
	if !ok || scope.ActorID <= 0 {
		return Scope{}, false
	}
	return scope, true
}

func ActorID(ctx context.Context) (int64, error) {
	scope, ok := ScopeFrom(ctx)
	if !ok {
		return 0, errUnauthenticated
	}
	return scope.ActorID, nil
}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

func WithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, clientIPKey{}, ip)
}

func ClientIP(ctx context.Context) string {
	ip, _ := ctx.Value(clientIPKey{}).(string)
	return ip
}
