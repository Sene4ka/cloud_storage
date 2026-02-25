package metadata

import (
	"context"
	"fmt"
	"time"

	"github.com/Sene4ka/cloud_storage/internal/api"
	"github.com/Sene4ka/cloud_storage/internal/models"
	"github.com/Sene4ka/cloud_storage/internal/repositories"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	api.UnimplementedMetadataServiceServer
	fileRepo *repositories.FileRepository
}

func NewServer(fileRepo *repositories.FileRepository) *Server {
	return &Server{fileRepo: fileRepo}
}

func (s *Server) CreateMetadata(ctx context.Context, req *api.CreateMetadataRequest) (*api.CreateMetadataResponse, error) {
	file := models.NewFile(
		req.UserId,
		req.Filename,
		req.OriginalName,
		req.MimeType,
		req.StoragePath,
		req.Bucket,
		req.Size,
		req.IsPublic,
		req.Tags,
	)

	if err := s.fileRepo.Create(ctx, file); err != nil {
		return nil, fmt.Errorf("failed to create metadata: %w", err)
	}

	return &api.CreateMetadataResponse{
		Metadata: convertToProto(file),
	}, nil
}

func (s *Server) GetMetadata(ctx context.Context, req *api.GetMetadataRequest) (*api.GetMetadataResponse, error) {
	file, err := s.fileRepo.GetByID(ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	if file.UserID != req.UserId && !file.IsPublic {
		return nil, fmt.Errorf("access denied")
	}

	return &api.GetMetadataResponse{
		Metadata: convertToProto(file),
	}, nil
}

func (s *Server) ListMetadata(ctx context.Context, req *api.ListMetadataRequest) (*api.ListMetadataResponse, error) {
	files, total, err := s.fileRepo.ListByUserID(
		ctx,
		req.UserId,
		int(req.Page),
		int(req.PageSize),
		req.SortBy,
		req.SortOrder,
		req.Search,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list metadata: %w", err)
	}

	protoFiles := make([]*api.FileMetadata, len(files))
	for i, file := range files {
		protoFiles[i] = convertToProto(file)
	}

	return &api.ListMetadataResponse{
		Items:    protoFiles,
		Total:    int32(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (s *Server) UpdateMetadata(ctx context.Context, req *api.UpdateMetadataRequest) (*api.UpdateMetadataResponse, error) {
	existing, err := s.fileRepo.GetByID(ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	if existing.UserID != req.UserId {
		return nil, fmt.Errorf("access denied")
	}

	existing.Filename = req.Filename
	existing.OriginalName = req.OriginalName
	existing.IsPublic = req.IsPublic
	existing.Tags = req.Tags
	existing.UpdatedAt = time.Now()
	if err := s.fileRepo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to update metadata: %w", err)
	}

	return &api.UpdateMetadataResponse{
		Metadata: convertToProto(existing),
	}, nil
}

func (s *Server) DeleteMetadata(ctx context.Context, req *api.DeleteMetadataRequest) (*api.DeleteMetadataResponse, error) {
	if err := s.fileRepo.Delete(ctx, req.Id, req.UserId); err != nil {
		return nil, fmt.Errorf("failed to delete metadata: %w", err)
	}
	return &api.DeleteMetadataResponse{Success: true}, nil
}

func (s *Server) CheckAccess(ctx context.Context, req *api.CheckAccessRequest) (*api.CheckAccessResponse, error) {
	hasAccess, storagePath, bucket, err := s.fileRepo.CheckAccess(ctx, req.FileId, req.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	return &api.CheckAccessResponse{
		HasAccess:   hasAccess,
		StoragePath: storagePath,
		Bucket:      bucket,
	}, nil
}

func convertToProto(file *models.File) *api.FileMetadata {
	return &api.FileMetadata{
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
	}
}
