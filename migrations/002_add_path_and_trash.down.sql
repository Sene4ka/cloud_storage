DROP INDEX IF EXISTS idx_files_user_path;
DROP INDEX IF EXISTS idx_files_user_trashed;
DROP INDEX IF EXISTS idx_files_trashed_at;

ALTER TABLE files DROP CONSTRAINT IF EXISTS files_user_path_original_unique;

ALTER TABLE files
DROP COLUMN IF EXISTS path,
DROP COLUMN IF EXISTS is_trashed,
DROP COLUMN IF EXISTS trashed_at;