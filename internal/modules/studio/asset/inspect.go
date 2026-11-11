package asset

import (
	"bytes"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"

	apperrors "poster.app/studio/internal/platform/errors"
)

const maxPixels = 24_000_000

type inspected struct {
	ContentType string
	Extension   string
	Width       int
	Height      int
}

func inspect(data []byte) (inspected, error) {
	kind, err := sniff(data)
	if err != nil {
		return inspected{}, err
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return inspected{}, apperrors.Invalid("图片已损坏或无法识别")
	}
	if format != kind.format {
		return inspected{}, apperrors.Invalid("图片内容与格式不一致")
	}
	if config.Width < 1 || config.Height < 1 || config.Width > 8000 || config.Height > 8000 {
		return inspected{}, apperrors.Invalid("图片尺寸不能超过 8000 像素")
	}
	if config.Width*config.Height > maxPixels {
		return inspected{}, apperrors.Invalid("图片像素过多")
	}
	return inspected{
		ContentType: kind.contentType,
		Extension:   kind.extension,
		Width:       config.Width,
		Height:      config.Height,
	}, nil
}

type sniffed struct {
	format      string
	contentType string
	extension   string
}

func sniff(data []byte) (sniffed, error) {
	switch {
	case len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF:
		return sniffed{format: "jpeg", contentType: "image/jpeg", extension: ".jpg"}, nil
	case len(data) >= 8 && bytes.Equal(data[:8], []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1A, '\n'}):
		return sniffed{format: "png", contentType: "image/png", extension: ".png"}, nil
	case len(data) >= 6 && (bytes.Equal(data[:6], []byte("GIF87a")) || bytes.Equal(data[:6], []byte("GIF89a"))):
		return sniffed{format: "gif", contentType: "image/gif", extension: ".gif"}, nil
	case len(data) >= 12 && bytes.Equal(data[:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP")):
		return sniffed{format: "webp", contentType: "image/webp", extension: ".webp"}, nil
	default:
		return sniffed{}, apperrors.Invalid("只支持 JPEG、PNG、GIF、WEBP 图片")
	}
}
