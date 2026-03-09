package metadata

import (
	"github.com/Sene4ka/cloud_storage/internal/models"
)

type CreateMetadataInput struct {
	UserID       string
	Filename     string
	OriginalName string
	Path         string
	Size         int64
	MimeType     string
	StoragePath  string
	Bucket       string
	IsPublic     bool
	Tags         map[string]string
}

type CreateMetadataOutput struct {
	File *models.File
}

type GetMetadataInput struct {
	FileID string
	UserID string
}

type GetMetadataOutput struct {
	File *models.File
}

type ListMetadataInput struct {
	UserID    string
	Page      int
	PageSize  int
	SortBy    string
	SortOrder string
	Search    string
	IsTrashed *bool
}

type ListMetadataOutput struct {
	Items    []*models.File
	Total    int64
	Page     int
	PageSize int
}

type UpdateMetadataInput struct {
	FileID       string
	UserID       string
	Filename     string
	OriginalName string
	Path         string
	IsPublic     bool
	Tags         map[string]string
}

type UpdateMetadataOutput struct {
	File *models.File
}

type CheckAccessInput struct {
	FileID string
	UserID string
}

type CheckAccessOutput struct {
	HasAccess   bool
	StoragePath string
	Bucket      string
}

type TrashFileInput struct {
	FileID string
	UserID string
}

type TrashFileOutput struct {
	Success bool
}

type RestoreFileInput struct {
	FileID string
	UserID string
}

type RestoreFileOutput struct {
	Success bool
}

type DeleteFileMetadataInput struct {
	FileID string
	UserID string
}

type DeleteFileMetadataOutput struct {
	Success bool
}
