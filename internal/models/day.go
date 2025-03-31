package models

import (
	"attune/pkg/apperrors"
	"time"
)

type DayRecord struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Quality   int       `json:"quality"`
	Mood      string    `json:"mood"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func NewDayRecord(
	userID string,
	quality int,
	mood string,
) (DayRecord, error) {
	if quality < 0 || quality > 10 {
		return DayRecord{}, apperrors.NewBadRequest().WithDescription("quality must be between 0 and 10")
	}
	if len(mood) > 255 {
		return DayRecord{}, apperrors.NewBadRequest().WithDescription("mood is too long")
	}

	now := time.Now()
	return DayRecord{
		ID:        userID,
		UserID:    userID,
		Quality:   quality,
		Mood:      mood,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (dr *DayRecord) UpdateQuality(quality int) error {
	if quality < 0 || quality > 10 {
		return apperrors.NewBadRequest().WithDescription("quality must be between 0 and 10")
	}

	dr.Quality = quality
	dr.UpdatedAt = time.Now()
	return nil
}

func (dr *DayRecord) UpdateMood(mood string) error {
	if len(mood) > 255 {
		return apperrors.NewBadRequest().WithDescription("mood is too long")
	}

	dr.Mood = mood
	dr.UpdatedAt = time.Now()
	return nil
}
