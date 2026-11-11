package domain

import apperrors "poster.app/studio/internal/platform/errors"

var errUnauthenticated = apperrors.Unauthorized("未登录")
