package service

import (
	"attune/internal/api"
	"attune/internal/models"
	"attune/internal/storage"
	"attune/pkg/apperrors"
	"attune/pkg/cache"
	"sync"
	"time"
)

const (
	cacheTTLWindow = 3 * time.Minute
)

type FocusSessionManager interface {
	Start(session models.FocusSession, duration time.Duration)
	Pause(userID string) error
	Resume(userID string) error
	Stop(userID string) error
}

type focusSessionManager struct {
	storages storage.Storages
	cache    cache.Cache
	apiCh    chan<- api.Trigger
}

type sessionData struct {
	mu        sync.Mutex
	session   models.FocusSession
	timer     *time.Timer
	remaining time.Duration
	lastStart time.Time
	paused    bool
	pauseCh   chan struct{}
	stopCh    chan struct{}
}

func NewFocusSessionManager(
	storages storage.Storages,
	c cache.Cache,
	apiCh chan<- api.Trigger,
) FocusSessionManager {
	return &focusSessionManager{
		storages: storages,
		cache:    c,
		apiCh:    apiCh,
	}
}

func (m *focusSessionManager) Start(session models.FocusSession, duration time.Duration) {
	data := &sessionData{
		session:   session,
		timer:     time.NewTimer(duration),
		paused:    false,
		remaining: duration,
		lastStart: time.Now(),
		pauseCh:   make(chan struct{}),
		stopCh:    make(chan struct{}),
	}

	go m.cache.SetWithTTL(session.UserID, data, duration+cacheTTLWindow)
	go m.track(data)
}

func (m *focusSessionManager) Pause(userID string) error {
	v, ok := m.cache.Get(userID)
	if !ok {
		return apperrors.NewNotFound().WithDescription("session not found")
	}
	data, ok := v.(*sessionData)
	if !ok {
		return apperrors.NewInternal().WithDescription("invalid session data type")
	}

	data.mu.Lock()
	defer data.mu.Unlock()

	if data.paused {
		return apperrors.NewBadRequest().WithDescription("session is already paused")
	}

	elapsed := time.Since(data.lastStart)
	if elapsed > data.remaining {
		data.remaining = 0
	} else {
		data.remaining -= elapsed
	}

	data.paused = true

	close(data.pauseCh)

	return nil
}

func (m *focusSessionManager) Resume(userID string) error {
	v, ok := m.cache.Get(userID)
	if !ok {
		return apperrors.NewNotFound().WithDescription("session not found")
	}
	data, ok := v.(*sessionData)
	if !ok {
		return apperrors.NewInternal().WithDescription("invalid session data type")
	}

	data.mu.Lock()
	defer data.mu.Unlock()

	if !data.paused {
		return apperrors.NewBadRequest().WithDescription("session is not paused")
	}

	data.timer = time.NewTimer(data.remaining)
	data.lastStart = time.Now()
	data.paused = false
	data.pauseCh = make(chan struct{})

	go m.track(data)

	return nil
}

func (m *focusSessionManager) Stop(userID string) error {
	v, ok := m.cache.Get(userID)
	if !ok {
		return apperrors.NewNotFound().WithDescription("session not found")
	}
	data, ok := v.(*sessionData)
	if !ok {
		return apperrors.NewInternal().WithDescription("invalid session data type")
	}

	close(data.stopCh)
	go m.cache.Delete(userID)

	return nil
}

func (m *focusSessionManager) track(data *sessionData) {
	waitForTimerStop := func() {
		if !data.timer.Stop() {
			<-data.timer.C
		}
	}

	select {
	case <-data.timer.C:
		m.completeSession(data, models.FocusSessionStatusCompleted)
	case <-data.pauseCh:
		waitForTimerStop()
	case <-data.stopCh:
		waitForTimerStop()
		m.completeSession(data, models.FocusSessionStatusStopped)
	}
}

func (m *focusSessionManager) completeSession(data *sessionData, sessionStatus models.FocusSessionStatus) {
	data.session.EndedAt = time.Now()

	if data.session.VendorID != "" {
		go func() {
			m.apiCh <- api.Trigger{
				VendorID:           data.session.VendorID,
				Type:               api.TriggerTypeFinishSession,
				FocusSessionStatus: sessionStatus,
			}
		}()
	}

	go m.cache.Delete(data.session.UserID)
}
