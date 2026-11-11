package response

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"poster.app/studio/internal/platform/domain"
	apperrors "poster.app/studio/internal/platform/errors"
)

type Response[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data,omitempty"`
}

type PageResponse[T any] struct {
	Data     []T   `json:"data"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

func OK[T any](data T) Response[T] {
	return Response[T]{Code: 200, Msg: "success", Data: data}
}

func SuccessResponse[T any](c *gin.Context, data Response[T]) {
	c.JSON(http.StatusOK, data)
}

func SuccessPage[T any](c *gin.Context, data PageResponse[T]) {
	if data.Data == nil {
		data.Data = []T{}
	}
	c.JSON(http.StatusOK, data)
}

func Write[T any](c *gin.Context, result Response[T], err error) {
	if err != nil {
		HandleError(c, err)
		return
	}
	if result.Code != 200 {
		HandleError(c, apperrors.Internal(errors.New("响应契约异常")))
		return
	}
	SuccessResponse(c, result)
}

func WritePage[T any](c *gin.Context, result PageResponse[T], err error) {
	if err != nil {
		HandleError(c, err)
		return
	}
	SuccessPage(c, result)
}

// ErrBodyTooLarge 由请求体限制读取器返回，绑定层据此映射为 413。
var ErrBodyTooLarge = errors.New("body too large")

func BindJSON(c *gin.Context, dst any) error {
	if err := c.ShouldBindJSON(dst); err != nil {
		if errors.Is(err, ErrBodyTooLarge) || bodyOverflow(c) {
			return apperrors.TooLarge("请求体过大")
		}
		return apperrors.Invalid("请求格式不正确")
	}
	if bodyOverflow(c) {
		return apperrors.TooLarge("请求体过大")
	}
	return nil
}

func bodyOverflow(c *gin.Context) bool {
	value, ok := c.Get("limited_body")
	if !ok {
		return false
	}
	body, ok := value.(*LimitedBody)
	return ok && body.Overflow
}

func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	if c.Request != nil && errors.Is(c.Request.Context().Err(), context.Canceled) {
		return
	}
	if errors.Is(err, context.DeadlineExceeded) {
		err = apperrors.Unavailable("处理超时")
	}
	appErr, ok := apperrors.As(err)
	if !ok {
		if errors.Is(err, ErrBodyTooLarge) {
			appErr = apperrors.TooLarge("请求体过大")
		} else {
			appErr = apperrors.Internal(err)
		}
	}
	code := appErr.AppCode()
	c.Set("app_code", code)
	c.Set("error_class", string(appErr.Class))
	if code >= 500 {
		slog.Default().Error("request failed",
			"request_id", domain.RequestID(c.Request.Context()),
			"route", c.FullPath(),
			"class", string(appErr.Class),
			"code", code,
			"err", appErr.Error(),
		)
	}
	c.JSON(http.StatusOK, Response[any]{
		Code: code,
		Msg:  appErr.Public,
	})
}

// LimitedBody 限制真实读取量，而不只检查 Content-Length。
type LimitedBody struct {
	io.ReadCloser
	Remain   int64
	Overflow bool
}

func (b *LimitedBody) Read(p []byte) (int, error) {
	if b.Overflow {
		return 0, ErrBodyTooLarge
	}
	if b.Remain <= 0 {
		var tmp [1]byte
		n, err := b.ReadCloser.Read(tmp[:])
		if n > 0 {
			b.Overflow = true
			return 0, ErrBodyTooLarge
		}
		return 0, err
	}
	if int64(len(p)) > b.Remain {
		p = p[:b.Remain]
	}
	n, err := b.ReadCloser.Read(p)
	b.Remain -= int64(n)
	if err != nil {
		return n, err
	}
	return n, nil
}
