package service

import (
	"attune/internal/dto"
	"attune/internal/models"
	"attune/internal/storage"
	"attune/pkg/cache"
	"attune/pkg/logger"
	"attune/pkg/transactor"
	"context"
)

type FocusSessionService interface {
	Create(ctx context.Context, input dto.CreateFocusSessionRequest) error
	List(ctx context.Context, filter storage.ListFocusSessionFilter) ([]models.FocusSession, int64, error)
	Update(ctx context.Context, input dto.UpdateFocusRequest) error
	Delete(ctx context.Context, id string) error
}

type focusSessionService struct {
	storages            storage.Storages
	focusSessionManager FocusSessionManager
	transactor          transactor.Transactor
	logger              logger.Logger
	cache               cache.Cache
}

func NewFocusSessionService(
	storages storage.Storages,
	focusSessionManager FocusSessionManager,
	transactor transactor.Transactor,
	logger logger.Logger,
	cache cache.Cache,
) FocusSessionService {
	return &focusSessionService{
		storages:            storages,
		focusSessionManager: focusSessionManager,
		transactor:          transactor,
		logger:              logger,
		cache:               cache,
	}
}

func (s *focusSessionService) Create(ctx context.Context, input dto.CreateFocusSessionRequest) error {
	const op = "focusSessionService.Create"
	log := s.logger.With("operation", op)

	users, _, err := s.storages.User.List(ctx, storage.ListUserFilter{
		VendorID: input.VendorID,
	})
	if err != nil {
		log.Error(ctx, "failed to list users", err)
		return err
	}
	user := users[0]

	focusSession, err := models.NewFocusSession(
		user.ID,
		input.Duration,
	)
	if err != nil {
		log.Error(ctx, "failed to create focus session", err)
		return err
	}

	err = s.transactor.Transact(ctx, func(ctx context.Context) error {
		if err := s.storages.FocusSession.Create(ctx, focusSession); err != nil {
			log.Error(ctx, "failed to create focus session in storage", err)
			return err
		}

		s.focusSessionManager.Start(focusSession)

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *focusSessionService) List(ctx context.Context, filter storage.ListFocusSessionFilter) ([]models.FocusSession, int64, error) {
	const op = "focusSessionService.List"

	log := s.logger.With("operation", op)

	focusSessions, count, err := s.storages.FocusSession.List(ctx, filter)
	if err != nil {
		log.Error(ctx, "failed to list focus sessions", err)
		return nil, 0, err
	}

	return focusSessions, count, nil
}

func (s *focusSessionService) Update(ctx context.Context, input dto.UpdateFocusRequest) error {
	const op = "focusSessionService.Update"

	log := s.logger.With("operation", op)

	err := s.transactor.Transact(ctx, func(ctx context.Context) error {
		if input.Type == dto.UpdateFocusRequestTypePause {
			if err := s.focusSessionManager.Pause(input.ID); err != nil {
				log.Error(ctx, "failed to pause focus session", err)
				return err
			}
		}
		if input.Type == dto.UpdateFocusRequestTypeResume {
			if err := s.focusSessionManager.Resume(input.ID); err != nil {
				log.Error(ctx, "failed to resume focus session", err)
				return err
			}
		}
		if input.Type == dto.UpdateFocusRequestTypeStop {
			if err := s.focusSessionManager.Stop(input.ID); err != nil {
				log.Error(ctx, "failed to stop focus session", err)
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *focusSessionService) Delete(ctx context.Context, id string) error {
	const op = "focusSessionService.Delete"

	log := s.logger.With("operation", op)

	err := s.storages.FocusSession.Delete(ctx, id)
	if err != nil {
		log.Error(ctx, "failed to delete focus session", err)
		return err
	}

	return nil
}
