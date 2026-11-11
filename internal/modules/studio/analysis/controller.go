package analysis

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
	group.POST("/analysis/create", c.create)
	group.POST("/analysis/one", c.one)
	group.POST("/analysis/search", c.search)
}

func (c *Controller) create(ctx *gin.Context) {
	var request AnalysisRequest
	if err := response.BindJSON(ctx, &request); err != nil {
		response.HandleError(ctx, err)
		return
	}
	result, err := c.resource.Create(ctx.Request.Context(), &request)
	response.Write(ctx, result, err)
}

func (c *Controller) one(ctx *gin.Context) {
	var request AnalysisRequest
	if err := response.BindJSON(ctx, &request); err != nil {
		response.HandleError(ctx, err)
		return
	}
	result, err := c.resource.One(ctx.Request.Context(), &request)
	response.Write(ctx, result, err)
}

func (c *Controller) search(ctx *gin.Context) {
	var search AnalysisSearch
	if err := response.BindJSON(ctx, &search); err != nil {
		response.HandleError(ctx, err)
		return
	}
	result, err := c.resource.Search(ctx.Request.Context(), &search)
	response.WritePage(ctx, result, err)
}
