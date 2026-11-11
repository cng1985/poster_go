package asset

import (
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"poster.app/studio/internal/platform/domain"
	apperrors "poster.app/studio/internal/platform/errors"
	"poster.app/studio/internal/platform/response"
)

type Controller struct {
	resource *Resource
	maxBytes int64
}

func NewController(resource *Resource) *Controller {
	return &Controller{resource: resource, maxBytes: resource.maxSize}
}

func (c *Controller) RegisterPublic(*gin.RouterGroup) {}

func (c *Controller) RegisterProtected(group *gin.RouterGroup) {
	group.POST("/asset/upload", c.upload)
	group.POST("/asset/one", c.one)
	group.POST("/asset/search", c.search)
	group.POST("/asset/delete", c.remove)
	group.GET("/asset/content/:id", c.content)
}

func (c *Controller) upload(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		response.HandleError(ctx, apperrors.Invalid("请选择图片文件"))
		return
	}
	opened, err := file.Open()
	if err != nil {
		response.HandleError(ctx, apperrors.Invalid("无法读取上传文件"))
		return
	}
	defer opened.Close()
	data, err := io.ReadAll(io.LimitReader(opened, c.maxBytes+1))
	if err != nil {
		response.HandleError(ctx, apperrors.Internal(err))
		return
	}
	if int64(len(data)) > c.maxBytes {
		response.HandleError(ctx, apperrors.TooLarge("图片超过大小限制"))
		return
	}
	result, err := c.resource.Create(ctx.Request.Context(), &AssetUploadRequest{FileName: file.Filename, Bytes: data})
	response.Write(ctx, result, err)
}

func (c *Controller) one(ctx *gin.Context) {
	var request AssetRequest
	if err := response.BindJSON(ctx, &request); err != nil {
		response.HandleError(ctx, err)
		return
	}
	result, err := c.resource.One(ctx.Request.Context(), &request)
	response.Write(ctx, result, err)
}

func (c *Controller) search(ctx *gin.Context) {
	var search AssetSearch
	if err := response.BindJSON(ctx, &search); err != nil {
		response.HandleError(ctx, err)
		return
	}
	result, err := c.resource.Search(ctx.Request.Context(), &search)
	response.WritePage(ctx, result, err)
}

func (c *Controller) remove(ctx *gin.Context) {
	var request AssetRequest
	if err := response.BindJSON(ctx, &request); err != nil {
		response.HandleError(ctx, err)
		return
	}
	result, err := c.resource.Delete(ctx.Request.Context(), &request)
	response.Write(ctx, result, err)
}

func (c *Controller) content(ctx *gin.Context) {
	id, err := domain.ParseID(ctx.Param("id"))
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	file, meta, err := c.resource.open(ctx.Request.Context(), id)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	defer file.Close()
	ctx.Header("Content-Type", meta.ContentType)
	ctx.Header("X-Content-Type-Options", "nosniff")
	ctx.Header("Cache-Control", "private, max-age=3600")
	ctx.Header("Content-Disposition", "inline; filename=\""+safeDisposition(meta.Name)+"\"")
	ctx.Status(http.StatusOK)
	_, _ = io.Copy(ctx.Writer, file)
}

func safeDisposition(name string) string {
	var b strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '_' || r == '-' {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "image"
	}
	return b.String()
}
