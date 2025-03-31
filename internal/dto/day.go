package dto

type CreateDayRecordRequest struct {
	UserID  string `json:"userId"`
	Quality int    `json:"quality"`
	Mood    string `json:"mood"`
}
