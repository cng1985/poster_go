package database

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	apperrors "poster.app/studio/internal/platform/errors"
)

func MapError(err error) error {
	return MapConflict(err, 409, "数据冲突")
}

func MapConflict(err error, code int, uniqueMessage string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperrors.NotFound("记录不存在")
	}
	lower := strings.ToLower(err.Error())
	if strings.Contains(lower, "unique constraint failed") ||
		strings.Contains(lower, "duplicate entry") ||
		strings.Contains(lower, "duplicate key") ||
		strings.Contains(lower, "cannot insert duplicate") {
		return apperrors.Conflict(code, uniqueMessage)
	}
	return apperrors.Internal(err)
}
