package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Sene4ka/cloud_storage/internal/api"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type MetadataClient interface {
	GetMetadata(ctx context.Context, in *api.GetMetadataRequest, opts ...grpc.CallOption) (*api.GetMetadataResponse, error)
	ListMetadata(ctx context.Context, in *api.ListMetadataRequest, opts ...grpc.CallOption) (*api.ListMetadataResponse, error)
	UpdateMetadata(ctx context.Context, in *api.UpdateMetadataRequest, opts ...grpc.CallOption) (*api.UpdateMetadataResponse, error)
	TrashFile(ctx context.Context, in *api.TrashFileRequest, opts ...grpc.CallOption) (*api.TrashFileResponse, error)
	RestoreFile(ctx context.Context, in *api.RestoreFileRequest, opts ...grpc.CallOption) (*api.RestoreFileResponse, error)
}

type FileClient interface {
	InitiateUpload(ctx context.Context, in *api.InitiateUploadRequest, opts ...grpc.CallOption) (*api.InitiateUploadResponse, error)
	CompleteUpload(ctx context.Context, in *api.CompleteUploadRequest, opts ...grpc.CallOption) (*api.CompleteUploadResponse, error)
	GetDownloadLink(ctx context.Context, in *api.GetDownloadLinkRequest, opts ...grpc.CallOption) (*api.GetDownloadLinkResponse, error)
	DeleteFile(ctx context.Context, in *api.DeleteFileRequest, opts ...grpc.CallOption) (*api.DeleteFileResponse, error)
}

type FileHandler struct {
	metadataClient MetadataClient
	fileClient     FileClient
}

func NewFileHandler(metadataClient MetadataClient, fileClient FileClient) *FileHandler {
	return &FileHandler{metadataClient: metadataClient, fileClient: fileClient}
}

func (h *FileHandler) HandleFiles(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	switch r.Method {
	case http.MethodGet:
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}

		pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
		if pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}

		var isTrashed *wrapperspb.BoolValue
		if thr := r.URL.Query().Get("is_trashed"); thr != "" {
			if val, err := strconv.ParseBool(thr); err == nil {
				isTrashed = wrapperspb.Bool(val)
			}
		}

		resp, err := h.metadataClient.ListMetadata(r.Context(), &api.ListMetadataRequest{
			UserId:    userID,
			Page:      int32(page),
			PageSize:  int32(pageSize),
			SortBy:    r.URL.Query().Get("sort_by"),
			SortOrder: r.URL.Query().Get("sort_order"),
			Search:    r.URL.Query().Get("search"),
			IsTrashed: isTrashed,
		})

		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		JSONResponse(w, http.StatusOK, resp)
	default:
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (h *FileHandler) HandleFileDetail(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	fileID := strings.TrimPrefix(r.URL.Path, "/api/v1/files/")
	switch r.Method {
	case http.MethodGet:
		resp, err := h.metadataClient.GetMetadata(r.Context(), &api.GetMetadataRequest{
			Id:     fileID,
			UserId: userID,
		})

		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusNotFound)
			return
		}
		JSONResponse(w, http.StatusOK, resp)
	case http.MethodPut:
		var req api.UpdateMetadataRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
			return
		}

		req.Id = fileID
		req.UserId = userID
		resp, err := h.metadataClient.UpdateMetadata(r.Context(), &req)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		JSONResponse(w, http.StatusOK, resp)
	case http.MethodDelete:
		_, err := h.fileClient.DeleteFile(r.Context(), &api.DeleteFileRequest{
			FileId: fileID,
			UserId: userID,
		})

		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}

		JSONResponse(w, http.StatusOK, map[string]bool{"success": true})
	default:
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (h *FileHandler) HandleInitiateUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("userID").(string)
	var req api.InitiateUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	req.UserId = userID
	resp, err := h.fileClient.InitiateUpload(r.Context(), &req)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", resp.UploadUrl)
	JSONResponse(w, http.StatusCreated, resp)
}

func (h *FileHandler) HandleCompleteUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("userID").(string)
	var req api.CompleteUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	req.UserId = userID
	resp, err := h.fileClient.CompleteUpload(r.Context(), &req)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, http.StatusOK, resp)
}

func (h *FileHandler) HandleDownloadLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("userID").(string)
	fileID := strings.TrimPrefix(r.URL.Path, "/api/v1/files/download/")
	expiresIn := int64(3600) // 1 hour default
	if exp := r.URL.Query().Get("expires_in"); exp != "" {
		if val, err := strconv.ParseInt(exp, 10, 64); err == nil {
			expiresIn = val
		}
	}

	resp, err := h.fileClient.GetDownloadLink(r.Context(), &api.GetDownloadLinkRequest{
		FileId:    fileID,
		UserId:    userID,
		ExpiresIn: expiresIn,
	})

	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusNotFound)
		return
	}
	JSONResponse(w, http.StatusOK, resp)
}

func (h *FileHandler) HandleTrashFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("userID").(string)
	fileID := strings.TrimPrefix(r.URL.Path, "/api/v1/files/trash/")

	resp, err := h.metadataClient.TrashFile(r.Context(), &api.TrashFileRequest{
		FileId: fileID,
		UserId: userID,
	})

	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, http.StatusOK, resp)
}

func (h *FileHandler) HandleRestoreFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("userID").(string)
	fileID := strings.TrimPrefix(r.URL.Path, "/api/v1/files/restore/")

	resp, err := h.metadataClient.RestoreFile(r.Context(), &api.RestoreFileRequest{
		FileId: fileID,
		UserId: userID,
	})

	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, http.StatusOK, resp)
}
