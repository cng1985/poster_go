package account

import (
	"context"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"poster.app/studio/internal/platform/auth"
	"poster.app/studio/internal/platform/config"
	"poster.app/studio/internal/platform/database"
	"poster.app/studio/internal/platform/domain"
	apperrors "poster.app/studio/internal/platform/errors"
	"poster.app/studio/internal/platform/response"
)

type Repository struct {
	db    *gorm.DB
	ttl   time.Duration
	dummy string
}

func NewRepository(db *gorm.DB, cfg *config.Config) (*Repository, error) {
	dummy, err := bcrypt.GenerateFromPassword([]byte("poster-dummy-password"), authCost)
	if err != nil {
		return nil, err
	}
	return &Repository{db: db, ttl: cfg.Runtime.SessionTTL, dummy: string(dummy)}, nil
}

const authCost = 12

func (r *Repository) Create(ctx context.Context, request *AccountRegisterRequest) (response.Response[AccountSessionResponse], error) {
	var result response.Response[AccountSessionResponse]
	hash, err := auth.HashPassword(request.Password)
	if err != nil {
		return result, apperrors.Invalid("密码需要 8 到 72 位")
	}
	raw, tokenHash, err := auth.NewToken()
	if err != nil {
		return result, apperrors.Internal(err)
	}
	now := time.Now().UTC()
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		entity := &Account{
			Username:     request.Username,
			PasswordHash: hash,
			DisplayName:  request.DisplayName,
			Status:       "active",
		}
		if err := tx.Omit(clause.Associations).Create(entity).Error; err != nil {
			if strings.Contains(err.Error(), "account.username") {
				return apperrors.Conflict(10001, "用户名已存在")
			}
			return database.MapConflict(err, 409, "数据冲突")
		}
		session := &Session{
			AccountID: entity.ID,
			TokenHash: tokenHash,
			ExpiresAt: now.Add(r.ttl),
		}
		if err := tx.Omit(clause.Associations).Create(session).Error; err != nil {
			return database.MapError(err)
		}
		result = response.OK(AccountSessionResponse{
			Token:   raw,
			Account: *AccountEntityToResponse(entity),
		})
		return nil
	})
	if err != nil {
		return response.Response[AccountSessionResponse]{}, err
	}
	return result, nil
}

func (r *Repository) Login(ctx context.Context, request *AccountLoginRequest) (response.Response[AccountSessionResponse], error) {
	var result response.Response[AccountSessionResponse]
	var entity Account
	err := r.db.WithContext(ctx).Where("username = ?", request.Username).Take(&entity).Error
	if err != nil {
		_ = auth.CheckPassword(r.dummy, request.Password)
		if err != nil && !isNotFound(err) {
			return result, database.MapError(err)
		}
		return result, badLogin()
	}
	if entity.Status != "active" || auth.CheckPassword(entity.PasswordHash, request.Password) != nil {
		return result, badLogin()
	}
	raw, tokenHash, err := auth.NewToken()
	if err != nil {
		return result, apperrors.Internal(err)
	}
	session := &Session{
		AccountID: entity.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().UTC().Add(r.ttl),
	}
	if err := r.db.WithContext(ctx).Omit(clause.Associations).Create(session).Error; err != nil {
		return result, database.MapError(err)
	}
	return response.OK(AccountSessionResponse{Token: raw, Account: *AccountEntityToResponse(&entity)}), nil
}

func (r *Repository) One(ctx context.Context, request *AccountRequest) (response.Response[AccountResponse], error) {
	var result response.Response[AccountResponse]
	actor, err := actorMatches(ctx, request.ID)
	if err != nil {
		return result, err
	}
	entity, err := r.findAccount(ctx, r.db, actor)
	if err != nil {
		return result, err
	}
	return response.OK(*AccountEntityToResponse(entity)), nil
}

func (r *Repository) Update(ctx context.Context, request *AccountUpdateRequest) (response.Response[AccountResponse], error) {
	var result response.Response[AccountResponse]
	actor, err := domain.ActorID(ctx)
	if err != nil {
		return result, err
	}
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Account{}).
			Where("id = ? and status = ?", actor, "active").
			Updates(map[string]any{
				"display_name": request.DisplayName,
				"updated_at":   time.Now().UTC(),
			})
		if res.Error != nil {
			return database.MapError(res.Error)
		}
		if res.RowsAffected == 0 {
			return apperrors.NotFound("账号不存在")
		}
		entity, err := r.findAccount(ctx, tx, actor)
		if err != nil {
			return err
		}
		result = response.OK(*AccountEntityToResponse(entity))
		return nil
	})
	if err != nil {
		return response.Response[AccountResponse]{}, err
	}
	return result, nil
}

func (r *Repository) Logout(ctx context.Context, request *AccountRequest) (response.Response[AccountResponse], error) {
	var result response.Response[AccountResponse]
	scope, ok := domain.ScopeFrom(ctx)
	if !ok {
		return result, apperrors.Unauthorized("未登录")
	}
	if request != nil && request.ID != 0 && request.ID != scope.ActorID {
		return result, apperrors.NotFound("账号不存在")
	}
	now := time.Now().UTC()
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Session{}).
			Where("id = ? and account_id = ? and revoked_at is null", scope.SessionID, scope.ActorID).
			Update("revoked_at", now)
		if res.Error != nil {
			return database.MapError(res.Error)
		}
		entity, err := r.findAccount(ctx, tx, scope.ActorID)
		if err != nil {
			return err
		}
		result = response.OK(*AccountEntityToResponse(entity))
		return nil
	})
	if err != nil {
		return response.Response[AccountResponse]{}, err
	}
	return result, nil
}

func (r *Repository) findAccount(ctx context.Context, tx *gorm.DB, id int64) (*Account, error) {
	var entity Account
	err := tx.WithContext(ctx).Where("id = ?", id).Take(&entity).Error
	if err != nil {
		return nil, database.MapError(err)
	}
	return &entity, nil
}

func actorMatches(ctx context.Context, requested int64) (int64, error) {
	actor, err := domain.ActorID(ctx)
	if err != nil {
		return 0, err
	}
	if requested != 0 && requested != actor {
		return 0, apperrors.NotFound("账号不存在")
	}
	return actor, nil
}

func badLogin() error {
	return apperrors.WithCode(apperrors.ClassUnauthorized, 10002, "用户名或密码错误")
}

func isNotFound(err error) bool {
	appErr, ok := apperrors.As(database.MapError(err))
	return ok && appErr.Class == apperrors.ClassNotFound
}
