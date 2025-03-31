package service

import (
	"attune/internal/models"
	"attune/internal/storage"
	"attune/pkg/apperrors"
	"attune/pkg/cache"
	"time"
)

type FocusSessionManager interface {
	Start(session models.FocusSession)
	Pause(sessionID string) error
	Resume(sessionID string) error
	Stop(sessionID string) error
}

type focusSessionManager struct {
	storages storage.Storages
	cache    cache.Cache
	worker   FocusWorker
}

type sessionData struct {
	session  models.FocusSession
	timer    *time.Timer
	paused   bool
	pauseCh  chan struct{}
	resumeCh chan struct{}
	stopCh   chan struct{}
}

func NewFocusSessionManager(
	storages storage.Storages,
	worker FocusWorker,
	c cache.Cache) FocusSessionManager {
	return &focusSessionManager{
		storages: storages,
		worker:   worker,
		cache:    c,
	}
}

func (m *focusSessionManager) Start(session models.FocusSession) {
	var duration time.Duration

	if session.EndedAt.IsZero() {
		duration = session.EndedAt.Sub(session.StartedAt)
	} else {
		duration = time.Until(session.EndedAt)
	}

	data := &sessionData{
		session:  session,
		timer:    time.NewTimer(duration),
		paused:   false,
		pauseCh:  make(chan struct{}),
		resumeCh: make(chan struct{}),
		stopCh:   make(chan struct{}),
	}

	m.cache.Set(session.ID, data)

	go m.track(data)
}

func (m *focusSessionManager) Pause(sessionID string) error {
	v, ok := m.cache.Get(sessionID)
	if !ok {
		return apperrors.NewNotFound().WithDescription("session not found")
	}
	data, ok := v.(*sessionData)
	if !ok {
		return apperrors.NewInternal().WithDescription("invalid session data type")
	}

	if data.paused {
		return apperrors.NewBadRequest().WithDescription("session is already paused")
	}

	data.paused = true

	close(data.pauseCh)

	return nil
}

func (m *focusSessionManager) Resume(sessionID string) error {
	v, ok := m.cache.Get(sessionID)
	if !ok {
		return apperrors.NewNotFound().WithDescription("session not found")
	}
	data, ok := v.(*sessionData)
	if !ok {
		return apperrors.NewInternal().WithDescription("invalid session data type")
	}

	if !data.paused {
		return apperrors.NewBadRequest().WithDescription("session is not paused")
	}

	remaining := time.Until(data.session.EndedAt)

	data.paused = false
	data.timer = time.NewTimer(remaining)
	data.pauseCh = make(chan struct{})
	data.resumeCh = make(chan struct{})

	go m.track(data)

	return nil
}

func (m *focusSessionManager) track(data *sessionData) {
	select {
	case <-data.timer.C:
		data.session.EndedAt = time.Now()

		m.worker.Enqueue(data.session)

		m.cache.Delete(data.session.ID)
	case <-data.pauseCh:
		if !data.timer.Stop() {
			<-data.timer.C
		}
	case <-data.stopCh:
		if !data.timer.Stop() {
			<-data.timer.C
		}
	}
}

func (m *focusSessionManager) Stop(sessionID string) error {
	v, ok := m.cache.Get(sessionID)
	if !ok {
		return apperrors.NewNotFound().WithDescription("session not found")
	}
	data, ok := v.(*sessionData)
	if !ok {
		return apperrors.NewInternal().WithDescription("invalid session data type")
	}

	close(data.stopCh)

	m.cache.Delete(sessionID)

	return nil
}
