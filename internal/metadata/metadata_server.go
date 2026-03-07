package metadata

import (
	"context"

	"github.com/Sene4ka/cloud_storage/internal/api"
	"github.com/Sene4ka/cloud_storage/internal/models"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MetadataService interface {
	GetMetadata(ctx context.Context, input *GetMetadataInput) (*GetMetadataOutput, error)
	ListMetadata(ctx context.Context, input *ListMetadataInput) (*ListMetadataOutput, error)
	UpdateMetadata(ctx context.Context, input *UpdateMetadataInput) (*UpdateMetadataOutput, error)
	CheckAccess(ctx context.Context, input *CheckAccessInput) (*CheckAccessOutput, error)
	TrashFile(ctx context.Context, input *TrashFileInput) (*TrashFileOutput, error)
	RestoreFile(ctx context.Context, input *RestoreFileInput) (*RestoreFileOutput, error)
}

type Server struct {
	api.UnimplementedMetadataServiceServer
	service MetadataService
}

func NewServer(service MetadataService) *Server {
	return &Server{service: service}
}

func (s *Server) GetMetadata(ctx context.Context, req *api.GetMetadataRequest) (*api.GetMetadataResponse, error) {
	out, err := s.service.GetMetadata(ctx, &GetMetadataInput{
		FileID: req.Id,
		UserID: req.UserId,
	})
	if err != nil {
		return nil, err
	}
	return &api.GetMetadataResponse{
		Metadata: convertToProto(out.File),
	}, nil
}

func (s *Server) ListMetadata(ctx context.Context, req *api.ListMetadataRequest) (*api.ListMetadataResponse, error) {
	var isThrashed *bool
	if req.IsTrashed != nil {
		val := req.IsTrashed.Value
		isThrashed = &val
	}

	out, err := s.service.ListMetadata(ctx, &ListMetadataInput{
		UserID:    req.UserId,
		Page:      int(req.Page),
		PageSize:  int(req.PageSize),
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
		Search:    req.Search,
		IsTrashed: isThrashed,
	})
	if err != nil {
		return nil, err
	}

	protoItems := make([]*api.FileMetadata, len(out.Items))
	for i, file := range out.Items {
		protoItems[i] = convertToProto(file)
	}

	return &api.ListMetadataResponse{
		Items:    protoItems,
		Total:    int32(out.Total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (s *Server) UpdateMetadata(ctx context.Context, req *api.UpdateMetadataRequest) (*api.UpdateMetadataResponse, error) {
	out, err := s.service.UpdateMetadata(ctx, &UpdateMetadataInput{
		FileID:       req.Id,
		UserID:       req.UserId,
		Filename:     req.Filename,
		OriginalName: req.OriginalName,
		IsPublic:     req.IsPublic,
		Tags:         req.Tags,
	})
	if err != nil {
		return nil, err
	}
	return &api.UpdateMetadataResponse{
		Metadata: convertToProto(out.File),
	}, nil
}

func (s *Server) CheckAccess(ctx context.Context, req *api.CheckAccessRequest) (*api.CheckAccessResponse, error) {
	out, err := s.service.CheckAccess(ctx, &CheckAccessInput{
		FileID: req.FileId,
		UserID: req.UserId,
	})
	if err != nil {
		return nil, err
	}
	return &api.CheckAccessResponse{
		HasAccess:   out.HasAccess,
		StoragePath: out.StoragePath,
		Bucket:      out.Bucket,
	}, nil
}

func (s *Server) TrashFile(ctx context.Context, req *api.TrashFileRequest) (*api.TrashFileResponse, error) {
	out, err := s.service.TrashFile(ctx, &TrashFileInput{
		FileID: req.FileId,
		UserID: req.UserId,
	})

	if err != nil {
		return nil, err
	}

	return &api.TrashFileResponse{Success: out.Success}, nil
}

func (s *Server) RestoreFile(ctx context.Context, req *api.RestoreFileRequest) (*api.RestoreFileResponse, error) {
	out, err := s.service.RestoreFile(ctx, &RestoreFileInput{
		FileID: req.FileId,
		UserID: req.UserId,
	})

	if err != nil {
		return nil, err
	}

	return &api.RestoreFileResponse{Success: out.Success}, nil
}

func convertToProto(file *models.File) *api.FileMetadata {
	var thrashedAt *timestamppb.Timestamp
	if file.TrashedAt != nil {
		thrashedAt = timestamppb.New(*file.TrashedAt)
	}

	return &api.FileMetadata{
		Id:           file.ID,
		UserId:       file.UserID,
		Filename:     file.Filename,
		OriginalName: file.OriginalName,
		Path:         file.Path,
		Size:         file.Size,
		MimeType:     file.MimeType,
		StoragePath:  file.StoragePath,
		Bucket:       file.Bucket,
		CreatedAt:    timestamppb.New(file.CreatedAt),
		UpdatedAt:    timestamppb.New(file.UpdatedAt),
		IsPublic:     file.IsPublic,
		Tags:         file.Tags,
		IsTrashed:    file.IsTrashed,
		TrashedAt:    thrashedAt,
	}
}
