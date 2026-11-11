package routing

import "github.com/gin-gonic/gin"

// Registrar 由业务模块实现。公开路由与认证路由必须显式登记。
type Registrar interface {
	RegisterPublic(group *gin.RouterGroup)
	RegisterProtected(group *gin.RouterGroup)
}
