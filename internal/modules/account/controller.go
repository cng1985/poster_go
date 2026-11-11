package account

import (
	"github.com/gin-gonic/gin"

	"poster.app/studio/internal/platform/response"
)

type Controller struct {
	resource *Resource
}

func NewController(resource *Resource) *Controller {
	return &Controller{resource: resource}
}

func (c *Controller) RegisterPublic(group *gin.RouterGroup) {
	group.POST("/account/register", c.register)
	group.POST("/account/login", c.login)
}

func (c *Controller) RegisterProtected(group *gin.RouterGroup) {
	group.POST("/account/logout", c.logout)
	group.POST("/account/one", c.one)
	group.POST("/account/update", c.update)
}

func (c *Controller) register(ctx *gin.Context) {
	var request AccountRegisterRequest
	if err := response.BindJSON(ctx, &request); err != nil {
		response.HandleError(ctx, err)
		return
	}
	result, err := c.resource.Create(ctx.Request.Context(), &request)
	response.Write(ctx, result, err)
}

func (c *Controller) login(ctx *gin.Context) {
	var request AccountLoginRequest
	if err := response.BindJSON(ctx, &request); err != nil {
		response.HandleError(ctx, err)
		return
	}
	result, err := c.resource.Login(ctx.Request.Context(), &request)
	response.Write(ctx, result, err)
}

func (c *Controller) logout(ctx *gin.Context) {
	var request AccountRequest
	if err := response.BindJSON(ctx, &request); err != nil {
		response.HandleError(ctx, err)
		return
	}
	result, err := c.resource.Logout(ctx.Request.Context(), &request)
	response.Write(ctx, result, err)
}

func (c *Controller) one(ctx *gin.Context) {
	var request AccountRequest
	if err := response.BindJSON(ctx, &request); err != nil {
		response.HandleError(ctx, err)
		return
	}
	result, err := c.resource.One(ctx.Request.Context(), &request)
	response.Write(ctx, result, err)
}

func (c *Controller) update(ctx *gin.Context) {
	var request AccountUpdateRequest
	if err := response.BindJSON(ctx, &request); err != nil {
		response.HandleError(ctx, err)
		return
	}
	result, err := c.resource.Update(ctx.Request.Context(), &request)
	response.Write(ctx, result, err)
}
