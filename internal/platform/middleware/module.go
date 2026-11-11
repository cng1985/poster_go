package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"middleware",
	fx.Provide(
		fx.Annotate(
			NewAuth,
			fx.ResultTags(`name:"authMW"`),
		),
	),
)

// 确保命名结果的类型稳定，避免 Fx 推断成具体别名。
var _ = fx.Annotate(func() gin.HandlerFunc { return nil })
