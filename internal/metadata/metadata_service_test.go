package metadata

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Sene4ka/cloud_storage/internal/models"
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

func (m *MockFileRepository) ListByUserID(
	ctx context.Context,
	userID string,
	page, pageSize int,
	sortBy, sortOrder, search string,
	isTrashed *bool,
) ([]*models.File, int, error) {
	args := m.Called(ctx, userID, page, pageSize, sortBy, sortOrder, search, isTrashed)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.File), args.Int(1), args.Error(2)
}

func (m *MockFileRepository) Update(ctx context.Context, file *models.File) error {
	args := m.Called(ctx, file)
	return args.Error(0)
}

func (m *MockFileRepository) CheckAccess(ctx context.Context, fileID, userID string) (bool, string, string, error) {
	args := m.Called(ctx, fileID, userID)
	return args.Bool(0), args.String(1), args.String(2), args.Error(3)
}

func (m *MockFileRepository) Delete(ctx context.Context, id, userID string) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockFileRepository) SetTrashed(ctx context.Context, fileID, userID string, isTrashed bool) error {
	args := m.Called(ctx, fileID, userID, isTrashed)
	return args.Error(0)
}

func TestMetadataService_GetMetadata_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockFileRepository)
	svc := NewMetadataService(mockRepo)

	expectedFile := &models.File{
		ID:           "file-123",
		UserID:       "user-456",
		Filename:     "test.txt",
		OriginalName: "test.txt",
		Path:         "/files/test.txt",
		Size:         1024,
		MimeType:     "text/plain",
		StoragePath:  "objects/file-123",
		Bucket:       "cloud-storage",
		IsPublic:     false,
		Tags:         map[string]string{"type": "document"},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsTrashed:    false,
		TrashedAt:    nil,
	}

	mockRepo.On("GetByID", mock.Anything, "file-123").Return(expectedFile, nil)

	input := &GetMetadataInput{
		FileID: "file-123",
		UserID: "user-456",
	}

	output, err := svc.GetMetadata(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, expectedFile, output.File)
	mockRepo.AssertExpectations(t)
}

func TestMetadataService_GetMetadata_AccessDenied(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockFileRepository)
	svc := NewMetadataService(mockRepo)

	otherUserFile := &models.File{
		ID:       "file-123",
		UserID:   "other-user",
		Filename: "test.txt",
		IsPublic: false,
		Tags:     map[string]string{},
	}

	mockRepo.On("GetByID", mock.Anything, "file-123").Return(otherUserFile, nil)

	input := &GetMetadataInput{
		FileID: "file-123",
		UserID: "user-456",
	}

	output, err := svc.GetMetadata(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access denied")
	assert.Nil(t, output)
	mockRepo.AssertExpectations(t)
}

func TestMetadataService_UpdateMetadata_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockFileRepository)
	svc := NewMetadataService(mockRepo)

	existingFile := &models.File{
		ID:           "file-123",
		UserID:       "user-456",
		Filename:     "old.txt",
		OriginalName: "old.txt",
		Path:         "/old",
		Size:         512,
		MimeType:     "text/plain",
		StoragePath:  "objects/file-123",
		Bucket:       "cloud-storage",
		IsPublic:     false,
		Tags:         map[string]string{"version": "1"},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsTrashed:    false,
		TrashedAt:    nil,
	}

	mockRepo.On("GetByID", mock.Anything, "file-123").Return(existingFile, nil)
	mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(f *models.File) bool {
		return f.Filename == "new.txt" && f.Tags["version"] == "2"
	})).Return(nil)

	input := &UpdateMetadataInput{
		FileID:       "file-123",
		UserID:       "user-456",
		Filename:     "new.txt",
		OriginalName: "new.txt",
		Path:         "/new",
		IsPublic:     true,
		Tags:         map[string]string{"version": "2", "status": "active"},
	}

	output, err := svc.UpdateMetadata(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "new.txt", output.File.Filename)
	mockRepo.AssertExpectations(t)
}

func TestMetadataService_TrashFile_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockFileRepository)
	svc := NewMetadataService(mockRepo)

	mockRepo.On("SetTrashed", mock.Anything, "file-123", "user-456", true).Return(nil)

	input := &TrashFileInput{
		FileID: "file-123",
		UserID: "user-456",
	}

	output, err := svc.TrashFile(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.True(t, output.Success)
	mockRepo.AssertExpectations(t)
}

func TestMetadataService_RestoreFile_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockFileRepository)
	svc := NewMetadataService(mockRepo)

	mockRepo.On("SetTrashed", mock.Anything, "file-123", "user-456", false).Return(nil)

	input := &RestoreFileInput{
		FileID: "file-123",
		UserID: "user-456",
	}

	output, err := svc.RestoreFile(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.True(t, output.Success)
	mockRepo.AssertExpectations(t)
}

func TestMetadataService_GetMetadata_RepoError(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockFileRepository)
	svc := NewMetadataService(mockRepo)

	mockRepo.On("GetByID", mock.Anything, "file-123").Return(nil, errors.New("db error"))

	input := &GetMetadataInput{
		FileID: "file-123",
		UserID: "user-456",
	}

	output, err := svc.GetMetadata(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get metadata")
	assert.Nil(t, output)
	mockRepo.AssertExpectations(t)
}

func TestMetadataService_ListMetadata_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockFileRepository)
	svc := NewMetadataService(mockRepo)

	files := []*models.File{
		{
			ID:       "file-1",
			UserID:   "user-456",
			Filename: "test1.txt",
			Tags:     map[string]string{"type": "doc"},
		},
		{
			ID:       "file-2",
			UserID:   "user-456",
			Filename: "test2.txt",
			Tags:     map[string]string{"type": "image"},
		},
	}

	mockRepo.On("ListByUserID", mock.Anything, "user-456", 1, 10, "created_at", "desc", "", (*bool)(nil)).
		Return(files, 2, nil)

	input := &ListMetadataInput{
		UserID:    "user-456",
		Page:      1,
		PageSize:  10,
		SortBy:    "created_at",
		SortOrder: "desc",
	}

	output, err := svc.ListMetadata(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Len(t, output.Items, 2)
	assert.Equal(t, int64(2), output.Total)
	mockRepo.AssertExpectations(t)
}
