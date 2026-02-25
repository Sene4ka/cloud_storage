package models

import (
	"time"

	"github.com/google/uuid"
)

type File struct {
	ID           string            `db:"id" json:"id"`
	UserID       string            `db:"user_id" json:"user_id"`
	Filename     string            `db:"filename" json:"filename"`
	OriginalName string            `db:"original_name" json:"original_name"`
	Size         int64             `db:"size" json:"size"`
	MimeType     string            `db:"mime_type" json:"mime_type"`
	StoragePath  string            `db:"storage_path" json:"storage_path"`
	Bucket       string            `db:"bucket" json:"bucket"`
	IsPublic     bool              `db:"is_public" json:"is_public"`
	Tags         map[string]string `db:"tags" json:"tags"`
	CreatedAt    time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time         `db:"updated_at" json:"updated_at"`
}

func NewFile(userID, filename, originalName, mimeType, storagePath, bucket string, size int64, isPublic bool, tags map[string]string) *File {
	return &File{
		ID:           uuid.New().String(),
		UserID:       userID,
		Filename:     filename,
		OriginalName: originalName,
		Size:         size,
		MimeType:     mimeType,
		StoragePath:  storagePath,
		Bucket:       bucket,
		IsPublic:     isPublic,
		Tags:         tags,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}
