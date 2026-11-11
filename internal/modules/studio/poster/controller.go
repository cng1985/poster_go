package poster

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

func (c *Controller) RegisterPublic(*gin.RouterGroup) {}

func (c *Controller) RegisterProtected(group *gin.RouterGroup) {
	group.POST("/poster/create", c.create)
	group.POST("/poster/update", c.update)
	group.POST("/poster/one", c.one)
	group.POST("/poster/search", c.search)
	group.POST("/poster/delete", c.delete)
	group.POST("/poster/duplicate", c.duplicate)
}

func (c *Controller) create(ctx *gin.Context) {
	var request PosterRequest
	if err := response.BindJSON(ctx, &request); err != nil {
		response.HandleError(ctx, err)
		return
	}
	result, err := c.resource.Create(ctx.Request.Context(), &request)
	response.Write(ctx, result, err)
}

func (c *Controller) update(ctx *gin.Context) {
	var request PosterRequest
	if err := response.BindJSON(ctx, &request); err != nil {
		response.HandleError(ctx, err)
		return
	}
	result, err := c.resource.Update(ctx.Request.Context(), &request)
	response.Write(ctx, result, err)
}

func (c *Controller) one(ctx *gin.Context) {
	var request PosterRequest
	if err := response.BindJSON(ctx, &request); err != nil {
		response.HandleError(ctx, err)
		return
	}
	result, err := c.resource.One(ctx.Request.Context(), &request)
	response.Write(ctx, result, err)
}

func (c *Controller) search(ctx *gin.Context) {
	var search PosterSearch
	if err := response.BindJSON(ctx, &search); err != nil {
		response.HandleError(ctx, err)
		return
	}
	result, err := c.resource.Search(ctx.Request.Context(), &search)
	response.WritePage(ctx, result, err)
}

func (c *Controller) delete(ctx *gin.Context) {
	var request PosterRequest
	if err := response.BindJSON(ctx, &request); err != nil {
		response.HandleError(ctx, err)
		return
	}
	result, err := c.resource.Delete(ctx.Request.Context(), &request)
	response.Write(ctx, result, err)
}

func (c *Controller) duplicate(ctx *gin.Context) {
	var request PosterDuplicateRequest
	if err := response.BindJSON(ctx, &request); err != nil {
		response.HandleError(ctx, err)
		return
	}
	result, err := c.resource.Duplicate(ctx.Request.Context(), &request)
	response.Write(ctx, result, err)
}
