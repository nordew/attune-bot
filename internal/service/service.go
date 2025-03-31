package service

import (
	"attune/internal/storage"
	"attune/pkg/cache"
	"attune/pkg/logger"
	"attune/pkg/transactor"
)

type Services struct {
	UserService         UserService
	UserSettingsService UserSettingsService
	FocusSessionManager FocusSessionManager
	FocusSessionService FocusSessionService
	cache               cache.Cache
}

func NewServices(
	storages storage.Storages,
	focusSessionManager FocusSessionManager,
	transactor transactor.Transactor,
	logger logger.Logger,
	cache cache.Cache,
) *Services {
	return &Services{
		UserService:         NewUserService(storages, logger),
		UserSettingsService: NewUserSettingsService(storages, logger),
		FocusSessionService: NewFocusSessionService(storages, focusSessionManager, transactor, logger, cache),
	}
}
