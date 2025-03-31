package api

import (
	"attune/internal/models"
	"context"
)

type ExternalAPI interface {
	Start(ctx context.Context) error
	SendMessage(ctx context.Context, message models.Message) error
}
