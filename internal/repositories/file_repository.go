package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/Sene4ka/cloud_storage/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FileRepository struct {
	db *pgxpool.Pool
}

func NewFileRepository(db *pgxpool.Pool) *FileRepository {
	return &FileRepository{db: db}
}

func (r *FileRepository) Create(ctx context.Context, file *models.File) error {
	query := `
		INSERT INTO files (id, user_id, filename, original_name, size, mime_type, storage_path, bucket, is_public, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	tags := formatTags(file.Tags)
	_, err := r.db.Exec(ctx, query,
		file.ID,
		file.UserID,
		file.Filename,
		file.OriginalName,
		file.Size,
		file.MimeType,
		file.StoragePath,
		file.Bucket,
		file.IsPublic,
		tags,
		file.CreatedAt,
		file.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	return nil
}

func (r *FileRepository) GetByID(ctx context.Context, id string) (*models.File, error) {
	query := `
		SELECT id, user_id, filename, original_name, size, mime_type, storage_path, bucket, is_public, tags, created_at, updated_at
		FROM files
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)
	var file models.File
	var tags string
	err := row.Scan(
		&file.ID,
		&file.UserID,
		&file.Filename,
		&file.OriginalName,
		&file.Size,
		&file.MimeType,
		&file.StoragePath,
		&file.Bucket,
		&file.IsPublic,
		&tags,
		&file.CreatedAt,
		&file.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("file not found")
		}
		return nil, fmt.Errorf("failed to get file by id: %w", err)
	}

	file.Tags = parseTags(tags)
	return &file, nil
}
func (r *FileRepository) ListByUserID(ctx context.Context, userID string, page, pageSize int, sortBy, sortOrder, search string) ([]*models.File, int, error) {
	offset := (page - 1) * pageSize

	countQuery := `SELECT COUNT(*) FROM files WHERE user_id = $1`

	var total int
	err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count files: %w", err)
	}

	query := `
		SELECT id, user_id, filename, original_name, size, mime_type, storage_path, bucket, is_public, tags, created_at, updated_at
		FROM files
		WHERE user_id = $1
	`

	args := []interface{}{userID}
	argCount := 1
	if search != "" {
		argCount++
		query += fmt.Sprintf(" AND (filename ILIKE $%d OR original_name ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+search+"%")
	}

	if sortBy != "" {
		validSortFields := map[string]bool{"created_at": true, "updated_at": true, "filename": true, "size": true}
		if validSortFields[sortBy] {
			order := "ASC"
			if strings.ToUpper(sortOrder) == "DESC" {
				order = "DESC"
			}
			query += fmt.Sprintf(" ORDER BY %s %s", sortBy, order)
		}
	} else {
		query += " ORDER BY created_at DESC"
	}

	argCount++
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list files: %w", err)
	}

	defer rows.Close()

	var files []*models.File
	for rows.Next() {
		var file models.File
		var tags string
		err := rows.Scan(
			&file.ID,
			&file.UserID,
			&file.Filename,
			&file.OriginalName,
			&file.Size,
			&file.MimeType,
			&file.StoragePath,
			&file.Bucket,
			&file.IsPublic,
			&tags,
			&file.CreatedAt,
			&file.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan file: %w", err)
		}
		file.Tags = parseTags(tags)
		files = append(files, &file)
	}

	return files, total, nil
}

func (r *FileRepository) Update(ctx context.Context, file *models.File) error {
	query := `
		UPDATE files
		SET filename = $1, original_name = $2, is_public = $3, tags = $4, updated_at = $5
		WHERE id = $6 AND user_id = $7
	`

	tags := formatTags(file.Tags)
	result, err := r.db.Exec(ctx, query,
		file.Filename,
		file.OriginalName,
		file.IsPublic,
		tags,
		file.UpdatedAt,
		file.ID,
		file.UserID,
	)

	if err != nil {
		return fmt.Errorf("failed to update file: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("file not found or access denied")
	}

	return nil
}

func (r *FileRepository) Delete(ctx context.Context, id, userID string) error {
	query := `DELETE FROM files WHERE id = $1 AND user_id = $2`

	result, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("file not found or access denied")
	}
	return nil
}

func (r *FileRepository) CheckAccess(ctx context.Context, fileID, userID string) (bool, string, string, error) {
	query := `
		SELECT storage_path, bucket, is_public, user_id
		FROM files
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, fileID)
	var storagePath, bucket string
	var isPublic bool
	var fileUserID string
	err := row.Scan(&storagePath, &bucket, &isPublic, &fileUserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, "", "", fmt.Errorf("file not found")
		}
		return false, "", "", fmt.Errorf("failed to check access: %w", err)
	}

	if isPublic {
		return true, storagePath, bucket, nil
	}

	if fileUserID == userID {
		return true, storagePath, bucket, nil
	}

	return false, "", "", nil
}

func formatTags(tags map[string]string) string {
	if tags == nil {
		return ""
	}

	var parts []string
	for k, v := range tags {
		parts = append(parts, k+"="+v)
	}

	return strings.Join(parts, ",")
}

func parseTags(tags string) map[string]string {
	result := make(map[string]string)
	if tags == "" {
		return result
	}

	parts := strings.Split(tags, ",")
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		}
	}

	return result
}
