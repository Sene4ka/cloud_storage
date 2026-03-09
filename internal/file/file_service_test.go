package file

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/Sene4ka/cloud_storage/configs"
	"github.com/Sene4ka/cloud_storage/internal/models"
	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockFileRepository struct {
	mock.Mock
}

func (m *MockFileRepository) Create(ctx context.Context, file *models.File) error {
	args := m.Called(ctx, file)
	return args.Error(0)
}

func (m *MockFileRepository) GetByID(ctx context.Context, id string) (*models.File, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.File), args.Error(1)
}

func (m *MockFileRepository) Delete(ctx context.Context, id, userID string) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockFileRepository) CheckAccess(ctx context.Context, fileID, userID string) (bool, string, string, error) {
	args := m.Called(ctx, fileID, userID)
	return args.Bool(0), args.String(1), args.String(2), args.Error(3)
}

type MockBlobStorage struct {
	mock.Mock
}

func (m *MockBlobStorage) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	args := m.Called(ctx, bucketName)
	return args.Bool(0), args.Error(1)
}

func (m *MockBlobStorage) MakeBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error {
	args := m.Called(ctx, bucketName, opts)
	return args.Error(0)
}

func (m *MockBlobStorage) StatObject(ctx context.Context, bucketName, objectName string, opts minio.StatObjectOptions) (minio.ObjectInfo, error) {
	args := m.Called(ctx, bucketName, objectName, opts)
	return args.Get(0).(minio.ObjectInfo), args.Error(1)
}

func (m *MockBlobStorage) RemoveObject(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error {
	args := m.Called(ctx, bucketName, objectName, opts)
	return args.Error(0)
}

type MockPresignedURLGenerator struct {
	mock.Mock
}

func (m *MockPresignedURLGenerator) PresignedPutObject(ctx context.Context, bucketName, objectName string, expires time.Duration) (*url.URL, error) {
	args := m.Called(ctx, bucketName, objectName, expires)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*url.URL), args.Error(1)
}

func (m *MockPresignedURLGenerator) PresignedGetObject(ctx context.Context, bucketName, objectName string, expires time.Duration, reqParams url.Values) (*url.URL, error) {
	args := m.Called(ctx, bucketName, objectName, expires, reqParams)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*url.URL), args.Error(1)
}

func TestFileService_InitiateUpload_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockFileRepository)
	mockStorage := new(MockBlobStorage)
	mockPresigned := new(MockPresignedURLGenerator)
	config := &configs.Config{
		MinIO: configs.MinIOConfig{
			BucketName: "cloud-storage",
		},
	}

	svc := NewFileService(mockRepo, mockStorage, mockPresigned, config)

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(f *models.File) bool {
		return f.UserID == "user-123" && f.Filename != ""
	})).Return(nil)

	presignedURL, _ := url.Parse("https://storage.example.com/upload/file-123")
	mockPresigned.On("PresignedPutObject", mock.Anything, "cloud-storage", mock.Anything, 15*time.Minute).Return(presignedURL, nil)

	input := &InitiateUploadInput{
		UserID:   "user-123",
		Filename: "test.txt",
		Path:     "/files",
		MimeType: "text/plain",
		Size:     1024,
		IsPublic: false,
		Tags:     map[string]string{"type": "document"},
	}

	output, err := svc.InitiateUpload(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.NotEmpty(t, output.FileID)
	assert.Contains(t, output.UploadURL, "storage.example.com")
	assert.Equal(t, "PUT", output.UploadMethod)
	mockRepo.AssertExpectations(t)
	mockPresigned.AssertExpectations(t)
}

func TestFileService_InitiateUpload_InvalidPath(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockFileRepository)
	mockStorage := new(MockBlobStorage)
	mockPresigned := new(MockPresignedURLGenerator)
	config := &configs.Config{
		MinIO: configs.MinIOConfig{
			BucketName: "cloud-storage",
		},
	}

	svc := NewFileService(mockRepo, mockStorage, mockPresigned, config)

	input := &InitiateUploadInput{
		UserID:   "user-123",
		Filename: "test.txt",
		Path:     "../../../etc/passwd",
		MimeType: "text/plain",
		Size:     1024,
	}

	output, err := svc.InitiateUpload(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid path")
	assert.Nil(t, output)
}

func TestFileService_InitiateUpload_RepoError(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockFileRepository)
	mockStorage := new(MockBlobStorage)
	mockPresigned := new(MockPresignedURLGenerator)
	config := &configs.Config{
		MinIO: configs.MinIOConfig{
			BucketName: "cloud-storage",
		},
	}

	svc := NewFileService(mockRepo, mockStorage, mockPresigned, config)

	mockRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

	input := &InitiateUploadInput{
		UserID:   "user-123",
		Filename: "test.txt",
		Path:     "/files",
		MimeType: "text/plain",
		Size:     1024,
	}

	output, err := svc.InitiateUpload(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create metadata")
	assert.Nil(t, output)
	mockRepo.AssertExpectations(t)
}

func TestFileService_CompleteUpload_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockFileRepository)
	mockStorage := new(MockBlobStorage)
	mockPresigned := new(MockPresignedURLGenerator)
	config := &configs.Config{
		MinIO: configs.MinIOConfig{
			BucketName: "cloud-storage",
		},
	}

	svc := NewFileService(mockRepo, mockStorage, mockPresigned, config)

	existingFile := &models.File{
		ID:          "file-123",
		UserID:      "user-123",
		StoragePath: "objects/file-123",
		Bucket:      "cloud-storage",
		CreatedAt:   time.Now(),
	}

	mockRepo.On("GetByID", mock.Anything, "file-123").Return(existingFile, nil)
	mockStorage.On("StatObject", mock.Anything, "cloud-storage", "objects/file-123", mock.Anything).Return(minio.ObjectInfo{}, nil)

	input := &CompleteUploadInput{
		FileID: "file-123",
		UserID: "user-123",
	}

	output, err := svc.CompleteUpload(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "objects/file-123", output.StoragePath)
	mockRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestFileService_CompleteUpload_AccessDenied(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockFileRepository)
	mockStorage := new(MockBlobStorage)
	mockPresigned := new(MockPresignedURLGenerator)
	config := &configs.Config{
		MinIO: configs.MinIOConfig{
			BucketName: "cloud-storage",
		},
	}

	svc := NewFileService(mockRepo, mockStorage, mockPresigned, config)

	otherUserFile := &models.File{
		ID:     "file-123",
		UserID: "other-user",
	}

	mockRepo.On("GetByID", mock.Anything, "file-123").Return(otherUserFile, nil)

	input := &CompleteUploadInput{
		FileID: "file-123",
		UserID: "user-123",
	}

	output, err := svc.CompleteUpload(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access denied")
	assert.Nil(t, output)
	mockRepo.AssertExpectations(t)
}

func TestFileService_CompleteUpload_FileNotFound(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockFileRepository)
	mockStorage := new(MockBlobStorage)
	mockPresigned := new(MockPresignedURLGenerator)
	config := &configs.Config{
		MinIO: configs.MinIOConfig{
			BucketName: "cloud-storage",
		},
	}

	svc := NewFileService(mockRepo, mockStorage, mockPresigned, config)

	mockRepo.On("GetByID", mock.Anything, "file-123").Return(nil, errors.New("not found"))

	input := &CompleteUploadInput{
		FileID: "file-123",
		UserID: "user-123",
	}

	output, err := svc.CompleteUpload(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file not found")
	assert.Nil(t, output)
	mockRepo.AssertExpectations(t)
}

func TestFileService_GetDownloadLink_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockFileRepository)
	mockStorage := new(MockBlobStorage)
	mockPresigned := new(MockPresignedURLGenerator)
	config := &configs.Config{
		MinIO: configs.MinIOConfig{
			BucketName: "cloud-storage",
		},
	}

	svc := NewFileService(mockRepo, mockStorage, mockPresigned, config)

	mockRepo.On("CheckAccess", mock.Anything, "file-123", "user-123").Return(true, "objects/file-123", "cloud-storage", nil)

	presignedURL, _ := url.Parse("https://storage.example.com/download/file-123")
	mockPresigned.On("PresignedGetObject", mock.Anything, "cloud-storage", "objects/file-123", time.Hour, mock.Anything).Return(presignedURL, nil)

	input := &GetDownloadLinkInput{
		FileID: "file-123",
		UserID: "user-123",
	}

	output, err := svc.GetDownloadLink(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Contains(t, output.DownloadURL, "storage.example.com")
	assert.Equal(t, "GET", output.Method)
	mockRepo.AssertExpectations(t)
	mockPresigned.AssertExpectations(t)
}

func TestFileService_GetDownloadLink_AccessDenied(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockFileRepository)
	mockStorage := new(MockBlobStorage)
	mockPresigned := new(MockPresignedURLGenerator)
	config := &configs.Config{
		MinIO: configs.MinIOConfig{
			BucketName: "cloud-storage",
		},
	}

	svc := NewFileService(mockRepo, mockStorage, mockPresigned, config)

	mockRepo.On("CheckAccess", mock.Anything, "file-123", "user-123").Return(false, "", "", nil)

	input := &GetDownloadLinkInput{
		FileID: "file-123",
		UserID: "user-123",
	}

	output, err := svc.GetDownloadLink(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access denied")
	assert.Nil(t, output)
	mockRepo.AssertExpectations(t)
}

func TestFileService_DeleteFile_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockFileRepository)
	mockStorage := new(MockBlobStorage)
	mockPresigned := new(MockPresignedURLGenerator)
	config := &configs.Config{
		MinIO: configs.MinIOConfig{
			BucketName: "cloud-storage",
		},
	}

	svc := NewFileService(mockRepo, mockStorage, mockPresigned, config)

	existingFile := &models.File{
		ID:          "file-123",
		UserID:      "user-123",
		StoragePath: "objects/file-123",
		Bucket:      "cloud-storage",
	}

	mockRepo.On("GetByID", mock.Anything, "file-123").Return(existingFile, nil)
	mockStorage.On("RemoveObject", mock.Anything, "cloud-storage", "objects/file-123", mock.Anything).Return(nil)
	mockRepo.On("Delete", mock.Anything, "file-123", "user-123").Return(nil)

	input := &DeleteFileInput{
		FileID: "file-123",
		UserID: "user-123",
	}

	output, err := svc.DeleteFile(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.True(t, output.Success)
	mockRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestFileService_DeleteFile_AccessDenied(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockFileRepository)
	mockStorage := new(MockBlobStorage)
	mockPresigned := new(MockPresignedURLGenerator)
	config := &configs.Config{
		MinIO: configs.MinIOConfig{
			BucketName: "cloud-storage",
		},
	}

	svc := NewFileService(mockRepo, mockStorage, mockPresigned, config)

	otherUserFile := &models.File{
		ID:     "file-123",
		UserID: "other-user",
	}

	mockRepo.On("GetByID", mock.Anything, "file-123").Return(otherUserFile, nil)

	input := &DeleteFileInput{
		FileID: "file-123",
		UserID: "user-123",
	}

	output, err := svc.DeleteFile(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access denied")
	assert.Nil(t, output)
	mockRepo.AssertExpectations(t)
}

func TestFileService_GetFileInfo_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockFileRepository)
	mockStorage := new(MockBlobStorage)
	mockPresigned := new(MockPresignedURLGenerator)
	config := &configs.Config{
		MinIO: configs.MinIOConfig{
			BucketName: "cloud-storage",
		},
	}

	svc := NewFileService(mockRepo, mockStorage, mockPresigned, config)

	file := &models.File{
		ID:       "file-123",
		UserID:   "user-123",
		Filename: "test.txt",
	}

	mockRepo.On("CheckAccess", mock.Anything, "file-123", "user-123").Return(true, "", "", nil)
	mockRepo.On("GetByID", mock.Anything, "file-123").Return(file, nil)

	input := &GetFileInfoInput{
		FileID: "file-123",
		UserID: "user-123",
	}

	output, err := svc.GetFileInfo(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "test.txt", output.File.Filename)
	mockRepo.AssertExpectations(t)
}

func TestFileService_GetFileInfo_AccessDenied(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockFileRepository)
	mockStorage := new(MockBlobStorage)
	mockPresigned := new(MockPresignedURLGenerator)
	config := &configs.Config{
		MinIO: configs.MinIOConfig{
			BucketName: "cloud-storage",
		},
	}

	svc := NewFileService(mockRepo, mockStorage, mockPresigned, config)

	mockRepo.On("CheckAccess", mock.Anything, "file-123", "user-123").Return(false, "", "", nil)

	input := &GetFileInfoInput{
		FileID: "file-123",
		UserID: "user-123",
	}

	output, err := svc.GetFileInfo(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access denied")
	assert.Nil(t, output)
	mockRepo.AssertExpectations(t)
}
