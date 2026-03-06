package metadata

import (
	"context"
	"fmt"
	"time"

	"github.com/Sene4ka/cloud_storage/internal/models"
)

type FileRepository interface {
	Create(ctx context.Context, file *models.File) error
	GetByID(ctx context.Context, id string) (*models.File, error)
	ListByUserID(ctx context.Context, userID string, page, pageSize int, sortBy, sortOrder, search string) ([]*models.File, int, error)
	Update(ctx context.Context, file *models.File) error
	Delete(ctx context.Context, id, userID string) error
	CheckAccess(ctx context.Context, fileID, userID string) (bool, string, string, error)
}

type metadataService struct {
	fileRepo FileRepository
}

func NewMetadataService(fileRepo FileRepository) *metadataService {
	return &metadataService{fileRepo: fileRepo}
}

func (s *metadataService) CreateMetadata(ctx context.Context, input *CreateMetadataInput) (*CreateMetadataOutput, error) {
	file := models.NewFile(
		input.UserID,
		input.Filename,
		input.OriginalName,
		input.MimeType,
		input.StoragePath,
		input.Bucket,
		input.Size,
		input.IsPublic,
		input.Tags,
	)

	if err := s.fileRepo.Create(ctx, file); err != nil {
		return nil, fmt.Errorf("failed to create metadata: %w", err)
	}

	return &CreateMetadataOutput{File: file}, nil
}

func (s *metadataService) GetMetadata(ctx context.Context, input *GetMetadataInput) (*GetMetadataOutput, error) {
	file, err := s.fileRepo.GetByID(ctx, input.FileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	if file.UserID != input.UserID && !file.IsPublic {
		return nil, fmt.Errorf("access denied")
	}

	return &GetMetadataOutput{File: file}, nil
}

func (s *metadataService) ListMetadata(ctx context.Context, input *ListMetadataInput) (*ListMetadataOutput, error) {
	files, total, err := s.fileRepo.ListByUserID(
		ctx,
		input.UserID,
		input.Page,
		input.PageSize,
		input.SortBy,
		input.SortOrder,
		input.Search,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list metadata: %w", err)
	}

	return &ListMetadataOutput{
		Items:    files,
		Total:    int64(total),
		Page:     input.Page,
		PageSize: input.PageSize,
	}, nil
}

func (s *metadataService) UpdateMetadata(ctx context.Context, input *UpdateMetadataInput) (*UpdateMetadataOutput, error) {
	existing, err := s.fileRepo.GetByID(ctx, input.FileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	if existing.UserID != input.UserID {
		return nil, fmt.Errorf("access denied")
	}

	existing.Filename = input.Filename
	existing.OriginalName = input.OriginalName
	existing.IsPublic = input.IsPublic
	existing.Tags = input.Tags
	existing.UpdatedAt = time.Now()

	if err := s.fileRepo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to update metadata: %w", err)
	}

	return &UpdateMetadataOutput{File: existing}, nil
}

func (s *metadataService) DeleteMetadata(ctx context.Context, input *DeleteMetadataInput) (*DeleteMetadataOutput, error) {
	if err := s.fileRepo.Delete(ctx, input.FileID, input.UserID); err != nil {
		return nil, fmt.Errorf("failed to delete metadata: %w", err)
	}
	return &DeleteMetadataOutput{Success: true}, nil
}

func (s *metadataService) CheckAccess(ctx context.Context, input *CheckAccessInput) (*CheckAccessOutput, error) {
	hasAccess, storagePath, bucket, err := s.fileRepo.CheckAccess(ctx, input.FileID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	return &CheckAccessOutput{
		HasAccess:   hasAccess,
		StoragePath: storagePath,
		Bucket:      bucket,
	}, nil
}
