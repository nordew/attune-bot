package models

import "time"

type UserSettings struct {
	ID               string    `json:"id"`
	UserID           string    `json:"userId"`
	SentDailyStatsAt time.Time `json:"sentDailyStatsAt"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

func NewUserSettings(userID string, sentDailyStatsAt time.Time) (UserSettings, error) {
	now := time.Now()

	return UserSettings{
		ID:               userID,
		UserID:           userID,
		SentDailyStatsAt: sentDailyStatsAt,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

func (us *UserSettings) UpdateSentDailyStatsAt(sentDailyStatsAt time.Time) {
	us.SentDailyStatsAt = sentDailyStatsAt
	us.UpdatedAt = time.Now()
}
