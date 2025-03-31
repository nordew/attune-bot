package service

import (
	"attune/internal/models"
	"attune/internal/storage"
	"attune/pkg/logger"
	"context"
	"time"
)

type FocusWorker interface {
	Enqueue(session models.FocusSession)
	Start(ctx context.Context)
	Stop()
}

type focusWorker struct {
	storages  storage.Storages
	logger    logger.Logger
	sessionCh chan models.FocusSession
	doneCh    chan struct{}
}

func NewFocusWorker(storages storage.Storages, logger logger.Logger) FocusWorker {
	return &focusWorker{
		storages:  storages,
		logger:    logger,
		sessionCh: make(chan models.FocusSession, 100),
		doneCh:    make(chan struct{}),
	}
}

func (w *focusWorker) Enqueue(session models.FocusSession) {
	select {
	case w.sessionCh <- session:
	default:
		w.logger.Error(context.Background(), "failed to enqueue focus session: channel is full", nil)
	}
}

func (w *focusWorker) Start(ctx context.Context) {
	w.logger.Info(ctx, "FocusWorker started")

	for {
		select {
		case session := <-w.sessionCh:
			go w.processSession(ctx, session)
		case <-ctx.Done():
			w.logger.Info(ctx, "FocusWorker stopping due to context cancellation")
			return
		case <-w.doneCh:
			w.logger.Info(ctx, "FocusWorker stopping due to done signal")
			return
		}
	}
}

func (w *focusWorker) processSession(ctx context.Context, session models.FocusSession) {
	const op = "FocusWorker.processSession"
	log := w.logger.With("operation", op, "sessionID", session.ID)

	if err := sendNotification(session); err != nil {
		log.Error(ctx, "failed to send notification", err)
	} else {
		log.Info(ctx, "notification sent")
	}

	if err := w.storages.FocusSession.Create(ctx, session); err != nil {
		log.Error(ctx, "failed to save session to database", err)
	} else {
		log.Info(ctx, "session saved to database")
	}
}

func (w *focusWorker) Stop() {
	close(w.doneCh)
	w.logger.Info(context.Background(), "FocusWorker stopped")
}

func sendNotification(session models.FocusSession) error {
	time.Sleep(100 * time.Millisecond)
	return nil
}
