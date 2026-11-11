package domain

import (
	"strconv"
	"strings"

	apperrors "poster.app/studio/internal/platform/errors"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 200
	MaxOffset       = 100000
)

// NormalizePage 把缺省的 0 视为未传，负数和超限视为明确非法值。
func NormalizePage(page, pageSize int) (int, int, error) {
	if page == 0 {
		page = DefaultPage
	}
	if pageSize == 0 {
		pageSize = DefaultPageSize
	}
	if page < 1 || pageSize < 1 || pageSize > MaxPageSize {
		return 0, 0, apperrors.Invalid("分页参数无效")
	}
	if int64(page-1) > int64(MaxOffset)/int64(pageSize) {
		return 0, 0, apperrors.Invalid("分页范围超出限制")
	}
	return page, pageSize, nil
}

func Offset(page, pageSize int) int {
	return (page - 1) * pageSize
}

// OrderBy 只接受白名单字段，避免把客户端排序文本拼进 SQL。
func OrderBy(sort, direction string, allowed map[string]string, fallbackSort, fallbackDirection string) (string, error) {
	column, ok := allowed[sort]
	if sort == "" || !ok {
		if sort == "" {
			column = allowed[fallbackSort]
		} else {
			return "", apperrors.Invalid("排序字段无效")
		}
	}
	if direction == "" {
		direction = fallbackDirection
	}
	switch strings.ToLower(direction) {
	case "asc", "desc":
	default:
		return "", apperrors.Invalid("排序方向无效")
	}
	return column + " " + strings.ToLower(direction) + ", id desc", nil
}

// ParseID 只接受十进制正整数字符串。
func ParseID(raw string) (int64, error) {
	if raw == "" {
		return 0, apperrors.Invalid("标识不正确")
	}
	for _, r := range raw {
		if r < '0' || r > '9' {
			return 0, apperrors.Invalid("标识不正确")
		}
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id < 1 {
		return 0, apperrors.Invalid("标识不正确")
	}
	return id, nil
}
