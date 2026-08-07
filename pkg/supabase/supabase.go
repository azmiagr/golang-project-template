package supabase

import (
	"bytes"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/azmiagr/golang-project-template/pkg/imageutil"
	"github.com/google/uuid"
	storage_go "github.com/supabase-community/storage-go"
)

type Supabase struct {
	client storage_go.Client
}

const (
	imageCacheControl = "31536000"
	pdfCacheControl   = "604800"
)

type Interface interface {
	UploadFile(file *multipart.FileHeader) (string, error)
	UploadPDF(file *multipart.FileHeader) (string, error)
	DeleteFile(fileURL string) error
	DeleteFileByPath(filePath string) error
	DeleteMultipleFiles(fileURLs []string) error
	UploadWebP(data []byte, folder string) (string, error)
}

func Init() Interface {
	url := fmt.Sprintf("%s/storage/v1", os.Getenv("SUPABASE_URL"))
	client := storage_go.NewClient(url, os.Getenv("SUPABASE_TOKEN"), nil)

	return Supabase{
		client: *client,
	}
}

func UploadOptionalImage(storage Interface, file *multipart.FileHeader, maxSize int64, sizeErr string) (string, error) {
	if file == nil {
		return "", nil
	}

	if file.Size > maxSize {
		return "", errors.New(sizeErr)
	}

	return storage.UploadFile(file)
}

func DeleteFileIfPresent(storage Interface, fileURL string) error {
	if fileURL == "" {
		return nil
	}

	return storage.DeleteFile(fileURL)
}

func (s Supabase) UploadFile(file *multipart.FileHeader) (string, error) {
	data, err := imageutil.ConvertToWebP(file)
	if err != nil {
		return "", err
	}

	path := uuid.NewString() + ".webp"
	contentType := "image/webp"
	cacheControl := imageCacheControl

	_, err = s.client.UploadFile(
		os.Getenv("SUPABASE_BUCKET"),
		path,
		bytes.NewReader(data),
		storage_go.FileOptions{
			ContentType:  &contentType,
			CacheControl: &cacheControl,
		},
	)

	if err != nil {
		return "", err
	}

	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s",
		os.Getenv("SUPABASE_URL"),
		os.Getenv("SUPABASE_BUCKET"),
		path,
	)

	return publicURL, nil
}

func (s Supabase) UploadPDF(file *multipart.FileHeader) (string, error) {
	if filepath.Ext(file.Filename) != ".pdf" {
		return "", fmt.Errorf("only PDF files are allowed")
	}

	if file.Size > 15*1024*1024 {
		return "", fmt.Errorf("file size exceeds 15MB limit")
	}

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	path := uuid.NewString() + ".pdf"
	contentType := "application/pdf"
	cacheControl := pdfCacheControl

	_, err = s.client.UploadFile(
		os.Getenv("SUPABASE_BUCKET"),
		path,
		src,
		storage_go.FileOptions{
			ContentType:  &contentType,
			CacheControl: &cacheControl,
		},
	)

	if err != nil {
		return "", err
	}

	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s",
		os.Getenv("SUPABASE_URL"),
		os.Getenv("SUPABASE_BUCKET"),
		path,
	)

	return publicURL, nil
}

func (s Supabase) DeleteFile(fileURL string) error {
	baseURL := fmt.Sprintf("%s/storage/v1/object/public/%s/",
		os.Getenv("SUPABASE_URL"),
		os.Getenv("SUPABASE_BUCKET"),
	)

	if len(fileURL) <= len(baseURL) {
		return fmt.Errorf("invalid file URL format")
	}

	filePath := fileURL[len(baseURL):]
	if filePath == "" {
		return fmt.Errorf("file path is empty")
	}

	return s.DeleteFileByPath(filePath)
}

func (s Supabase) DeleteFileByPath(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	_, err := s.client.RemoveFile(
		os.Getenv("SUPABASE_BUCKET"),
		[]string{filePath},
	)

	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

func (s Supabase) DeleteMultipleFiles(fileURLs []string) error {
	if len(fileURLs) == 0 {
		return fmt.Errorf("no files to delete")
	}

	baseURL := fmt.Sprintf("%s/storage/v1/object/public/%s/",
		os.Getenv("SUPABASE_URL"),
		os.Getenv("SUPABASE_BUCKET"),
	)

	var filePaths []string
	for _, fileURL := range fileURLs {
		if len(fileURL) > len(baseURL) {
			filePath := fileURL[len(baseURL):]
			if filePath != "" {
				filePaths = append(filePaths, filePath)
			}
		}
	}

	if len(filePaths) == 0 {
		return fmt.Errorf("no valid file paths found")
	}

	_, err := s.client.RemoveFile(
		os.Getenv("SUPABASE_BUCKET"),
		filePaths,
	)

	if err != nil {
		return fmt.Errorf("failed to delete files: %w", err)
	}

	return nil
}

func (s Supabase) UploadWebP(data []byte, folder string) (string, error) {
	path := folder + "/" + uuid.NewString() + ".webp"
	contentType := "image/webp"
	cacheControl := imageCacheControl

	_, err := s.client.UploadFile(
		os.Getenv("SUPABASE_BUCKET"),
		path,
		bytes.NewReader(data),
		storage_go.FileOptions{
			ContentType:  &contentType,
			CacheControl: &cacheControl,
		},
	)
	if err != nil {
		return "", err
	}

	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s",
		os.Getenv("SUPABASE_URL"),
		os.Getenv("SUPABASE_BUCKET"),
		path,
	)

	return publicURL, nil
}
