package telegram

import (
	"attune/internal/dto"
	"context"
	tb "gopkg.in/telebot.v4"
	"strconv"
	"time"

	"attune/pkg/apperrors"
)

const (
	keyFocus15            = "focus_15"
	keyFocus30            = "focus_30"
	keyFocus60            = "focus_60"
	keyFocusCustom        = "focus_custom"
	msgFocusSessionPrompt = "🔔 *Select your focus session duration:*"
	msgCustomPrompt       = "⌨️ _Please enter your desired duration_\n(e.g., `45m` for 45 minutes)."
	msgSessionStarted     = "✅ *Your focus session has started!*\nDuration: "
	msgInvalidDuration    = "❌ *Invalid duration format.*\nPlease try again (e.g., `45m`)."
)

func (a *API) createFocusSession(c tb.Context) error {
	const op = "API.createFocusSession"
	log := a.logger.With("operation", op)

	btn15 := tb.InlineButton{Unique: keyFocus15, Text: "15 min"}
	btn30 := tb.InlineButton{Unique: keyFocus30, Text: "30 min"}
	btn60 := tb.InlineButton{Unique: keyFocus60, Text: "60 min"}
	btnCustom := tb.InlineButton{Unique: keyFocusCustom, Text: "Custom 📝"}

	inlineKeys := [][]tb.InlineButton{
		{btn15, btn30},
		{btn60, btnCustom},
	}

	markup := &tb.ReplyMarkup{InlineKeyboard: inlineKeys}
	opts := &tb.SendOptions{ParseMode: tb.ModeMarkdown}

	if _, err := a.bot.Send(c.Sender(), msgFocusSessionPrompt, markup, opts); err != nil {
		log.Error(context.Background(), "failed to send focus session menu", err)
		return apperrors.NewInternal().WithDescriptionAndCause("failed to send focus session menu", err)
	}

	return nil
}

func (a *API) registerFocusSessionCallbacks() {
	a.bot.Handle(&tb.InlineButton{Unique: keyFocus15}, func(c tb.Context) error {
		return a.startFocusSession(c, 15*time.Minute)
	})
	a.bot.Handle(&tb.InlineButton{Unique: keyFocus30}, func(c tb.Context) error {
		return a.startFocusSession(c, 30*time.Minute)
	})
	a.bot.Handle(&tb.InlineButton{Unique: keyFocus60}, func(c tb.Context) error {
		return a.startFocusSession(c, 60*time.Minute)
	})
	a.bot.Handle(&tb.InlineButton{Unique: keyFocusCustom}, func(c tb.Context) error {
		a.cache.Set(strconv.FormatInt(c.Sender().ID, 10), true)
		opts := &tb.SendOptions{ParseMode: tb.ModeMarkdown}
		_, _ = a.bot.Send(c.Sender(), msgCustomPrompt, opts)
		return nil
	})

	a.bot.Handle(tb.OnText, func(c tb.Context) error {
		userID := strconv.FormatInt(c.Sender().ID, 10)
		if _, ok := a.cache.Get(userID); ok {
			input := c.Message().Text
			duration, err := time.ParseDuration(input)
			if err != nil {
				opts := &tb.SendOptions{ParseMode: tb.ModeMarkdown}
				_, _ = a.bot.Send(c.Sender(), msgInvalidDuration, opts)
				return nil
			}
			a.cache.Delete(userID)
			return a.startFocusSession(c, duration)
		}
		return nil
	})
}

func (a *API) startFocusSession(c tb.Context, duration time.Duration) error {
	const op = "API.startFocusSession"
	log := a.logger.With("operation", op)

	vendorID := strconv.FormatInt(c.Sender().ID, 10)
	req := dto.CreateFocusSessionRequest{
		VendorID: vendorID,
		Duration: duration,
	}

	if err := a.services.FocusSessionService.Create(context.Background(), req); err != nil {
		log.Error(context.Background(), "failed to create focus session", err)
		return apperrors.NewInternal().WithDescriptionAndCause("failed to create focus session", err)
	}

	opts := &tb.SendOptions{ParseMode: tb.ModeMarkdown}
	confirmationMsg := msgSessionStarted + "`" + duration.String() + "`"

	if _, err := a.bot.Send(c.Sender(), confirmationMsg, opts); err != nil {
		log.Error(context.Background(), "failed to send focus session start confirmation", err)
		return apperrors.NewInternal().WithDescriptionAndCause("failed to send focus session confirmation", err)
	}

	return nil
}
