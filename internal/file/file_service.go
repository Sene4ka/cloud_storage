package file

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/Sene4ka/cloud_storage/configs"
	"github.com/Sene4ka/cloud_storage/internal/models"
	"github.com/Sene4ka/cloud_storage/internal/utils"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type FileRepository interface {
	Create(ctx context.Context, file *models.File) error
	GetByID(ctx context.Context, id string) (*models.File, error)
	Delete(ctx context.Context, id, userID string) error
	CheckAccess(ctx context.Context, fileID, userID string) (bool, string, string, error)
}

type BlobStorage interface {
	BucketExists(ctx context.Context, bucketName string) (bool, error)
	MakeBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error
	StatObject(ctx context.Context, bucketName, objectName string, opts minio.StatObjectOptions) (minio.ObjectInfo, error)
	RemoveObject(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error
}

type PresignedURLGenerator interface {
	PresignedPutObject(ctx context.Context, bucketName, objectName string, expires time.Duration) (*url.URL, error)
	PresignedGetObject(ctx context.Context, bucketName, objectName string, expires time.Duration, reqParams url.Values) (*url.URL, error)
}

type fileService struct {
	fileRepo        FileRepository
	storage         BlobStorage
	presignedClient PresignedURLGenerator
	config          *configs.Config
}

func NewFileService(fileRepo FileRepository, storage BlobStorage, presignedClient PresignedURLGenerator, config *configs.Config) *fileService {
	return &fileService{
		fileRepo:        fileRepo,
		storage:         storage,
		presignedClient: presignedClient,
		config:          config,
	}
}

func NewFileServiceWithMinio(fileRepo FileRepository, config *configs.Config) (*fileService, error) {
	minioClient, err := minio.New(config.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.MinIO.AccessKeyID, config.MinIO.SecretAccessKey, ""),
		Secure: config.MinIO.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, config.MinIO.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}
	if !exists {
		err = minioClient.MakeBucket(ctx, config.MinIO.BucketName, minio.MakeBucketOptions{Region: config.MinIO.Region})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	presignedEndpoint := config.MinIO.PublicEndpoint
	if presignedEndpoint == "" {
		presignedEndpoint = config.MinIO.Endpoint
	}
	presignedClient, err := minio.New(presignedEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.MinIO.AccessKeyID, config.MinIO.SecretAccessKey, ""),
		Secure: config.MinIO.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create presigned minio client: %w", err)
	}

	storage := NewMinIOAdapter(minioClient)
	presigned := NewMinIOAdapter(presignedClient)

	return NewFileService(fileRepo, storage, presigned, config), nil
}

func (s *fileService) InitiateUpload(ctx context.Context, input *InitiateUploadInput) (*InitiateUploadOutput, error) {
	if err := utils.ValidatePath(input.Path); err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	uniqueFilename := generateUniqueFilename(input.Filename)
	storagePath := fmt.Sprintf("%s/%s/%s", input.UserID, time.Now().Format("2006/01/02"), uniqueFilename)
	file := models.NewFile(
		input.UserID,
		uniqueFilename,
		input.Filename,
		input.Path,
		input.MimeType,
		storagePath,
		s.config.MinIO.BucketName,
		input.Size,
		input.IsPublic,
		input.Tags,
	)

	if err := s.fileRepo.Create(ctx, file); err != nil {
		return nil, fmt.Errorf("failed to create metadata: %w", err)
	}

	presignedURL, err := s.presignedClient.PresignedPutObject(ctx, s.config.MinIO.BucketName, storagePath, 15*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to generate upload URL: %w", err)
	}

	return &InitiateUploadOutput{
		FileID:       file.ID,
		UploadURL:    presignedURL.String(),
		UploadMethod: "PUT",
		Headers:      map[string]string{},
		ExpiresIn:    int64(15 * time.Minute / time.Second),
	}, nil
}

func (s *fileService) CompleteUpload(ctx context.Context, input *CompleteUploadInput) (*CompleteUploadOutput, error) {
	file, err := s.fileRepo.GetByID(ctx, input.FileID)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}
	if file.UserID != input.UserID {
		return nil, fmt.Errorf("access denied")
	}
	_, err = s.storage.StatObject(ctx, s.config.MinIO.BucketName, file.StoragePath, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("file not found in storage: %w", err)
	}
	return &CompleteUploadOutput{
		StoragePath: file.StoragePath,
		CreatedAt:   file.CreatedAt,
	}, nil
}

func (s *fileService) GetDownloadLink(ctx context.Context, input *GetDownloadLinkInput) (*GetDownloadLinkOutput, error) {
	hasAccess, storagePath, bucket, err := s.fileRepo.CheckAccess(ctx, input.FileID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %w", err)
	}
	if !hasAccess {
		return nil, fmt.Errorf("access denied")
	}
	expires := time.Hour
	if input.ExpiresIn > 0 {
		expires = time.Duration(input.ExpiresIn) * time.Second
	}
	presignedURL, err := s.presignedClient.PresignedGetObject(ctx, bucket, storagePath, expires, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate download URL: %w", err)
	}
	return &GetDownloadLinkOutput{
		DownloadURL: presignedURL.String(),
		Method:      "GET",
		Headers:     map[string]string{},
		ExpiresIn:   int64(expires / time.Second),
	}, nil
}

func (s *fileService) DeleteFile(ctx context.Context, input *DeleteFileInput) (*DeleteFileOutput, error) {
	file, err := s.fileRepo.GetByID(ctx, input.FileID)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}
	if file.UserID != input.UserID {
		return nil, fmt.Errorf("access denied")
	}
	err = s.storage.RemoveObject(ctx, file.Bucket, file.StoragePath, minio.RemoveObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to delete from storage: %w", err)
	}
	if err := s.fileRepo.Delete(ctx, input.FileID, input.UserID); err != nil {
		return nil, fmt.Errorf("failed to delete metadata: %w", err)
	}
	return &DeleteFileOutput{Success: true}, nil
}

func (s *fileService) GetFileInfo(ctx context.Context, input *GetFileInfoInput) (*GetFileInfoOutput, error) {
	hasAccess, _, _, err := s.fileRepo.CheckAccess(ctx, input.FileID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %w", err)
	}
	if !hasAccess {
		return nil, fmt.Errorf("access denied")
	}
	file, err := s.fileRepo.GetByID(ctx, input.FileID)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}
	return &GetFileInfoOutput{File: file}, nil
}

func generateUniqueFilename(original string) string {
	ext := ""
	if idx := len(original) - 1; idx > 0 {
		for i := len(original) - 1; i >= 0; i-- {
			if original[i] == '.' {
				ext = original[i:]
				original = original[:i]
				break
			}
		}
	}
	return fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
}
