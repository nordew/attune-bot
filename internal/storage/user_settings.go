package storage

import (
	"attune/internal/models"
	"attune/pkg/apperrors"
	"context"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type UserSettingsStorage interface {
	Create(ctx context.Context, settings models.UserSettings) error
	List(ctx context.Context, filter ListUserSettingsFilter) ([]models.UserSettings, error)
	UpdateSentDailyStatsAt(ctx context.Context, userID string, sentDailyStatsAt time.Time) error
	Delete(ctx context.Context, userID string) error
}

type ListUserSettingsFilter struct {
	ID                   string    `json:"id"`
	UserID               string    `json:"userId"`
	SentDailyStatsAfter  time.Time `json:"sentDailyStatsAfter"`
	SentDailyStatsBefore time.Time `json:"sentDailyStatsBefore"`
}

type userSettingsStorage struct {
	conn    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

func NewUserSettingsStorage(conn *pgxpool.Pool) UserSettingsStorage {
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	return &userSettingsStorage{
		conn:    conn,
		builder: builder,
	}
}

func (s *userSettingsStorage) Create(ctx context.Context, settings models.UserSettings) error {
	query, args, err := s.builder.
		Insert(userSettingsTableName).
		Columns(
			"id",
			"user_id",
			"sent_daily_stats_at",
			"created_at",
			"updated_at",
		).
		Values(
			settings.ID,
			settings.UserID,
			settings.SentDailyStatsAt,
			settings.CreatedAt,
			settings.UpdatedAt,
		).
		ToSql()
	if err != nil {
		return apperrors.NewInternal().WithDescriptionAndCause("failed to build create user settings query", err)
	}

	_, err = s.conn.Exec(ctx, query, args...)
	if err != nil {
		return apperrors.NewInternal().WithDescriptionAndCause("failed to create user settings", err)
	}

	return nil
}

func (s *userSettingsStorage) List(ctx context.Context, filter ListUserSettingsFilter) ([]models.UserSettings, error) {
	var settingsList []models.UserSettings

	qb := s.builder.
		Select(
			"id",
			"user_id",
			"sent_daily_stats_at",
			"created_at",
			"updated_at",
		).
		From(userSettingsTableName)

	if filter.ID != "" {
		qb = qb.Where(squirrel.Eq{"id": filter.ID})
	}
	if filter.UserID != "" {
		qb = qb.Where(squirrel.Eq{"user_id": filter.UserID})
	}
	if !filter.SentDailyStatsAfter.IsZero() {
		qb = qb.Where(squirrel.Gt{"sent_daily_stats_at": filter.SentDailyStatsAfter})
	}
	if !filter.SentDailyStatsBefore.IsZero() {
		qb = qb.Where(squirrel.Lt{"sent_daily_stats_at": filter.SentDailyStatsBefore})
	}

	query, args, err := qb.ToSql()
	if err != nil {
		return nil, apperrors.NewInternal().WithDescriptionAndCause("failed to build list user settings query", err)
	}

	rows, err := s.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, apperrors.NewInternal().WithDescriptionAndCause("failed to list user settings", err)
	}
	defer rows.Close()

	for rows.Next() {
		var settings models.UserSettings
		if err := rows.Scan(
			&settings.ID,
			&settings.UserID,
			&settings.SentDailyStatsAt,
			&settings.CreatedAt,
			&settings.UpdatedAt,
		); err != nil {
			return nil, apperrors.NewInternal().WithDescriptionAndCause("failed to scan user settings", err)
		}

		settingsList = append(settingsList, settings)
	}
	if len(settingsList) == 0 {
		return nil, apperrors.NewNotFound().WithDescription("no user settings found")
	}

	return settingsList, nil
}

func (s *userSettingsStorage) UpdateSentDailyStatsAt(
	ctx context.Context,
	userID string,
	sentDailyStatsAt time.Time,
) error {
	query, args, err := s.builder.
		Update(userSettingsTableName).
		Set("sent_daily_stats_at", sentDailyStatsAt).
		Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"user_id": userID}).
		ToSql()
	if err != nil {
		return apperrors.NewInternal().WithDescriptionAndCause("failed to build update user settings query", err)
	}

	result, err := s.conn.Exec(ctx, query, args...)
	if err != nil {
		return apperrors.NewInternal().WithDescriptionAndCause("failed to update user settings", err)
	}
	if result.RowsAffected() == 0 {
		return apperrors.NewNotFound().WithDescription("user settings not found")
	}
	return nil
}

func (s *userSettingsStorage) Delete(ctx context.Context, userID string) error {
	query, args, err := s.builder.
		Delete(userSettingsTableName).
		Where(squirrel.Eq{"user_id": userID}).
		ToSql()
	if err != nil {
		return apperrors.NewInternal().WithDescriptionAndCause("failed to build delete user settings query", err)
	}

	result, err := s.conn.Exec(ctx, query, args...)
	if err != nil {
		return apperrors.NewInternal().WithDescriptionAndCause("failed to delete user settings", err)
	}
	if result.RowsAffected() == 0 {
		return apperrors.NewNotFound().WithDescription("user settings not found")
	}
	return nil
}
