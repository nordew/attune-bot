package telegram

import (
	"attune/pkg/apperrors"
	tb "gopkg.in/telebot.v4"
)

const (
	keyFocus15     = "focus_15"
	keyFocus30     = "focus_30"
	keyFocus60     = "focus_60"
	keyFocusCustom = "focus_custom"
)

var (
	msgFocusSessionPrompt = "🔔 *Select your focus session duration:*"
)

// SendFocusSessionMenu sends a menu to the user with options for focus session durations and a custom option
func (a *API) SendFocusSessionMenu(c tb.Context, customPrompt string) error {
	btn15 := tb.InlineButton{Unique: keyFocus15, Text: "15 min"}
	btn30 := tb.InlineButton{Unique: keyFocus30, Text: "30 min"}
	btn60 := tb.InlineButton{Unique: keyFocus60, Text: "60 min"}
	btnCustom := tb.InlineButton{Unique: keyFocusCustom, Text: "Custom 📝"}

	inlineKeys := [][]tb.InlineButton{
		{btn15, btn30},
		{btn60, btnCustom},
	}

	markup := &tb.ReplyMarkup{InlineKeyboard: inlineKeys}
	opts := &tb.SendOptions{
		ParseMode:   tb.ModeMarkdown,
		ReplyMarkup: markup,
	}

	if customPrompt != "" {
		msgFocusSessionPrompt = customPrompt
	}

	if _, err := a.bot.Send(c.Sender(), msgFocusSessionPrompt, opts); err != nil {
		return apperrors.NewInternal().WithDescriptionAndCause(ErrMsgFocusSessionMenu, err)
	}

	return nil
}
