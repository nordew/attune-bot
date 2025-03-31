-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS focus_sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    quality INTEGER NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_focus_sessions_user_id ON focus_sessions (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_focus_sessions_user_id;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS focus_sessions;
-- +goose StatementEnd
