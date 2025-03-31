package service

import (
	"attune/internal/dto"
	"attune/internal/models"
	"attune/internal/storage"
	"attune/pkg/logger"
	"context"
)

type DayService interface {
	Create(ctx context.Context, input dto.CreateDayRecordRequest) error
	List(ctx context.Context, filter storage.ListDayRecordFilter) ([]models.DayRecord, int64, error)
	Delete(ctx context.Context, id string) error
	//Update(ctx context.Context, input dto.UpdateUserRequest) (models.User, error)
}

type dayService struct {
	storages storage.Storages
	logger   logger.Logger
}

func NewDayService(storages storage.Storages, logger logger.Logger) DayService {
	return &dayService{
		storages: storages,
		logger:   logger,
	}
}

func (s *dayService) Create(ctx context.Context, input dto.CreateDayRecordRequest) error {
	const op = "dayService.Create"

	log := s.logger.With("operation", op)

	dayRecord, err := models.NewDayRecord(
		input.UserID,
		input.Quality,
		input.Mood,
	)
	if err != nil {
		log.Error(ctx, "failed to create day record", err)
		return err
	}

	err = s.storages.DayRecord.Create(ctx, dayRecord)
	if err != nil {
		log.Error(ctx, "failed to create day record in storage", err)
		return err
	}

	return nil
}

func (s *dayService) List(ctx context.Context, filter storage.ListDayRecordFilter) ([]models.DayRecord, int64, error) {
	const op = "dayService.List"

	log := s.logger.With("operation", op)

	dayRecords, count, err := s.storages.DayRecord.List(ctx, filter)
	if err != nil {
		log.Error(ctx, "failed to list day records", err)
		return nil, 0, err
	}

	return dayRecords, count, nil
}

func (s *dayService) Delete(ctx context.Context, id string) error {
	const op = "dayService.Delete"

	log := s.logger.With("operation", op)

	err := s.storages.DayRecord.Delete(ctx, id)
	if err != nil {
		log.Error(ctx, "failed to delete day record", err)
		return err
	}

	return nil
}
