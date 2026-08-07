package imageutil

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/chai2010/webp"
)

var allowedImageExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

func ConvertToWebP(file *multipart.FileHeader) ([]byte, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedImageExts[ext] {
		return nil, fmt.Errorf("unsupported image format: %s", ext)
	}

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open image file: %w", err)
	}
	defer src.Close()

	if ext == ".webp" {
		data, err := io.ReadAll(src)
		if err != nil {
			return nil, fmt.Errorf("failed to read webp image: %w", err)
		}
		if _, err := webp.DecodeConfig(bytes.NewReader(data)); err != nil {
			return nil, fmt.Errorf("failed to decode webp image: %w", err)
		}
		return data, nil
	}

	img, _, err := image.Decode(src)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	var buf bytes.Buffer
	err = webp.Encode(&buf, img, &webp.Options{Lossless: false, Quality: 80})
	if err != nil {
		return nil, fmt.Errorf("failed to encode image to webp: %w", err)
	}

	return buf.Bytes(), nil
}
