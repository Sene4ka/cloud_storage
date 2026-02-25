package file

import (
	"context"
	"fmt"
	"time"

	"github.com/Sene4ka/cloud_storage/configs"
	"github.com/Sene4ka/cloud_storage/internal/api"
	"github.com/Sene4ka/cloud_storage/internal/models"
	"github.com/Sene4ka/cloud_storage/internal/repositories"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	api.UnimplementedFileServiceServer
	fileRepo        *repositories.FileRepository
	minioClient     *minio.Client
	presignedClient *minio.Client
	config          *configs.Config
}

func NewServer(fileRepo *repositories.FileRepository, config *configs.Config) (*Server, error) {
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

	return &Server{
		fileRepo:        fileRepo,
		minioClient:     minioClient,
		presignedClient: presignedClient,
		config:          config,
	}, nil
}

func (s *Server) InitiateUpload(ctx context.Context, req *api.InitiateUploadRequest) (*api.InitiateUploadResponse, error) {
	uniqueFilename := generateUniqueFilename(req.Filename)
	storagePath := fmt.Sprintf("%s/%s/%s", req.UserId, time.Now().Format("2006/01/02"), uniqueFilename)
	file := models.NewFile(
		req.UserId,
		uniqueFilename,
		req.Filename,
		req.MimeType,
		storagePath,
		s.config.MinIO.BucketName,
		req.Size,
		req.IsPublic,
		req.Tags,
	)

	if err := s.fileRepo.Create(ctx, file); err != nil {
		return nil, fmt.Errorf("failed to create metadata: %w", err)
	}

	presignedURL, err := s.presignedClient.PresignedPutObject(ctx, s.config.MinIO.BucketName, storagePath, 15*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to generate upload URL: %w", err)
	}

	return &api.InitiateUploadResponse{
		FileId:       file.ID,
		UploadUrl:    presignedURL.String(),
		UploadMethod: "PUT",
		Headers:      map[string]string{},
		ExpiresIn:    int64(15 * time.Minute / time.Second),
		Success:      true,
	}, nil

}

func (s *Server) CompleteUpload(ctx context.Context, req *api.CompleteUploadRequest) (*api.CompleteUploadResponse, error) {
	file, err := s.fileRepo.GetByID(ctx, req.FileId)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	if file.UserID != req.UserId {
		return nil, fmt.Errorf("access denied")
	}

	_, err = s.minioClient.StatObject(ctx, s.config.MinIO.BucketName, file.StoragePath, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("file not found in storage: %w", err)
	}

	return &api.CompleteUploadResponse{
		Success:     true,
		StoragePath: file.StoragePath,
		CreatedAt:   timestamppb.New(file.CreatedAt),
	}, nil
}

func (s *Server) GetDownloadLink(ctx context.Context, req *api.GetDownloadLinkRequest) (*api.GetDownloadLinkResponse, error) {
	hasAccess, storagePath, bucket, err := s.fileRepo.CheckAccess(ctx, req.FileId, req.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		return nil, fmt.Errorf("access denied")
	}

	expiresIn := time.Hour
	if req.ExpiresIn > 0 {
		expiresIn = time.Duration(req.ExpiresIn) * time.Second
	}

	presignedURL, err := s.presignedClient.PresignedGetObject(ctx, bucket, storagePath, expiresIn, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate download URL: %w", err)
	}

	return &api.GetDownloadLinkResponse{
		DownloadUrl: presignedURL.String(),
		Method:      "GET",
		Headers:     map[string]string{},
		ExpiresIn:   int64(expiresIn / time.Second),
	}, nil
}

func (s *Server) DeleteFile(ctx context.Context, req *api.DeleteFileRequest) (*api.DeleteFileResponse, error) {
	file, err := s.fileRepo.GetByID(ctx, req.FileId)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	if file.UserID != req.UserId {
		return nil, fmt.Errorf("access denied")
	}

	err = s.minioClient.RemoveObject(ctx, file.Bucket, file.StoragePath, minio.RemoveObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to delete from storage: %w", err)
	}

	if err := s.fileRepo.Delete(ctx, req.FileId, req.UserId); err != nil {
		return nil, fmt.Errorf("failed to delete metadata: %w", err)
	}

	return &api.DeleteFileResponse{Success: true}, nil
}

func (s *Server) GetFileInfo(ctx context.Context, req *api.GetFileInfoRequest) (*api.GetFileInfoResponse, error) {
	hasAccess, _, _, err := s.fileRepo.CheckAccess(ctx, req.FileId, req.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		return nil, fmt.Errorf("access denied")
	}

	file, err := s.fileRepo.GetByID(ctx, req.FileId)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	return &api.GetFileInfoResponse{
		Id:           file.ID,
		UserId:       file.UserID,
		Filename:     file.Filename,
		OriginalName: file.OriginalName,
		Size:         file.Size,
		MimeType:     file.MimeType,
		StoragePath:  file.StoragePath,
		Bucket:       file.Bucket,
		CreatedAt:    timestamppb.New(file.CreatedAt),
		UpdatedAt:    timestamppb.New(file.UpdatedAt),
		IsPublic:     file.IsPublic,
		Tags:         file.Tags,
	}, nil
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
