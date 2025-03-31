package storage

import (
	"attune/internal/models"
	"attune/pkg/apperrors"
	"context"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type FocusSessionStorage interface {
	Create(ctx context.Context, session models.FocusSession) error
	List(ctx context.Context, filter ListFocusSessionFilter) ([]models.FocusSession, int64, error)
	Update(ctx context.Context, session models.FocusSession) error
	Delete(ctx context.Context, id string) error
}

type ListFocusSessionFilter struct {
	ID     string `json:"id"`
	UserID string `json:"userId"`
}

type focusSessionStorage struct {
	conn    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

func NewFocusSessionStorage(conn *pgxpool.Pool) FocusSessionStorage {
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	return &focusSessionStorage{
		conn:    conn,
		builder: builder,
	}
}

func (s *focusSessionStorage) Create(ctx context.Context, session models.FocusSession) error {
	query, args, err := s.builder.
		Insert(focusSessionsTableName).
		Columns(
			"id",
			"user_id",
			"quality",
			"started_at",
			"ended_at",
			"created_at",
			"updated_at",
		).
		Values(
			session.ID,
			session.UserID,
			session.Quality,
			session.StartedAt,
			session.EndedAt,
			session.CreatedAt,
			session.UpdatedAt,
		).
		ToSql()
	if err != nil {
		return apperrors.NewInternal().WithDescriptionAndCause("failed to build create focus session query", err)
	}

	_, err = s.conn.Exec(ctx, query, args...)
	if err != nil {
		return apperrors.NewInternal().WithDescriptionAndCause("failed to create focus session", err)
	}

	return nil
}

func (s *focusSessionStorage) List(ctx context.Context, filter ListFocusSessionFilter) ([]models.FocusSession, int64, error) {
	var sessions []models.FocusSession
	var totalCount int64

	qb := s.builder.
		Select(
			"id",
			"user_id",
			"quality",
			"started_at",
			"ended_at",
			"created_at",
			"updated_at",
			"COUNT(*) OVER() AS total_count",
		).
		From(focusSessionsTableName)

	if filter.ID != "" {
		qb = qb.Where(squirrel.Eq{"id": filter.ID})
	}
	if filter.UserID != "" {
		qb = qb.Where(squirrel.Eq{"user_id": filter.UserID})
	}

	query, args, err := qb.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewInternal().WithDescriptionAndCause("failed to build list focus sessions query", err)
	}

	rows, err := s.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, apperrors.NewInternal().WithDescriptionAndCause("failed to list focus sessions", err)
	}
	defer rows.Close()

	for rows.Next() {
		var session models.FocusSession
		var count int64
		if err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.Quality,
			&session.StartedAt,
			&session.EndedAt,
			&session.CreatedAt,
			&session.UpdatedAt,
			&count,
		); err != nil {
			return nil, 0, apperrors.NewInternal().WithDescriptionAndCause("failed to scan focus session", err)
		}
		if totalCount == 0 {
			totalCount = count
		}
		sessions = append(sessions, session)
	}
	if len(sessions) == 0 {
		return nil, 0, apperrors.NewNotFound().WithDescription("no focus sessions found")
	}

	return sessions, totalCount, nil
}
func (s *focusSessionStorage) Update(ctx context.Context, session models.FocusSession) error {
	query, args, err := s.builder.
		Update(focusSessionsTableName).
		Set("quality", session.Quality).
		Set("started_at", session.StartedAt).
		Set("ended_at", session.EndedAt).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"id": session.ID}).
		ToSql()
	if err != nil {
		return apperrors.NewInternal().WithDescriptionAndCause("failed to build update focus session query", err)
	}

	result, err := s.conn.Exec(ctx, query, args...)
	if err != nil {
		return apperrors.NewInternal().WithDescriptionAndCause("failed to update focus session", err)
	}
	if result.RowsAffected() == 0 {
		return apperrors.NewNotFound().WithDescription("focus session not found")
	}

	return nil
}

func (s *focusSessionStorage) Delete(ctx context.Context, id string) error {
	query, args, err := s.builder.
		Delete(focusSessionsTableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return apperrors.NewInternal().WithDescriptionAndCause("failed to build delete focus session query", err)
	}

	result, err := s.conn.Exec(ctx, query, args...)
	if err != nil {
		return apperrors.NewInternal().WithDescriptionAndCause("failed to delete focus session", err)
	}
	if result.RowsAffected() == 0 {
		return apperrors.NewNotFound().WithDescription("focus session not found")
	}

	return nil
}
