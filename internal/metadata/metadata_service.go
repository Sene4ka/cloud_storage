package metadata

import (
	"context"
	"fmt"
	"time"

	"github.com/Sene4ka/cloud_storage/internal/metrics"
	"github.com/Sene4ka/cloud_storage/internal/models"
	"github.com/Sene4ka/cloud_storage/internal/utils"
)

type FileRepository interface {
	Create(ctx context.Context, file *models.File) error
	GetByID(ctx context.Context, id string) (*models.File, error)
	ListByUserID(ctx context.Context, userID string, page, pageSize int, sortBy, sortOrder, search string, isTrashed *bool) ([]*models.File, int, error)
	Update(ctx context.Context, file *models.File) error
	CheckAccess(ctx context.Context, fileID, userID string) (bool, string, string, error)
	Delete(ctx context.Context, id, userID string) error
	SetTrashed(ctx context.Context, fileID, userID string, isTrashed bool) error
}

type metadataService struct {
	fileRepo FileRepository
}

func NewMetadataService(fileRepo FileRepository) *metadataService {
	return &metadataService{fileRepo: fileRepo}
}

func (s *metadataService) GetMetadata(ctx context.Context, input *GetMetadataInput) (output *GetMetadataOutput, err error) {
	defer func() {
		status := "success"
		if err != nil {
			status = "error"
		}
		metrics.RecordMetadataOperation("get_metadata", status)
	}()

	file, err := s.fileRepo.GetByID(ctx, input.FileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	if file.UserID != input.UserID && !file.IsPublic {
		return nil, fmt.Errorf("access denied")
	}

	return &GetMetadataOutput{File: file}, nil
}

func (s *metadataService) ListMetadata(ctx context.Context, input *ListMetadataInput) (output *ListMetadataOutput, err error) {
	defer func() {
		status := "success"
		if err != nil {
			status = "error"
		}
		metrics.RecordMetadataOperation("list_metadata", status)
	}()

	files, total, err := s.fileRepo.ListByUserID(
		ctx,
		input.UserID,
		input.Page,
		input.PageSize,
		input.SortBy,
		input.SortOrder,
		input.Search,
		input.IsTrashed,
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

func (s *metadataService) UpdateMetadata(ctx context.Context, input *UpdateMetadataInput) (output *UpdateMetadataOutput, err error) {
	defer func() {
		status := "success"
		if err != nil {
			status = "error"
		}
		metrics.RecordMetadataOperation("update_metadata", status)
	}()

	if err := utils.ValidatePath(input.Path); err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	existing, err := s.fileRepo.GetByID(ctx, input.FileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	if existing.UserID != input.UserID {
		return nil, fmt.Errorf("access denied")
	}

	existing.Filename = input.Filename
	existing.OriginalName = input.OriginalName
	existing.Path = input.Path
	existing.IsPublic = input.IsPublic
	existing.Tags = input.Tags
	existing.UpdatedAt = time.Now()

	if err := s.fileRepo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to update metadata: %w", err)
	}

	return &UpdateMetadataOutput{File: existing}, nil
}

func (s *metadataService) CheckAccess(ctx context.Context, input *CheckAccessInput) (output *CheckAccessOutput, err error) {
	defer func() {
		status := "success"
		if err != nil {
			status = "error"
		}
		metrics.RecordMetadataOperation("check_access", status)
	}()

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

func (s *metadataService) TrashFile(ctx context.Context, input *TrashFileInput) (output *TrashFileOutput, err error) {
	defer func() {
		status := "success"
		if err != nil {
			status = "error"
		}
		metrics.RecordMetadataOperation("trash_file", status)
	}()

	if err := s.fileRepo.SetTrashed(ctx, input.FileID, input.UserID, true); err != nil {
		return nil, err
	}
	return &TrashFileOutput{Success: true}, nil
}

func (s *metadataService) RestoreFile(ctx context.Context, input *RestoreFileInput) (output *RestoreFileOutput, err error) {
	defer func() {
		status := "success"
		if err != nil {
			status = "error"
		}
		metrics.RecordMetadataOperation("restore_file", status)
	}()

	if err := s.fileRepo.SetTrashed(ctx, input.FileID, input.UserID, false); err != nil {
		return nil, err
	}
	return &RestoreFileOutput{Success: true}, nil
}
