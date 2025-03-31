-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS day_records (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    quality INTEGER NOT NULL,
    mood VARCHAR(255) NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_day_records_user_id ON day_records (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_day_records_user_id;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS day_records;
-- +goose StatementEnd
