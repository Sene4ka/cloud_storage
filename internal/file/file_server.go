package file

import (
	"context"

	"github.com/Sene4ka/cloud_storage/internal/api"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type FileService interface {
	InitiateUpload(ctx context.Context, input *InitiateUploadInput) (*InitiateUploadOutput, error)
	CompleteUpload(ctx context.Context, input *CompleteUploadInput) (*CompleteUploadOutput, error)
	GetDownloadLink(ctx context.Context, input *GetDownloadLinkInput) (*GetDownloadLinkOutput, error)
	DeleteFile(ctx context.Context, input *DeleteFileInput) (*DeleteFileOutput, error)
	GetFileInfo(ctx context.Context, input *GetFileInfoInput) (*GetFileInfoOutput, error)
}

type Server struct {
	api.UnimplementedFileServiceServer
	service FileService
}

func NewServer(service FileService) *Server {
	return &Server{service: service}
}

func (s *Server) InitiateUpload(ctx context.Context, req *api.InitiateUploadRequest) (*api.InitiateUploadResponse, error) {
	out, err := s.service.InitiateUpload(ctx, &InitiateUploadInput{
		UserID:   req.UserId,
		Filename: req.Filename,
		Path:     req.Path,
		MimeType: req.MimeType,
		Size:     req.Size,
		IsPublic: req.IsPublic,
		Tags:     req.Tags,
	})
	if err != nil {
		return nil, err
	}
	return &api.InitiateUploadResponse{
		FileId:       out.FileID,
		UploadUrl:    out.UploadURL,
		UploadMethod: out.UploadMethod,
		Headers:      out.Headers,
		ExpiresIn:    out.ExpiresIn,
		Success:      true,
	}, nil
}

func (s *Server) CompleteUpload(ctx context.Context, req *api.CompleteUploadRequest) (*api.CompleteUploadResponse, error) {
	out, err := s.service.CompleteUpload(ctx, &CompleteUploadInput{
		FileID: req.FileId,
		UserID: req.UserId,
		ETag:   req.Etag,
	})
	if err != nil {
		return nil, err
	}
	return &api.CompleteUploadResponse{
		Success:     true,
		StoragePath: out.StoragePath,
		CreatedAt:   timestamppb.New(out.CreatedAt),
	}, nil
}

func (s *Server) GetDownloadLink(ctx context.Context, req *api.GetDownloadLinkRequest) (*api.GetDownloadLinkResponse, error) {
	out, err := s.service.GetDownloadLink(ctx, &GetDownloadLinkInput{
		FileID:    req.FileId,
		UserID:    req.UserId,
		ExpiresIn: req.ExpiresIn,
	})
	if err != nil {
		return nil, err
	}
	return &api.GetDownloadLinkResponse{
		DownloadUrl: out.DownloadURL,
		Method:      out.Method,
		Headers:     out.Headers,
		ExpiresIn:   out.ExpiresIn,
	}, nil
}

func (s *Server) DeleteFile(ctx context.Context, req *api.DeleteFileRequest) (*api.DeleteFileResponse, error) {
	out, err := s.service.DeleteFile(ctx, &DeleteFileInput{
		FileID: req.FileId,
		UserID: req.UserId,
	})
	if err != nil {
		return nil, err
	}
	return &api.DeleteFileResponse{Success: out.Success}, nil
}
