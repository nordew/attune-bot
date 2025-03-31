package storage

import (
	"attune/internal/models"
	"attune/pkg/apperrors"
	"context"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type DayRecordStorage interface {
	Create(ctx context.Context, record models.DayRecord) error
	List(ctx context.Context, filter ListDayRecordFilter) ([]models.DayRecord, int64, error)
	Update(ctx context.Context, record models.DayRecord) error
	Delete(ctx context.Context, id string) error
}

type ListDayRecordFilter struct {
	ID     string `json:"id"`
	UserID string `json:"userId"`
}

type dayRecordStorage struct {
	conn    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

func NewDayRecordStorage(conn *pgxpool.Pool) DayRecordStorage {
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	return &dayRecordStorage{
		conn:    conn,
		builder: builder,
	}
}

func (s *dayRecordStorage) Create(ctx context.Context, record models.DayRecord) error {
	query, args, err := s.builder.
		Insert(dayRecordsTableName).
		Columns(
			"id",
			"user_id",
			"quality",
			"mood",
			"created_at",
			"updated_at",
		).
		Values(
			record.ID,
			record.UserID,
			record.Quality,
			record.Mood,
			record.CreatedAt,
			record.UpdatedAt,
		).
		ToSql()
	if err != nil {
		return apperrors.NewInternal().WithDescriptionAndCause("failed to build create day record query", err)
	}

	_, err = s.conn.Exec(ctx, query, args...)
	if err != nil {
		return apperrors.NewInternal().WithDescriptionAndCause("failed to create day record", err)
	}

	return nil
}

func (s *dayRecordStorage) List(ctx context.Context, filter ListDayRecordFilter) ([]models.DayRecord, int64, error) {
	var records []models.DayRecord
	var totalCount int64

	qb := s.builder.
		Select(
			"id",
			"user_id",
			"quality",
			"mood",
			"created_at",
			"updated_at",
			"COUNT(*) OVER() AS total_count",
		).
		From(dayRecordsTableName)

	if filter.ID != "" {
		qb = qb.Where(squirrel.Eq{"id": filter.ID})
	}
	if filter.UserID != "" {
		qb = qb.Where(squirrel.Eq{"user_id": filter.UserID})
	}

	query, args, err := qb.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewInternal().WithDescriptionAndCause("failed to build list day records query", err)
	}

	rows, err := s.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, apperrors.NewInternal().WithDescriptionAndCause("failed to list day records", err)
	}
	defer rows.Close()

	for rows.Next() {
		var record models.DayRecord
		var count int64
		if err := rows.Scan(
			&record.ID,
			&record.UserID,
			&record.Quality,
			&record.Mood,
			&record.CreatedAt,
			&record.UpdatedAt,
			&count,
		); err != nil {
			return nil, 0, apperrors.NewInternal().WithDescriptionAndCause("failed to scan day record", err)
		}
		if totalCount == 0 {
			totalCount = count
		}
		records = append(records, record)
	}
	if len(records) == 0 {
		return nil, 0, apperrors.NewNotFound().WithDescription("day records not found")
	}

	return records, totalCount, nil
}

func (s *dayRecordStorage) Update(ctx context.Context, record models.DayRecord) error {
	query, args, err := s.builder.
		Update(dayRecordsTableName).
		Set("quality", record.Quality).
		Set("mood", record.Mood).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"id": record.ID}).
		ToSql()
	if err != nil {
		return apperrors.NewInternal().WithDescriptionAndCause("failed to build update day record query", err)
	}

	result, err := s.conn.Exec(ctx, query, args...)
	if err != nil {
		return apperrors.NewInternal().WithDescriptionAndCause("failed to update day record", err)
	}
	if result.RowsAffected() == 0 {
		return apperrors.NewNotFound().WithDescription("day record not found")
	}

	return nil
}

func (s *dayRecordStorage) Delete(ctx context.Context, id string) error {
	query, args, err := s.builder.
		Delete(dayRecordsTableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return apperrors.NewInternal().WithDescriptionAndCause("failed to build delete day record query", err)
	}

	result, err := s.conn.Exec(ctx, query, args...)
	if err != nil {
		return apperrors.NewInternal().WithDescriptionAndCause("failed to delete day record", err)
	}
	if result.RowsAffected() == 0 {
		return apperrors.NewNotFound().WithDescription("day record not found")
	}

	return nil
}
