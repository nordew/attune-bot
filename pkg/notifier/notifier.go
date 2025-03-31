package notifier

import (
	"attune/pkg/apperrors"
	"context"
	"github.com/google/uuid"
	"time"
)

const (
	ErrRecipientIDRequired = "recipientId is required"
	ErrTitleRequired       = "title is required and must be less than 255 characters"
	ErrMessageRequired     = "message is required and must be less than 1024 characters"
)

type NotificationType string

const (
	NotificationTypePush  NotificationType = "push"
	NotificationTypeInApp NotificationType = "in_app"
)

type Notification struct {
	ID          string           `json:"id"`
	Type        NotificationType `json:"type"`
	RecipientID string           `json:"recipientId"`
	Title       string           `json:"title"`
	Message     string           `json:"message"`
	IssuedAt    time.Time        `json:"issuedAt"`
}

func NewNotification(
	id string,
	notificationType NotificationType,
	recipientID, title, message string,
) (Notification, error) {
	if _, err := uuid.Parse(id); err != nil {
		return Notification{}, apperrors.NewBadRequest().WithDescription("invalid notification id")
	}
	if recipientID == "" {
		return Notification{}, apperrors.NewBadRequest().WithDescription(ErrRecipientIDRequired)
	}
	if title == "" || len(title) > 255 {
		return Notification{}, apperrors.NewBadRequest().WithDescription(ErrTitleRequired)
	}
	if message == "" || len(message) > 1024 {
		return Notification{}, apperrors.NewBadRequest().WithDescription(ErrMessageRequired)
	}

	return Notification{
		ID:          id,
		RecipientID: recipientID,
		Title:       title,
		Type:        notificationType,
		Message:     message,
		IssuedAt:    time.Now(),
	}, nil
}

type Notifier interface {
	Send(ctx context.Context, notification Notification) error
	SendBatch(ctx context.Context, notifications []Notification) error
}
