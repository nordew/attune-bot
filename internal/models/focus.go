package models

import (
	"attune/pkg/apperrors"
	"github.com/google/uuid"
	"time"
)

type FocusSessionStatus string

const (
	FocusSessionStatusActive    FocusSessionStatus = "active"
	FocusSessionStatusCompleted FocusSessionStatus = "completed"
	FocusSessionStatusStopped   FocusSessionStatus = "stopped"
)

var (
	ErrInvalidDuration = "Duration must be between 1 minute and 24 hours"
	ErrInvalidQuality  = "Invalid quality value, must be between 0 and 10"
)

type FocusSession struct {
	ID        string             `json:"id"`
	UserID    string             `json:"userId"`
	VendorID  string             `json:"vendorId"`
	Status    FocusSessionStatus `json:"status"`
	Quality   int                `json:"quality"`
	StartedAt time.Time          `json:"startedAt"`
	EndedAt   time.Time          `json:"endedAt"`
	CreatedAt time.Time          `json:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt"`
}

func NewFocusSession(
	userID string,
	Duration time.Duration,
) (FocusSession, error) {
	if Duration <= 0 || Duration > time.Hour*24 || Duration < time.Minute {
		return FocusSession{}, apperrors.NewBadRequest().WithDescription(ErrInvalidDuration)
	}

	now := time.Now()
	return FocusSession{
		ID:        uuid.NewString(),
		UserID:    userID,
		Status:    FocusSessionStatusActive,
		StartedAt: now,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (fs *FocusSession) UpdateQuality(quality int) error {
	if quality < 0 || quality > 10 {
		return apperrors.NewBadRequest().WithDescription(ErrInvalidQuality)
	}

	fs.Quality = quality
	fs.UpdatedAt = time.Now()

	return nil
}
