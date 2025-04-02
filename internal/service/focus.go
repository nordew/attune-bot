package service

import (
	"attune/internal/dto"
	"attune/internal/models"
	"attune/internal/storage"
	"attune/pkg/apperrors"
	"attune/pkg/cache"
	"attune/pkg/logger"
	"attune/pkg/transactor"
	"context"
	"fmt"
	"time"
)

var (
	errMsgListUsers         = "failed to list users for VendorID %s"
	errMsgNoUserFound       = "no user found for VendorID %s"
	errMsgCreateSession     = "failed to create focus session"
	errMsgTransactionFail   = "transaction failed while starting focus session"
	errMsgListSessions      = "failed to list focus sessions"
	errMsgUpdateInvalidType = "invalid update request type"
	errMsgUpdateFailure     = "failed to update focus session with type %s"
	errMsgDeleteSession     = "failed to delete focus session with id %s"
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
		errMsg := fmt.Sprintf(errMsgListUsers, input.VendorID)
		log.Error(ctx, errMsg, err)
		return err
	}
	user := users[0]

	focusSession, err := models.NewFocusSession(user.ID, input.Duration)
	if err != nil {
		log.Error(ctx, errMsgCreateSession, err)
		return err
	}

	if input.VendorID != "" {
		focusSession.VendorID = input.VendorID
	}

	if err := s.storages.FocusSession.Create(ctx, focusSession); err != nil {
		log.Error(ctx, errMsgCreateSession, err)
		return apperrors.NewInternal().WithDescriptionAndCause(errMsgCreateSession, err)
	}

	go s.focusSessionManager.Start(focusSession, input.Duration)

	return nil
}

func (s *focusSessionService) List(ctx context.Context, filter storage.ListFocusSessionFilter) ([]models.FocusSession, int64, error) {
	const op = "focusSessionService.List"
	log := s.logger.With("operation", op)

	focusSessions, count, err := s.storages.FocusSession.List(ctx, filter)
	if err != nil {
		log.Error(ctx, errMsgListSessions, err)
		return nil, 0, apperrors.NewInternal().WithDescriptionAndCause(errMsgListSessions, err)
	}

	return focusSessions, count, nil
}

func (s *focusSessionService) Update(ctx context.Context, input dto.UpdateFocusRequest) error {
	log := s.logger.With("operation", "focusSessionService.Update")

	users, _, err := s.storages.User.List(ctx, storage.ListUserFilter{
		VendorID: input.VendorID,
	})
	if err != nil {
		return apperrors.NewInternal().WithDescriptionAndCause(fmt.Sprintf(errMsgListUsers, input.VendorID), err)
	}
	user := users[0]

	return s.transactor.Transact(ctx, func(ctx context.Context) error {
		switch input.Type {
		case dto.UpdateFocusRequestTypePause:
			return s.focusSessionManager.Pause(user.ID)
		case dto.UpdateFocusRequestTypeResume:
			return s.focusSessionManager.Resume(user.ID)
		case dto.UpdateFocusRequestTypeStop:
			return s.focusSessionManager.Stop(user.ID)
		case dto.UpdateFocusRequestTypeQuality:
			sessions, _, err := s.storages.FocusSession.List(ctx, storage.ListFocusSessionFilter{
				UserID: user.ID,
			})
			if err != nil {
				log.Error(ctx, errMsgListSessions, err)
				return apperrors.NewInternal().WithDescriptionAndCause(fmt.Sprintf(errMsgListSessions, user.ID), err)
			}

			session := sessions[0]
			session.Quality = input.Quality
			session.Status = input.Status
			session.UpdatedAt = time.Now()
			session.EndedAt = time.Now()
			if err := s.storages.FocusSession.Update(ctx, session); err != nil {
				log.Error(ctx, errMsgUpdateFailure, err)
				return apperrors.NewInternal().WithDescriptionAndCause(fmt.Sprintf(errMsgUpdateFailure, input.Type), err)
			}

			return nil
		default:
			log.Error(ctx, errMsgUpdateInvalidType)
			return apperrors.NewBadRequest().WithDescription(errMsgUpdateInvalidType)
		}
	})
}

func (s *focusSessionService) Delete(ctx context.Context, id string) error {
	const op = "focusSessionService.Delete"
	log := s.logger.With("operation", op)

	if err := s.storages.FocusSession.Delete(ctx, id); err != nil {
		log.Error(ctx, errMsgDeleteSession, err)
		return apperrors.NewInternal().WithDescriptionAndCause(fmt.Sprintf(errMsgDeleteSession, id), err)
	}

	return nil
}
