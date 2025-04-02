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

var (
	ErrMsgListUsers   = "failed to list users"
	ErrMsgCreateUser  = "failed to create user"
	ErrMsgSendWelcome = "failed to send welcome message"
)

func registerStartCommand(api *API) {
	api.bot.Handle("/start", func(c tb.Context) error {
		err := api.handleStart(c)
		if err != nil {
			api.logger.Error(context.Background(), "Error handling /start command", err, "user", c.Sender().ID)
		}

		return err
	})
}

func (a *API) handleStart(c tb.Context) error {
	ctx := context.Background()

	vendorID := strconv.FormatInt(c.Sender().ID, 10)

	users, _, err := a.services.UserService.List(ctx, storage.ListUserFilter{
		VendorID: vendorID,
	})
	if err != nil && !apperrors.IsCode(err, apperrors.NotFound) {
		return apperrors.NewInternal().WithDescriptionAndCause(ErrMsgListUsers, err)
	}

	if err != nil || len(users) == 0 {
		createReq := dto.CreateUserRequest{
			VendorID:   vendorID,
			VendorType: "telegram",
			Name:       c.Sender().FirstName,
		}
		if err := a.services.UserService.Create(ctx, createReq); err != nil {
			return apperrors.NewInternal().WithDescriptionAndCause(ErrMsgCreateUser, err)
		}
	}

	sendOpts := &tb.SendOptions{
		ParseMode: tb.ModeMarkdown,
	}
	if _, err := a.bot.Send(c.Sender(), msgWelcome, sendOpts); err != nil {
		return apperrors.NewInternal().WithDescriptionAndCause(ErrMsgSendWelcome, err)
	}

	return a.createFocusSession(c)
}
