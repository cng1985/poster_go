package account

import (
	"context"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"poster.app/studio/internal/platform/config"
	"poster.app/studio/internal/platform/domain"
	apperrors "poster.app/studio/internal/platform/errors"
	"poster.app/studio/internal/platform/limit"
	"poster.app/studio/internal/platform/response"
)

var usernamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{2,31}$`)

type Resource struct {
	repo      *Repository
	failures  *limit.FailureWindow
	registers *limit.Window
}

func NewResource(repo *Repository, cfg *config.Config) *Resource {
	return &Resource{
		repo:      repo,
		failures:  limit.NewFailureWindow(cfg.Auth.LoginMax, cfg.Runtime.LoginWindow),
		registers: limit.NewWindow(5, time.Hour),
	}
}

func (r *Resource) Create(ctx context.Context, request *AccountRegisterRequest) (response.Response[AccountSessionResponse], error) {
	var zero response.Response[AccountSessionResponse]
	if request == nil {
		return zero, apperrors.Invalid("请求不能为空")
	}
	request.Username = strings.TrimSpace(request.Username)
	request.DisplayName = strings.TrimSpace(request.DisplayName)
	if !usernamePattern.MatchString(request.Username) {
		return zero, apperrors.Invalid("用户名需要以字母开头，3 到 32 位，只能包含字母、数字和下划线")
	}
	if !validPassword(request.Password) {
		return zero, apperrors.Invalid("密码需要 8 到 72 位，并同时包含字母和数字")
	}
	if request.DisplayName == "" {
		request.DisplayName = request.Username
	}
	if utf8.RuneCountInString(request.DisplayName) > 40 {
		return zero, apperrors.Invalid("显示名称不能超过 40 个字")
	}
	ip := clientKey(ctx)
	if err := r.registers.Allow("register:" + ip); err != nil {
		return zero, err
	}
	return r.repo.Create(ctx, request)
}

func (r *Resource) Login(ctx context.Context, request *AccountLoginRequest) (response.Response[AccountSessionResponse], error) {
	var zero response.Response[AccountSessionResponse]
	if request == nil {
		return zero, apperrors.Invalid("请求不能为空")
	}
	request.Username = strings.TrimSpace(request.Username)
	if request.Username == "" || request.Password == "" {
		return zero, apperrors.Invalid("请输入用户名和密码")
	}
	if len(request.Username) > 32 || len(request.Password) > 72 {
		return zero, apperrors.Invalid("请输入用户名和密码")
	}
	ip := clientKey(ctx)
	if err := r.failures.Rejected(ip); err != nil {
		return zero, err
	}
	result, err := r.repo.Login(ctx, request)
	if err != nil {
		if appErr, ok := apperrors.As(err); ok && appErr.Code == 10002 {
			r.failures.Fail(ip)
		}
		return zero, err
	}
	r.failures.Success(ip)
	return result, nil
}

func (r *Resource) One(ctx context.Context, request *AccountRequest) (response.Response[AccountResponse], error) {
	if request == nil {
		request = &AccountRequest{}
	}
	return r.repo.One(ctx, request)
}

func (r *Resource) Update(ctx context.Context, request *AccountUpdateRequest) (response.Response[AccountResponse], error) {
	var zero response.Response[AccountResponse]
	if request == nil {
		return zero, apperrors.Invalid("请求不能为空")
	}
	request.DisplayName = strings.TrimSpace(request.DisplayName)
	if request.DisplayName == "" || utf8.RuneCountInString(request.DisplayName) > 40 {
		return zero, apperrors.Invalid("显示名称需要 1 到 40 个字")
	}
	if _, err := domain.ActorID(ctx); err != nil {
		return zero, err
	}
	return r.repo.Update(ctx, request)
}

func (r *Resource) Logout(ctx context.Context, request *AccountRequest) (response.Response[AccountResponse], error) {
	if request == nil {
		request = &AccountRequest{}
	}
	return r.repo.Logout(ctx, request)
}

func validPassword(password string) bool {
	if len(password) < 8 || len(password) > 72 {
		return false
	}
	var letter, digit bool
	for _, r := range password {
		if unicode.IsLetter(r) {
			letter = true
		}
		if unicode.IsDigit(r) {
			digit = true
		}
	}
	return letter && digit
}

func clientKey(ctx context.Context) string {
	ip := domain.ClientIP(ctx)
	if ip == "" {
		return "unknown"
	}
	return ip
}
