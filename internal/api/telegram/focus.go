package telegram

import (
	"context"
	"strconv"
	"time"

	"attune/internal/dto"
	"attune/internal/models"
	"attune/pkg/apperrors"

	tb "gopkg.in/telebot.v4"
)

const (
	prefixRateFocusQuality = "rate_focus_quality_"
	prefixCustomDuration   = "custom_"

	keyFocusPause  = "focus_pause"
	keyFocusResume = "focus_resume"
	keyFocusStop   = "focus_stop"

	msgCustomPrompt    = "⌨️ _Please enter your desired duration_\n(e.g., `45m` for 45 minutes)."
	msgSessionStarted  = "✅ *Your focus session has started!*\nDuration: "
	msgInvalidDuration = "❌ *Invalid duration format.*\nPlease try again (e.g., `45m`)."
	msgSessionPaused   = "⏸️ *Your focus session is paused.*"
	msgSessionResumed  = "▶️ *Your focus session has resumed!*"
	msgSessionStopped  = "🛑 *Your focus session has been stopped.*"
	msgFocusQuality    = "How was your focus quality? Please select a value between 1 and 10."
	msgSessionFinished = "✅ *Your focus session has finished!*"

	msgInvalidRating      = "Invalid rating. Please send a number between 1 and 10."
	msgRatingOutOfRange   = "Rating must be between 1 and 10."
	msgThankRating        = "Thank you for rating your focus quality!"
	msgFailedUpdateRating = "Failed to update quality rating."
)

var (
	ErrMsgFocusSessionMenu         = "failed to send focus session menu"
	ErrMsgFocusSessionConfirmation = "failed to send focus session confirmation"
	ErrMsgFocusSessionUpdate       = "failed to update focus session"
	ErrMsgSendConfirmation         = "failed to send confirmation"
	ErrMsgInvalidVendorID          = "invalid vendor ID"
	ErrMsgFinishConfirmation       = "failed to send finish confirmation"
	ErrMsgFocusQualityPrompt       = "failed to send focus quality prompt"
)

func (a *API) registerFocusSessionCallbacks() {
	predefined := map[string]time.Duration{
		keyFocus15: 15 * time.Minute,
		keyFocus30: 30 * time.Minute,
		keyFocus60: 60 * time.Minute,
	}
	for key, duration := range predefined {
		k, d := key, duration
		a.bot.Handle(&tb.InlineButton{Unique: k}, func(c tb.Context) error {
			err := a.startFocusSession(c, d)
			if err != nil {
				a.logger.Error(context.Background(), "Error starting focus session", err, "duration", d.String(), "user", c.Sender().ID)
			}
			return err
		})
	}

	a.bot.Handle(&tb.InlineButton{Unique: keyFocusCustom}, func(c tb.Context) error {
		userID := strconv.FormatInt(c.Sender().ID, 10)
		a.cache.Set(prefixCustomDuration+userID, true)

		_, err := a.bot.Send(c.Sender(), msgCustomPrompt, &tb.SendOptions{ParseMode: tb.ModeMarkdown})
		if err != nil {
			a.logger.Error(context.Background(), "Failed to send custom duration prompt", err, "user", c.Sender().ID)
		}
		return nil
	})

	a.bot.Handle(tb.OnText, func(c tb.Context) error {
		userID := strconv.FormatInt(c.Sender().ID, 10)

		if _, ok := a.cache.Get(prefixCustomDuration + userID); ok {
			input := c.Message().Text

			duration, err := time.ParseDuration(input)
			if err != nil {
				a.logger.Error(context.Background(), "Invalid duration format", err, "input", input, "user", c.Sender().ID)
				_, _ = a.bot.Send(c.Sender(), msgInvalidDuration, &tb.SendOptions{ParseMode: tb.ModeMarkdown})
				return nil
			}

			a.cache.Delete(prefixCustomDuration + userID)
			err = a.startFocusSession(c, duration)
			if err != nil {
				return err
			}

			return nil
		} else if focusSessionStatus, ok := a.cache.Get(prefixRateFocusQuality + userID); ok { // Check if user is rating focus quality
			focusSessionStatus, ok := focusSessionStatus.(models.FocusSessionStatus)
			if !ok {
				a.logger.Error(context.Background(), "Failed to type cast focus session status", nil, "user", c.Sender().ID)
				return nil
			}

			input := c.Message().Text

			rating, err := strconv.Atoi(input)
			if err != nil {
				_, _ = a.bot.Send(c.Sender(), msgInvalidRating, &tb.SendOptions{ParseMode: tb.ModeMarkdown})
				return nil
			}

			if rating < 1 || rating > 10 {
				a.logger.Info(context.Background(), "Rating out of range", "rating", rating, "user", c.Sender().ID)
				_, _ = a.bot.Send(c.Sender(), msgRatingOutOfRange, &tb.SendOptions{ParseMode: tb.ModeMarkdown})
				return nil
			}

			updateDTO := dto.UpdateFocusRequest{
				VendorID: userID,
				Type:     dto.UpdateFocusRequestTypeQuality,
				Quality:  rating,
				Status:   focusSessionStatus,
			}
			if err := a.services.FocusSessionService.Update(context.Background(), updateDTO); err != nil {
				_, _ = a.bot.Send(c.Sender(), msgFailedUpdateRating, &tb.SendOptions{ParseMode: tb.ModeMarkdown})
				return err
			}

			a.cache.Delete(prefixRateFocusQuality + userID)

			_, _ = a.bot.Send(c.Sender(), msgThankRating, &tb.SendOptions{ParseMode: tb.ModeMarkdown})

			if err := a.SendFocusSessionMenu(c, ""); err != nil {
				a.logger.Error(context.Background(), "Failed to send focus session menu after rating", err, "user", c.Sender().ID)
				return apperrors.NewInternal().WithDescriptionAndCause(ErrMsgFocusSessionMenu, err)
			}

			return nil
		}
		return nil
	})

	controls := []struct {
		key     string
		handler func(c tb.Context) error
	}{
		{key: keyFocusPause, handler: a.pauseFocusSession},
		{key: keyFocusResume, handler: a.resumeFocusSession},
		{key: keyFocusStop, handler: a.stopFocusSession},
	}
	for _, ctrl := range controls {
		ctrlCopy := ctrl
		a.bot.Handle(&tb.InlineButton{Unique: ctrlCopy.key}, func(c tb.Context) error {
			err := ctrlCopy.handler(c)
			if err != nil {
				return err
			}

			return nil
		})
	}
}

func (a *API) createFocusSession(c tb.Context) error {
	if err := a.SendFocusSessionMenu(c, ""); err != nil {
		a.logger.Error(context.Background(), "Failed to send focus session menu", err, "user", c.Sender().ID)
		return apperrors.NewInternal().WithDescriptionAndCause(ErrMsgFocusSessionMenu, err)
	}

	return nil
}

func (a *API) startFocusSession(c tb.Context, duration time.Duration) error {
	vendorID := strconv.FormatInt(c.Sender().ID, 10)

	req := dto.CreateFocusSessionRequest{
		VendorID: vendorID,
		Duration: duration,
	}

	if err := a.services.FocusSessionService.Create(context.Background(), req); err != nil {
		return err
	}

	confirmationMsg := msgSessionStarted + "`" + duration.String() + "`"
	opts := &tb.SendOptions{ParseMode: tb.ModeMarkdown}

	sentMsg, err := a.bot.Send(c.Sender(), confirmationMsg, opts)
	if err != nil {
		return err
	}

	if err := a.editFocusSessionMessage(sentMsg, focusStateStarted, false); err != nil {
		return err
	}

	go a.cache.Set("FocusMsg:"+vendorID, sentMsg)

	return nil
}

func (a *API) updateFocusSession(
	c tb.Context,
	updateType dto.UpdateFocusRequestType,
) error {
	vendorID := strconv.FormatInt(c.Sender().ID, 10)

	updateDTO := dto.UpdateFocusRequest{
		VendorID: vendorID,
		Type:     updateType,
	}
	if err := a.services.FocusSessionService.Update(context.Background(), updateDTO); err != nil {
		return apperrors.NewInternal().WithDescriptionAndCause(ErrMsgFocusSessionUpdate, err)
	}

	return nil
}

func (a *API) pauseFocusSession(c tb.Context) error {
	if err := a.updateFocusSession(c, dto.UpdateFocusRequestTypePause); err != nil {
		return err
	}

	userID := strconv.FormatInt(c.Sender().ID, 10)
	rawMsg, ok := a.cache.Get("FocusMsg:" + userID)
	if !ok {
		return nil
	}
	msg, ok := rawMsg.(*tb.Message)
	if !ok {
		return nil
	}

	return a.editFocusSessionMessage(msg, focusStatePaused, true)
}

func (a *API) resumeFocusSession(c tb.Context) error {
	if err := a.updateFocusSession(c, dto.UpdateFocusRequestTypeResume); err != nil {
		return err
	}

	userID := strconv.FormatInt(c.Sender().ID, 10)
	rawMsg, ok := a.cache.Get("FocusMsg:" + userID)
	if !ok {
		return nil
	}
	msg, ok := rawMsg.(*tb.Message)
	if !ok {
		return nil
	}

	return a.editFocusSessionMessage(msg, focusStateResumed, false)
}

func (a *API) stopFocusSession(c tb.Context) error {
	return a.updateFocusSession(c, dto.UpdateFocusRequestTypeStop)
}

func (a *API) finishFocusSession(
	_ context.Context,
	vendorID string,
	focusSessionStatus models.FocusSessionStatus,
) error {
	chatID, err := strconv.ParseInt(vendorID, 10, 64)
	if err != nil {
		return apperrors.NewInternal().WithDescriptionAndCause(ErrMsgInvalidVendorID, err)
	}

	vendorChat := &tb.Chat{ID: chatID}
	opts := &tb.SendOptions{ParseMode: tb.ModeMarkdown}
	if _, err := a.bot.Send(vendorChat, msgSessionFinished, opts); err != nil {
		return apperrors.NewInternal().WithDescriptionAndCause(ErrMsgFinishConfirmation, err)
	}

	if _, err := a.bot.Send(vendorChat, msgFocusQuality, opts); err != nil {
		return apperrors.NewInternal().WithDescriptionAndCause(ErrMsgFocusQualityPrompt, err)
	}

	a.cache.Set(prefixRateFocusQuality+vendorID, focusSessionStatus)

	return nil
}
