package service

import (
	"attune/internal/dto"
	"attune/internal/models"
	"attune/internal/storage"
	"attune/pkg/logger"
	"context"
	"github.com/google/uuid"
)

type UserService interface {
	Create(ctx context.Context, input dto.CreateUserRequest) error
	List(ctx context.Context, filter storage.ListUserFilter) ([]models.User, int64, error)
	// TODO: implement update user
	//Update(ctx context.Context, input dto.UpdateUserRequest) (models.User, error)
	Delete(ctx context.Context, id string) error
}

type userService struct {
	storages storage.Storages
	logger   logger.Logger
}

func NewUserService(storages storage.Storages, logger logger.Logger) UserService {
	return &userService{
		storages: storages,
		logger:   logger,
	}
}

func (s *userService) Create(ctx context.Context, input dto.CreateUserRequest) error {
	const op = "userService.Create"

	log := s.logger.With("operation", op)

	user, err := models.NewUser(
		uuid.New().String(),
		input.VendorID,
		input.VendorType,
		input.Name,
	)
	if err != nil {
		log.Error(ctx, "failed to create user", err)
		return err
	}

	err = s.storages.User.Create(ctx, user)
	if err != nil {
		log.Error(ctx, "failed to create user in storage", err)
		return err
	}

	return nil
}

func (s *userService) List(ctx context.Context, filter storage.ListUserFilter) ([]models.User, int64, error) {
	const op = "userService.List"

	log := s.logger.With("operation", op)

	users, total, err := s.storages.User.List(ctx, filter)
	if err != nil {
		log.Error(ctx, "failed to list users", err)
		return nil, 0, err
	}

	return users, total, nil
}

func (s *userService) Delete(ctx context.Context, id string) error {
	const op = "userService.Delete"

	log := s.logger.With("operation", op)

	err := s.storages.User.Delete(ctx, id)
	if err != nil {
		log.Error(ctx, "failed to delete user", err)
		return err
	}

	return nil
}
