package storage

import (
	"attune/internal/models"
	"attune/pkg/apperrors"
	"context"
	"errors"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type UserStorage interface {
	Create(ctx context.Context, user models.User) error
	List(ctx context.Context, filter ListUserFilter) ([]models.User, int64, error)
	Update(ctx context.Context, user models.User) (models.User, error)
	Delete(ctx context.Context, id string) error
}

type ListUserFilter struct {
	ID         string `json:"id"`
	VendorID   string `json:"vendorId"`
	Name       string `json:"name"`
	VendorType string `json:"vendorType"`
}

type userStorage struct {
	conn    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

func NewUserStorage(conn *pgxpool.Pool) UserStorage {
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	return &userStorage{
		conn:    conn,
		builder: builder,
	}
}

func (s *userStorage) Create(ctx context.Context, user models.User) error {
	query, args, err := s.builder.
		Insert(userTableName).
		Columns(
			"id",
			"vendor_id",
			"vendor_type",
			"name",
			"created_at",
			"updated_at",
		).
		Values(user.ID, user.VendorID, user.VendorType, user.Name, time.Now(), time.Now()).
		ToSql()
	if err != nil {
		return err
	}

	_, err = s.conn.Exec(ctx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == codeUnique {
			return apperrors.NewConflict().WithDescription("user already exists")
		}

		return apperrors.NewInternal().WithDescriptionAndCause("failed to create user", err)
	}

	return nil
}

func (s *userStorage) List(ctx context.Context, filter ListUserFilter) ([]models.User, int64, error) {
	var users []models.User
	var totalCount int64

	qb := s.builder.
		Select(
			"id",
			"vendor_id",
			"vendor_type",
			"name",
			"COUNT(*) OVER() AS total_count",
		).
		From(userTableName)

	if filter.ID != "" {
		qb = qb.Where(squirrel.Eq{"id": filter.ID})
	}
	if filter.VendorID != "" {
		qb = qb.Where(squirrel.Eq{"vendor_id": filter.VendorID})
	}
	if filter.Name != "" {
		qb = qb.Where(squirrel.Eq{"name": filter.Name})
	}
	if filter.VendorType != "" {
		qb = qb.Where(squirrel.Eq{"vendor_type": filter.VendorType})
	}

	query, args, err := qb.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewInternal().WithDescriptionAndCause("failed to build query", err)
	}

	rows, err := s.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, apperrors.NewInternal().WithDescriptionAndCause("failed to execute query", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			user  models.User
			count int64
		)
		if err := rows.Scan(&user.ID, &user.VendorID, &user.VendorType, &user.Name, &count); err != nil {
			return nil, 0, apperrors.NewInternal().WithDescriptionAndCause("failed to scan user", err)
		}

		totalCount = count
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if len(users) == 0 {
		return nil, 0, apperrors.NewNotFound().WithDescription("no users found")
	}

	return users, totalCount, nil
}

func (s *userStorage) Update(ctx context.Context, user models.User) (models.User, error) {
	query, args, err := s.builder.
		Update(userTableName).
		Set("vendor_id", user.VendorID).
		Set("vendor_type", user.VendorType).
		Set("name", user.Name).
		Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": user.ID}).
		Suffix("RETURNING id, vendor_id, vendor_type, name").
		ToSql()
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, apperrors.NewNotFound().WithDescription("user not found")
		}

		return models.User{}, apperrors.NewInternal().WithDescriptionAndCause("failed to update user", err)
	}

	var updatedUser models.User
	err = s.conn.QueryRow(ctx, query, args...).Scan(
		&updatedUser.ID,
		&updatedUser.VendorID,
		&updatedUser.VendorType,
		&updatedUser.Name,
	)
	if err != nil {
		return models.User{}, apperrors.NewInternal().WithDescriptionAndCause("failed to scan updated user", err)
	}

	return updatedUser, nil
}

func (s *userStorage) Delete(ctx context.Context, id string) error {
	query, args, err := s.builder.
		Delete(userTableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return apperrors.NewInternal().WithDescriptionAndCause("failed to build delete query", err)
	}

	ct, err := s.conn.Exec(ctx, query, args...)
	if err != nil {
		return apperrors.NewInternal().WithDescriptionAndCause("failed to execute delete query", err)
	}

	if ct.RowsAffected() == 0 {
		return apperrors.NewNotFound().WithDescription("user not found")
	}

	return nil
}
