package account

import (
	"context"
	"time"

	"gorm.io/gorm"

	"poster.app/studio/internal/platform/auth"
	"poster.app/studio/internal/platform/domain"
	apperrors "poster.app/studio/internal/platform/errors"
)

type SessionResolver struct {
	db *gorm.DB
}

func NewSessionResolver(db *gorm.DB) *SessionResolver {
	return &SessionResolver{db: db}
}

func (s *SessionResolver) Resolve(ctx context.Context, rawToken string) (domain.Scope, error) {
	if len(rawToken) < 16 || len(rawToken) > 200 {
		return domain.Scope{}, apperrors.Unauthorized("未登录")
	}
	var session Session
	err := s.db.WithContext(ctx).Where("token_hash = ?", auth.HashToken(rawToken)).Take(&session).Error
	if err != nil {
		return domain.Scope{}, apperrors.Unauthorized("未登录")
	}
	if session.RevokedAt != nil || !session.ExpiresAt.After(time.Now().UTC()) {
		return domain.Scope{}, apperrors.Unauthorized("登录已失效")
	}
	var entity Account
	if err := s.db.WithContext(ctx).Where("id = ?", session.AccountID).Take(&entity).Error; err != nil {
		return domain.Scope{}, apperrors.Unauthorized("未登录")
	}
	if entity.Status != "active" {
		return domain.Scope{}, apperrors.Forbidden("账号不可用")
	}
	return domain.Scope{
		ActorID:   entity.ID,
		SessionID: session.ID,
		RequestID: domain.RequestID(ctx),
	}, nil
}
