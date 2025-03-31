package dto

import "time"

type CreateUserSettingsRequest struct {
	UserID           string    `json:"userId"`
	SentDailyStatsAt time.Time `json:"sentDailyStatsAt"`
}
