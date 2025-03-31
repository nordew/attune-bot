package models

import (
	"attune/pkg/apperrors"
	"time"
)

type FocusSession struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Quality   int       `json:"quality"`
	StartedAt time.Time `json:"startedAt"`
	EndedAt   time.Time `json:"endedAt"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func NewFocusSession(
	userID string,
	Duration time.Duration,
) (FocusSession, error) {
	if Duration <= 0 {
		return FocusSession{}, apperrors.NewBadRequest().WithDescription("duration must be greater than 0")
	}

	now := time.Now()
	return FocusSession{
		ID:        userID,
		UserID:    userID,
		StartedAt: now,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (fs *FocusSession) UpdateQuality(quality int) error {
	if quality < 0 || quality > 100 {
		return apperrors.NewBadRequest().WithDescription("quality must be between 0 and 100")
	}

	fs.Quality = quality
	fs.UpdatedAt = time.Now()

	return nil
}
