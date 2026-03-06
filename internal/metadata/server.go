package metadata

import (
	"context"

	"github.com/Sene4ka/cloud_storage/internal/api"
	"github.com/Sene4ka/cloud_storage/internal/models"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MetadataService interface {
	CreateMetadata(ctx context.Context, input *CreateMetadataInput) (*CreateMetadataOutput, error)
	GetMetadata(ctx context.Context, input *GetMetadataInput) (*GetMetadataOutput, error)
	ListMetadata(ctx context.Context, input *ListMetadataInput) (*ListMetadataOutput, error)
	UpdateMetadata(ctx context.Context, input *UpdateMetadataInput) (*UpdateMetadataOutput, error)
	DeleteMetadata(ctx context.Context, input *DeleteMetadataInput) (*DeleteMetadataOutput, error)
	CheckAccess(ctx context.Context, input *CheckAccessInput) (*CheckAccessOutput, error)
}

type Server struct {
	api.UnimplementedMetadataServiceServer
	service MetadataService
}

func NewServer(service MetadataService) *Server {
	return &Server{service: service}
}

func (s *Server) CreateMetadata(ctx context.Context, req *api.CreateMetadataRequest) (*api.CreateMetadataResponse, error) {
	out, err := s.service.CreateMetadata(ctx, &CreateMetadataInput{
		UserID:       req.UserId,
		Filename:     req.Filename,
		OriginalName: req.OriginalName,
		Size:         req.Size,
		MimeType:     req.MimeType,
		StoragePath:  req.StoragePath,
		Bucket:       req.Bucket,
		IsPublic:     req.IsPublic,
		Tags:         req.Tags,
	})
	if err != nil {
		return nil, err
	}
	return &api.CreateMetadataResponse{
		Metadata: convertToProto(out.File),
	}, nil
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
	out, err := s.service.ListMetadata(ctx, &ListMetadataInput{
		UserID:    req.UserId,
		Page:      int(req.Page),
		PageSize:  int(req.PageSize),
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
		Search:    req.Search,
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

func (s *Server) DeleteMetadata(ctx context.Context, req *api.DeleteMetadataRequest) (*api.DeleteMetadataResponse, error) {
	out, err := s.service.DeleteMetadata(ctx, &DeleteMetadataInput{
		FileID: req.Id,
		UserID: req.UserId,
	})
	if err != nil {
		return nil, err
	}
	return &api.DeleteMetadataResponse{Success: out.Success}, nil
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
