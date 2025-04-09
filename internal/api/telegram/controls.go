package telegram

import (
	"fmt"

	tb "gopkg.in/telebot.v4"
)

type focusStateType string

const (
	focusStateStarted focusStateType = "started"
	focusStatePaused  focusStateType = "paused"
	focusStateResumed focusStateType = "resumed"
	focusStateStopped focusStateType = "stopped"
)

const (
	msgControlFocusSession = "Control your focus session:"
)

func (a *API) editFocusSessionMessage(msg *tb.Message, state focusStateType, paused bool) error {
	var stateIcon, stateText string
	switch state {
	case focusStateStarted:
		stateIcon = "✅"
		stateText = "Your focus session has started!"
	case focusStatePaused:
		stateIcon = "⏸"
		stateText = "Your focus session is paused."
	case focusStateResumed:
		stateIcon = "▶️"
		stateText = "Your focus session has resumed."
	case focusStateStopped:
		stateIcon = "⏹"
		stateText = "Your focus session has stopped."
	default:
		stateIcon = "ℹ️"
		stateText = "Focus session updated."
	}

	var actionButton tb.InlineButton
	if paused {
		actionButton = tb.InlineButton{Unique: keyFocusResume, Text: "Resume"}
	} else {
		actionButton = tb.InlineButton{Unique: keyFocusPause, Text: "Pause"}
	}

	stopButton := tb.InlineButton{Unique: keyFocusStop, Text: "Stop"}

	markup := &tb.ReplyMarkup{
		InlineKeyboard: [][]tb.InlineButton{
			{actionButton, stopButton},
		},
	}

	updatedText := fmt.Sprintf(
		"%s %s\n\nControl your focus session:",
		stateIcon,
		stateText,
	)

	opts := &tb.SendOptions{
		ParseMode:   tb.ModeMarkdown,
		ReplyMarkup: markup,
	}

	_, err := a.bot.Edit(msg, updatedText, opts)
	return err
}
