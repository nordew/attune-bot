package telegram

import (
	"attune/internal/dto"
	"attune/internal/storage"
	"attune/pkg/apperrors"
	"context"
	"strconv"

	tb "gopkg.in/telebot.v4"
)

const (
	msgWelcome = "👋 *Welcome to Attune!* ✨\n\n" +
		"_I'm here to help you track your mood 😊 and stay focused 🎯._\n\n" +
		"Ready to start your journey? 🚀"
)

func RegisterStartCommand(api *API) {
	api.bot.Handle("/start", func(c tb.Context) error {
		return api.handleStart(c)
	})
}

func (a *API) handleStart(c tb.Context) error {
	const op = "API.handleStart"
	ctx := context.Background()

	log := a.logger.With("operation", op, "user", c.Sender().ID)

	vendorID := strconv.FormatInt(c.Sender().ID, 10)

	users, _, err := a.services.UserService.List(ctx, storage.ListUserFilter{
		VendorID: vendorID,
	})
	if err != nil && !apperrors.IsCode(err, apperrors.NotFound) {
		return apperrors.NewInternal().WithDescriptionAndCause("failed to list users", err)
	}

	if err != nil || len(users) == 0 {
		createReq := dto.CreateUserRequest{
			VendorID:   vendorID,
			VendorType: "telegram",
			Name:       c.Sender().FirstName,
		}
		if err := a.services.UserService.Create(ctx, createReq); err != nil {
			return apperrors.NewInternal().WithDescriptionAndCause("failed to create user", err)
		}
	}

	sendOpts := &tb.SendOptions{
		ParseMode: tb.ModeMarkdown,
	}
	if _, err := a.bot.Send(c.Sender(), msgWelcome, sendOpts); err != nil {
		log.Error(ctx, "failed to send welcome message", err)
		return apperrors.NewInternal().WithDescriptionAndCause("failed to send welcome message", err)
	}

	return nil
}
