package media

import (
	"bytes"
	"strings"
	"unicode"
)

// SniffImage reports the file extension and MIME type for a supported image.
func SniffImage(b []byte) (ext, mime string, ok bool) {
	switch {
	case bytes.HasPrefix(b, []byte("\x89PNG\r\n\x1a\n")):
		return "png", "image/png", true
	case bytes.HasPrefix(b, []byte("\xff\xd8\xff")):
		return "jpg", "image/jpeg", true
	case bytes.HasPrefix(b, []byte("GIF87a")), bytes.HasPrefix(b, []byte("GIF89a")):
		return "gif", "image/gif", true
	case len(b) > 12 && bytes.Equal(b[:4], []byte("RIFF")) && bytes.Equal(b[8:12], []byte("WEBP")):
		return "webp", "image/webp", true
	default:
		if SafeSVG(b) {
			return "svg", "image/svg+xml", true
		}
		return "", "", false
	}
}

// SafeSVG accepts simple SVG documents and rejects active content.
func SafeSVG(b []byte) bool {
	if len(b) == 0 || len(b) > 2<<20 {
		return false
	}
	lower := strings.ToLower(string(b))
	if !strings.Contains(lower, "<svg") {
		return false
	}
	compact := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, lower)
	for _, bad := range []string{
		"<script",
		"javascript:",
		"foreignobject",
		"<iframe",
		"<embed",
		"<object",
		"onload=",
		"onerror=",
		"onclick=",
		"<!entity",
	} {
		if strings.Contains(compact, bad) {
			return false
		}
	}
	return true
}
