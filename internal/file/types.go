package file

import (
	"time"

	"github.com/Sene4ka/cloud_storage/internal/models"
)

type InitiateUploadInput struct {
	UserID   string
	Filename string
	Path     string
	MimeType string
	Size     int64
	IsPublic bool
	Tags     map[string]string
}

type InitiateUploadOutput struct {
	FileID       string
	UploadURL    string
	UploadMethod string
	Headers      map[string]string
	ExpiresIn    int64
}

type CompleteUploadInput struct {
	FileID string
	UserID string
	ETag   string
}

type CompleteUploadOutput struct {
	StoragePath string
	CreatedAt   time.Time
}

type GetDownloadLinkInput struct {
	FileID    string
	UserID    string
	ExpiresIn int64
}

type GetDownloadLinkOutput struct {
	DownloadURL string
	Method      string
	Headers     map[string]string
	ExpiresIn   int64
}

type DeleteFileInput struct {
	FileID string
	UserID string
}

type DeleteFileOutput struct {
	Success bool
}

type GetFileInfoInput struct {
	FileID string
	UserID string
}

type GetFileInfoOutput struct {
	File *models.File
}
