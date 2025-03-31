package service

import (
	"attune/internal/dto"
	"attune/internal/models"
	"attune/internal/storage"
	"attune/pkg/logger"
	"context"
	"time"
)

type UserSettingsService interface {
	Create(ctx context.Context, input dto.CreateUserSettingsRequest) error
	List(ctx context.Context, filter storage.ListUserSettingsFilter) ([]models.UserSettings, error)
	UpdateSentDailyStatsAt(ctx context.Context, userID string, sentDailyStatsAt time.Time) error
	Delete(ctx context.Context, userID string) error
}

type userSettingsService struct {
	storages storage.Storages
	logger   logger.Logger
}

func NewUserSettingsService(storages storage.Storages, logger logger.Logger) UserSettingsService {
	return &userSettingsService{
		storages: storages,
		logger:   logger,
	}
}

func (s *userSettingsService) Create(ctx context.Context, input dto.CreateUserSettingsRequest) error {
	const op = "userSettingsService.Create"

	log := s.logger.With("operation", op)

	settings, err := models.NewUserSettings(
		input.UserID,
		input.SentDailyStatsAt,
	)
	if err != nil {
		log.Error(ctx, "failed to create user settings", err)
		return err
	}

	err = s.storages.UserSettings.Create(ctx, settings)
	if err != nil {
		log.Error(ctx, "failed to create user settings in storage", err)
		return err
	}

	return nil
}

func (s *userSettingsService) List(ctx context.Context, filter storage.ListUserSettingsFilter) ([]models.UserSettings, error) {
	const op = "userSettingsService.List"

	log := s.logger.With("operation", op)

	settings, err := s.storages.UserSettings.List(ctx, filter)
	if err != nil {
		log.Error(ctx, "failed to list user settings", err)
		return nil, err
	}

	return settings, nil
}

func (s *userSettingsService) UpdateSentDailyStatsAt(ctx context.Context, userID string, sentDailyStatsAt time.Time) error {
	const op = "userSettingsService.UpdateSentDailyStatsAt"

	log := s.logger.With("operation", op)

	err := s.storages.UserSettings.UpdateSentDailyStatsAt(ctx, userID, sentDailyStatsAt)
	if err != nil {
		log.Error(ctx, "failed to update sent daily stats at", err)
		return err
	}

	return nil
}

func (s *userSettingsService) Delete(ctx context.Context, userID string) error {
	const op = "userSettingsService.Delete"

	log := s.logger.With("operation", op)

	err := s.storages.UserSettings.Delete(ctx, userID)
	if err != nil {
		log.Error(ctx, "failed to delete user settings", err)
		return err
	}

	return nil
}
