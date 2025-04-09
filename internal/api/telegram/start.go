package telegram

import (
	"context"
	"strconv"

	"attune/internal/dto"
	"attune/internal/storage"
	"attune/pkg/apperrors"

	tb "gopkg.in/telebot.v4"
)

const (
	msgWelcome = "👋 *Welcome to Attune!* ✨\n\n" +
		"_I'm here to help you track your mood 😊 and stay focused 🎯._\n\n" +
		"Ready to start your journey? 🚀"
	msgWelcomeBack = "👋 *Welcome back to Attune!* ✨\n\n"
	msgRoadmap     = "🚀 *Roadmap* 🚀\n\n" +
		"Stay tuned! We'll soon have:\n" +
		"• Day quality charts and history to track your daily well-being 📊\n" +
		"• Enhanced focus sessions to keep you on track 🎯\n" +
		"• Personalized AI-powered journaling for deeper insights 🤖📝"
)

var (
	ErrMsgListUsers   = "failed to list users"
	ErrMsgCreateUser  = "failed to create user"
	ErrMsgSendWelcome = "failed to send welcome message"
	ErrMsgSendRoadmap = "failed to send roadmap message"
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

	_, userCount, err := a.services.UserService.List(ctx, storage.ListUserFilter{VendorID: vendorID})
	if err != nil && !apperrors.IsCode(err, apperrors.NotFound) {
		return apperrors.NewInternal().WithDescriptionAndCause(ErrMsgListUsers, err)
	} else if err == nil && userCount > 0 {
		return a.welcomeBack(c)
	}

	createReq := dto.CreateUserRequest{
		VendorID:   vendorID,
		VendorType: "telegram",
		Name:       c.Sender().FirstName,
	}
	if err := a.services.UserService.Create(ctx, createReq); err != nil {
		return apperrors.NewInternal().WithDescriptionAndCause(ErrMsgCreateUser, err)
	}
	sendOpts := &tb.SendOptions{ParseMode: tb.ModeMarkdown}
	if _, err := a.bot.Send(c.Sender(), msgWelcome, sendOpts); err != nil {
		return apperrors.NewInternal().WithDescriptionAndCause(ErrMsgSendWelcome, err)
	}

	if _, err := a.bot.Send(c.Sender(), msgRoadmap, sendOpts); err != nil {
		return apperrors.NewInternal().WithDescriptionAndCause(ErrMsgSendRoadmap, err)
	}

	return a.createFocusSession(c)
}

func (a *API) welcomeBack(c tb.Context) error {
	sendOpts := &tb.SendOptions{ParseMode: tb.ModeMarkdown}
	if _, err := a.bot.Send(c.Sender(), msgWelcomeBack, sendOpts); err != nil {
		return apperrors.NewInternal().WithDescriptionAndCause(ErrMsgSendWelcome, err)
	}

	return a.createFocusSession(c)
}
