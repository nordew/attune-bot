-- +goose Up
-- +goose StatementBegin
ALTER TABLE focus_sessions ADD COLUMN IF NOT EXISTS status VARCHAR(255) DEFAULT 'active' NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE focus_sessions DROP COLUMN IF EXISTS status;
-- +goose StatementEnd
