ALTER TABLE files
ADD COLUMN IF NOT EXISTS path TEXT NOT NULL DEFAULT '/',
ADD COLUMN IF NOT EXISTS is_trashed BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS trashed_at TIMESTAMP WITH TIME ZONE DEFAULT NULL;

ALTER TABLE files ADD CONSTRAINT IF NOT EXISTS files_user_path_original_unique UNIQUE (user_id, path, original_name);

CREATE INDEX IF NOT EXISTS idx_files_user_path ON files(user_id, path);
CREATE INDEX IF NOT EXISTS idx_files_user_trashed ON files(user_id, is_trashed) WHERE is_trashed = TRUE;
CREATE INDEX IF NOT EXISTS idx_files_trashed_at ON files(trashed_at) WHERE trashed_at IS NOT NULL;