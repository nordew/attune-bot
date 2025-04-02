package api

import (
	"attune/internal/models"
	"context"
)

type TriggerType string

type Trigger struct {
	VendorID           string
	Type               TriggerType
	FocusSessionStatus models.FocusSessionStatus
}

const (
	TriggerTypeFinishSession TriggerType = "finish_session"
)

type ExternalAPI interface {
	Start(ctx context.Context) error
	Trigger(ctx context.Context, vendorID string, trigger Trigger) error
	SendMessage(ctx context.Context, message models.Message) error
}
