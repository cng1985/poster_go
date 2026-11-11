package asset

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"path"
	"strings"
	"unicode/utf8"

	"poster.app/studio/internal/platform/config"
	"poster.app/studio/internal/platform/domain"
	apperrors "poster.app/studio/internal/platform/errors"
	"poster.app/studio/internal/platform/response"
	"poster.app/studio/internal/platform/storage"
)

type Resource struct {
	repo    *Repository
	files   *storage.Local
	maxSize int64
}

func NewResource(repo *Repository, files *storage.Local, cfg *config.Config) *Resource {
	return &Resource{repo: repo, files: files, maxSize: cfg.Upload.MaxBytes}
}

func (r *Resource) Create(ctx context.Context, request *AssetUploadRequest) (response.Response[AssetResponse], error) {
	var zero response.Response[AssetResponse]
	owner, err := domain.ActorID(ctx)
	if err != nil {
		return zero, err
	}
	if request == nil || len(request.Bytes) == 0 {
		return zero, apperrors.Invalid("请选择图片文件")
	}
	if int64(len(request.Bytes)) > r.maxSize {
		return zero, apperrors.TooLarge("图片超过大小限制")
	}
	meta, err := inspect(request.Bytes)
	if err != nil {
		return zero, err
	}
	key, err := newKey(owner, meta.Extension)
	if err != nil {
		return zero, apperrors.Internal(err)
	}
	if err := r.files.Save(key, request.Bytes); err != nil {
		return zero, apperrors.Internal(err)
	}
	result, err := r.repo.Create(ctx, &Asset{
		OwnerID: owner, StorageKey: key, OriginalName: cleanName(request.FileName, meta.Extension),
		ContentType: meta.ContentType, ByteSize: int64(len(request.Bytes)),
		Width: meta.Width, Height: meta.Height,
	})
	if err != nil {
		_ = r.files.Remove(key)
		return zero, err
	}
	return result, nil
}

func (r *Resource) One(ctx context.Context, request *AssetRequest) (response.Response[AssetResponse], error) {
	var zero response.Response[AssetResponse]
	if _, err := domain.ActorID(ctx); err != nil {
		return zero, err
	}
	if request == nil || request.ID < 1 {
		return zero, apperrors.Invalid("素材标识不正确")
	}
	return r.repo.One(ctx, request)
}

func (r *Resource) Search(ctx context.Context, search *AssetSearch) (response.PageResponse[AssetSimple], error) {
	var zero response.PageResponse[AssetSimple]
	if _, err := domain.ActorID(ctx); err != nil {
		return zero, err
	}
	if search == nil {
		search = &AssetSearch{}
	}
	page, pageSize, err := domain.NormalizePage(search.Page, search.PageSize)
	if err != nil {
		return zero, err
	}
	search.Page, search.PageSize = page, pageSize
	return r.repo.Search(ctx, search)
}

func (r *Resource) Delete(ctx context.Context, request *AssetRequest) (response.Response[AssetResponse], error) {
	var zero response.Response[AssetResponse]
	if request == nil || request.ID < 1 {
		return zero, apperrors.Invalid("素材标识不正确")
	}
	current, err := r.repo.find(ctx, request.ID)
	if err != nil {
		return zero, err
	}
	used, err := r.repo.Referenced(ctx, request.ID)
	if err != nil {
		return zero, err
	}
	if used {
		return zero, apperrors.Conflict(10005, "这张图片还在海报里，先从作品中拿掉再删除。")
	}
	result, err := r.repo.Delete(ctx, request)
	if err != nil {
		return zero, err
	}
	_ = r.files.Remove(current.StorageKey)
	return result, nil
}

func (r *Resource) open(ctx context.Context, id int64) (*os.File, AssetResponse, error) {
	entity, err := r.repo.find(ctx, id)
	if err != nil {
		return nil, AssetResponse{}, err
	}
	file, err := r.files.Open(entity.StorageKey)
	if err != nil {
		return nil, AssetResponse{}, apperrors.NotFound("素材文件不存在")
	}
	return file, *AssetEntityToResponse(entity), nil
}

func newKey(owner int64, ext string) (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "assets/" + hex.EncodeToString([]byte{byte(owner), byte(owner >> 8)}) + "/" + hex.EncodeToString(buf) + ext, nil
}

func cleanName(name, ext string) string {
	base := path.Base(strings.ReplaceAll(name, "\\", "/"))
	base = strings.TrimSpace(base)
	if base == "." || base == "/" || base == "" {
		return "image" + ext
	}
	if utf8.RuneCountInString(base) > 80 {
		base = trim(base, 80)
	}
	return base
}

func trim(value string, max int) string {
	count := 0
	for index := range value {
		if count == max {
			return value[:index]
		}
		count++
	}
	return value
}
