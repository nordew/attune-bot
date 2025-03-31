package notifier

import (
	"attune/pkg/apperrors"
	"context"
	"log"

	"golang.org/x/sync/errgroup"
)

var (
	ErrNotificationFailed           = "notification failed"
	ErrOneOrMoreNotificationsFailed = "one or more notifications failed"
)

type Impl struct {
}

func NewNotifier() Notifier {
	return &Impl{}
}

func (n *Impl) Send(ctx context.Context, notification Notification) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	log.Printf("Sending notification to recipient '%s': Title: %s, Message: %s",
		notification.RecipientID, notification.Title, notification.Message)

	return nil
}

func (n *Impl) SendBatch(ctx context.Context, notifications []Notification) error {
	if len(notifications) == 0 {
		return nil
	}

	g, ctx := errgroup.WithContext(ctx)
	for _, notification := range notifications {
		notification := notification

		g.Go(func() error {
			if err := n.Send(ctx, notification); err != nil {
				return apperrors.NewInternal().WithDescriptionAndCause(ErrNotificationFailed, err)
			}

			return nil
		})

	}

	if err := g.Wait(); err != nil {
		return apperrors.NewInternal().WithDescriptionAndCause(ErrOneOrMoreNotificationsFailed, err)
	}

	return nil
}
