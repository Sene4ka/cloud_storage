package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Sene4ka/cloud_storage/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MockMetadataClient struct {
	mock.Mock
}

func (m *MockMetadataClient) GetMetadata(ctx context.Context, in *api.GetMetadataRequest, opts ...grpc.CallOption) (*api.GetMetadataResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.GetMetadataResponse), args.Error(1)
}

func (m *MockMetadataClient) ListMetadata(ctx context.Context, in *api.ListMetadataRequest, opts ...grpc.CallOption) (*api.ListMetadataResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.ListMetadataResponse), args.Error(1)
}

func (m *MockMetadataClient) UpdateMetadata(ctx context.Context, in *api.UpdateMetadataRequest, opts ...grpc.CallOption) (*api.UpdateMetadataResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.UpdateMetadataResponse), args.Error(1)
}

func (m *MockMetadataClient) CheckAccess(ctx context.Context, in *api.CheckAccessRequest, opts ...grpc.CallOption) (*api.CheckAccessResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.CheckAccessResponse), args.Error(1)
}

func (m *MockMetadataClient) TrashFile(ctx context.Context, in *api.TrashFileRequest, opts ...grpc.CallOption) (*api.TrashFileResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.TrashFileResponse), args.Error(1)
}

func (m *MockMetadataClient) RestoreFile(ctx context.Context, in *api.RestoreFileRequest, opts ...grpc.CallOption) (*api.RestoreFileResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.RestoreFileResponse), args.Error(1)
}

type MockFileClient struct {
	mock.Mock
}

func (m *MockFileClient) InitiateUpload(ctx context.Context, in *api.InitiateUploadRequest, opts ...grpc.CallOption) (*api.InitiateUploadResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.InitiateUploadResponse), args.Error(1)
}

func (m *MockFileClient) CompleteUpload(ctx context.Context, in *api.CompleteUploadRequest, opts ...grpc.CallOption) (*api.CompleteUploadResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.CompleteUploadResponse), args.Error(1)
}

func (m *MockFileClient) GetDownloadLink(ctx context.Context, in *api.GetDownloadLinkRequest, opts ...grpc.CallOption) (*api.GetDownloadLinkResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.GetDownloadLinkResponse), args.Error(1)
}

func (m *MockFileClient) DeleteFile(ctx context.Context, in *api.DeleteFileRequest, opts ...grpc.CallOption) (*api.DeleteFileResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.DeleteFileResponse), args.Error(1)
}

func TestFileHandler_HandleFiles_Success(t *testing.T) {
	t.Parallel()

	mockMetadata := new(MockMetadataClient)
	mockFile := new(MockFileClient)
	handler := NewFileHandler(mockMetadata, mockFile)

	mockMetadata.On("ListMetadata", mock.Anything, mock.Anything).Return(&api.ListMetadataResponse{
		Items: []*api.FileMetadata{
			{Id: "file-1", Filename: "test1.txt", UserId: "user-123"},
			{Id: "file-2", Filename: "test2.txt", UserId: "user-123"},
		},
		Total:    2,
		Page:     1,
		PageSize: 20,
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/files?page=1&page_size=20", nil)
	req = ContextWithUser(req, "user-123")
	rr := httptest.NewRecorder()

	handler.HandleFiles(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "file-1")
	mockMetadata.AssertExpectations(t)
}

func TestFileHandler_HandleFiles_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	mockMetadata := new(MockMetadataClient)
	mockFile := new(MockFileClient)
	handler := NewFileHandler(mockMetadata, mockFile)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/files", nil)
	req = ContextWithUser(req, "user-123")
	rr := httptest.NewRecorder()

	handler.HandleFiles(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestFileHandler_HandleFiles_WithFilters(t *testing.T) {
	t.Parallel()

	mockMetadata := new(MockMetadataClient)
	mockFile := new(MockFileClient)
	handler := NewFileHandler(mockMetadata, mockFile)

	mockMetadata.On("ListMetadata", mock.Anything, mock.Anything).Return(&api.ListMetadataResponse{
		Items: []*api.FileMetadata{},
		Total: 0,
		Page:  1,
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/files?page=1&page_size=10&is_trashed=true&sort_by=created_at", nil)
	req = ContextWithUser(req, "user-123")
	rr := httptest.NewRecorder()

	handler.HandleFiles(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockMetadata.AssertExpectations(t)
}

func TestFileHandler_HandleFileDetail_Get_Success(t *testing.T) {
	t.Parallel()

	mockMetadata := new(MockMetadataClient)
	mockFile := new(MockFileClient)
	handler := NewFileHandler(mockMetadata, mockFile)

	mockMetadata.On("GetMetadata", mock.Anything, mock.Anything).Return(&api.GetMetadataResponse{
		Metadata: &api.FileMetadata{
			Id:           "file-123",
			Filename:     "test.txt",
			UserId:       "user-123",
			OriginalName: "test.txt",
			Size:         1024,
			MimeType:     "text/plain",
			IsPublic:     false,
			IsTrashed:    false,
			CreatedAt:    timestamppb.New(time.Now()),
			UpdatedAt:    timestamppb.New(time.Now()),
		},
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/file-123", nil)
	req = ContextWithUser(req, "user-123")
	rr := httptest.NewRecorder()

	handler.HandleFileDetail(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "file-123")
	assert.Contains(t, rr.Body.String(), "test.txt")
	mockMetadata.AssertExpectations(t)
}

func TestFileHandler_HandleFileDetail_Get_NotFound(t *testing.T) {
	t.Parallel()

	mockMetadata := new(MockMetadataClient)
	mockFile := new(MockFileClient)
	handler := NewFileHandler(mockMetadata, mockFile)

	mockMetadata.On("GetMetadata", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/file-123", nil)
	req = ContextWithUser(req, "user-123")
	rr := httptest.NewRecorder()

	handler.HandleFileDetail(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockMetadata.AssertExpectations(t)
}

func TestFileHandler_HandleFileDetail_Update_Success(t *testing.T) {
	t.Parallel()

	mockMetadata := new(MockMetadataClient)
	mockFile := new(MockFileClient)
	handler := NewFileHandler(mockMetadata, mockFile)

	mockMetadata.On("UpdateMetadata", mock.Anything, mock.Anything).Return(&api.UpdateMetadataResponse{
		Metadata: &api.FileMetadata{
			Id:       "file-123",
			Filename: "updated.txt",
			UserId:   "user-123",
			IsPublic: true,
		},
	}, nil)

	req := NewTestRequest(http.MethodPut, "/api/v1/files/file-123", map[string]interface{}{
		"filename":  "updated.txt",
		"is_public": true,
		"tags":      map[string]string{"version": "2"},
	})
	req = ContextWithUser(req, "user-123")
	rr := httptest.NewRecorder()

	handler.HandleFileDetail(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "updated.txt")
	mockMetadata.AssertExpectations(t)
}

func TestFileHandler_HandleFileDetail_Update_InvalidBody(t *testing.T) {
	t.Parallel()

	mockMetadata := new(MockMetadataClient)
	mockFile := new(MockFileClient)
	handler := NewFileHandler(mockMetadata, mockFile)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/files/file-123", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	req = ContextWithUser(req, "user-123")
	rr := httptest.NewRecorder()

	handler.HandleFileDetail(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid request body")
}

func TestFileHandler_HandleFileDetail_Delete_Success(t *testing.T) {
	t.Parallel()

	mockMetadata := new(MockMetadataClient)
	mockFile := new(MockFileClient)
	handler := NewFileHandler(mockMetadata, mockFile)

	mockFile.On("DeleteFile", mock.Anything, mock.Anything).Return(&api.DeleteFileResponse{
		Success: true,
	}, nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/files/file-123", nil)
	req = ContextWithUser(req, "user-123")
	rr := httptest.NewRecorder()

	handler.HandleFileDetail(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "true")
	mockFile.AssertExpectations(t)
}

func TestFileHandler_HandleFileDetail_Delete_Error(t *testing.T) {
	t.Parallel()

	mockMetadata := new(MockMetadataClient)
	mockFile := new(MockFileClient)
	handler := NewFileHandler(mockMetadata, mockFile)

	mockFile.On("DeleteFile", mock.Anything, mock.Anything).Return(nil, errors.New("delete failed"))

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/files/file-123", nil)
	req = ContextWithUser(req, "user-123")
	rr := httptest.NewRecorder()

	handler.HandleFileDetail(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockFile.AssertExpectations(t)
}

func TestFileHandler_HandleFileDetail_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	mockMetadata := new(MockMetadataClient)
	mockFile := new(MockFileClient)
	handler := NewFileHandler(mockMetadata, mockFile)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/files/file-123", nil)
	req = ContextWithUser(req, "user-123")
	rr := httptest.NewRecorder()

	handler.HandleFileDetail(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestFileHandler_HandleInitiateUpload_Success(t *testing.T) {
	t.Parallel()

	mockMetadata := new(MockMetadataClient)
	mockFile := new(MockFileClient)
	handler := NewFileHandler(mockMetadata, mockFile)

	mockFile.On("InitiateUpload", mock.Anything, mock.Anything).Return(&api.InitiateUploadResponse{
		FileId:       "file-123",
		UploadUrl:    "https://storage.example.com/upload/file-123",
		UploadMethod: "PUT",
		Headers:      map[string]string{},
		ExpiresIn:    900,
		Success:      true,
	}, nil)

	req := NewTestRequest(http.MethodPost, "/api/v1/files/upload", map[string]interface{}{
		"filename":  "test.txt",
		"path":      "/files",
		"mime_type": "text/plain",
		"size":      1024,
		"is_public": false,
		"tags":      map[string]string{"type": "document"},
	})
	req = ContextWithUser(req, "user-123")
	rr := httptest.NewRecorder()

	handler.HandleInitiateUpload(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Equal(t, "https://storage.example.com/upload/file-123", rr.Header().Get("Location"))
	assert.Contains(t, rr.Body.String(), "file-123")
	mockFile.AssertExpectations(t)
}

func TestFileHandler_HandleInitiateUpload_InvalidMethod(t *testing.T) {
	t.Parallel()

	mockMetadata := new(MockMetadataClient)
	mockFile := new(MockFileClient)
	handler := NewFileHandler(mockMetadata, mockFile)

	req := NewTestRequest(http.MethodGet, "/api/v1/files/upload", nil)
	req = ContextWithUser(req, "user-123")
	rr := httptest.NewRecorder()

	handler.HandleInitiateUpload(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestFileHandler_HandleInitiateUpload_InvalidBody(t *testing.T) {
	t.Parallel()

	mockMetadata := new(MockMetadataClient)
	mockFile := new(MockFileClient)
	handler := NewFileHandler(mockMetadata, mockFile)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/upload", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	req = ContextWithUser(req, "user-123")
	rr := httptest.NewRecorder()

	handler.HandleInitiateUpload(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid request body")
}

func TestFileHandler_HandleCompleteUpload_Success(t *testing.T) {
	t.Parallel()

	mockMetadata := new(MockMetadataClient)
	mockFile := new(MockFileClient)
	handler := NewFileHandler(mockMetadata, mockFile)

	mockFile.On("CompleteUpload", mock.Anything, mock.Anything).Return(&api.CompleteUploadResponse{
		Success:     true,
		StoragePath: "objects/user-123/file-123",
		CreatedAt:   timestamppb.New(time.Now()),
	}, nil)

	req := NewTestRequest(http.MethodPost, "/api/v1/files/upload/complete", map[string]interface{}{
		"file_id": "file-123",
		"etag":    "abc123",
	})
	req = ContextWithUser(req, "user-123")
	rr := httptest.NewRecorder()

	handler.HandleCompleteUpload(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "true")
	mockFile.AssertExpectations(t)
}

func TestFileHandler_HandleCompleteUpload_InvalidMethod(t *testing.T) {
	t.Parallel()

	mockMetadata := new(MockMetadataClient)
	mockFile := new(MockFileClient)
	handler := NewFileHandler(mockMetadata, mockFile)

	req := NewTestRequest(http.MethodGet, "/api/v1/files/upload/complete", nil)
	req = ContextWithUser(req, "user-123")
	rr := httptest.NewRecorder()

	handler.HandleCompleteUpload(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestFileHandler_HandleDownloadLink_Success(t *testing.T) {
	t.Parallel()

	mockMetadata := new(MockMetadataClient)
	mockFile := new(MockFileClient)
	handler := NewFileHandler(mockMetadata, mockFile)

	mockFile.On("GetDownloadLink", mock.Anything, mock.Anything).Return(&api.GetDownloadLinkResponse{
		DownloadUrl: "https://storage.example.com/download/file-123?token=abc",
		Method:      "GET",
		Headers:     map[string]string{},
		ExpiresIn:   3600,
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/download/file-123", nil)
	req = ContextWithUser(req, "user-123")
	rr := httptest.NewRecorder()

	handler.HandleDownloadLink(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "download")
	mockFile.AssertExpectations(t)
}

func TestFileHandler_HandleDownloadLink_WithExpires(t *testing.T) {
	t.Parallel()

	mockMetadata := new(MockMetadataClient)
	mockFile := new(MockFileClient)
	handler := NewFileHandler(mockMetadata, mockFile)

	mockFile.On("GetDownloadLink", mock.Anything, mock.Anything).Return(&api.GetDownloadLinkResponse{
		DownloadUrl: "https://storage.example.com/download/file-123",
		Method:      "GET",
		ExpiresIn:   7200,
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/download/file-123?expires_in=7200", nil)
	req = ContextWithUser(req, "user-123")
	rr := httptest.NewRecorder()

	handler.HandleDownloadLink(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockFile.AssertExpectations(t)
}

func TestFileHandler_HandleDownloadLink_InvalidMethod(t *testing.T) {
	t.Parallel()

	mockMetadata := new(MockMetadataClient)
	mockFile := new(MockFileClient)
	handler := NewFileHandler(mockMetadata, mockFile)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/download/file-123", nil)
	req = ContextWithUser(req, "user-123")
	rr := httptest.NewRecorder()

	handler.HandleDownloadLink(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestFileHandler_HandleTrashFile_Success(t *testing.T) {
	t.Parallel()

	mockMetadata := new(MockMetadataClient)
	mockFile := new(MockFileClient)
	handler := NewFileHandler(mockMetadata, mockFile)

	mockMetadata.On("TrashFile", mock.Anything, mock.Anything).Return(&api.TrashFileResponse{
		Success: true,
	}, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/trash/file-123", nil)
	req = ContextWithUser(req, "user-123")
	rr := httptest.NewRecorder()

	handler.HandleTrashFile(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "true")
	mockMetadata.AssertExpectations(t)
}

func TestFileHandler_HandleTrashFile_InvalidMethod(t *testing.T) {
	t.Parallel()

	mockMetadata := new(MockMetadataClient)
	mockFile := new(MockFileClient)
	handler := NewFileHandler(mockMetadata, mockFile)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/trash/file-123", nil)
	req = ContextWithUser(req, "user-123")
	rr := httptest.NewRecorder()

	handler.HandleTrashFile(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestFileHandler_HandleRestoreFile_Success(t *testing.T) {
	t.Parallel()

	mockMetadata := new(MockMetadataClient)
	mockFile := new(MockFileClient)
	handler := NewFileHandler(mockMetadata, mockFile)

	mockMetadata.On("RestoreFile", mock.Anything, mock.Anything).Return(&api.RestoreFileResponse{
		Success: true,
	}, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/restore/file-123", nil)
	req = ContextWithUser(req, "user-123")
	rr := httptest.NewRecorder()

	handler.HandleRestoreFile(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "true")
	mockMetadata.AssertExpectations(t)
}

func TestFileHandler_HandleRestoreFile_InvalidMethod(t *testing.T) {
	t.Parallel()

	mockMetadata := new(MockMetadataClient)
	mockFile := new(MockFileClient)
	handler := NewFileHandler(mockMetadata, mockFile)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/restore/file-123", nil)
	req = ContextWithUser(req, "user-123")
	rr := httptest.NewRecorder()

	handler.HandleRestoreFile(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}
