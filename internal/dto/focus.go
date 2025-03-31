package dto

import "time"

type UpdateFocusRequestType string

const (
	UpdateFocusRequestTypePause  UpdateFocusRequestType = "pause"
	UpdateFocusRequestTypeResume UpdateFocusRequestType = "resume"
	UpdateFocusRequestTypeStop   UpdateFocusRequestType = "stop"
)

type CreateFocusSessionRequest struct {
	VendorID string        `json:"vendorId"`
	Duration time.Duration `json:"duration"`
}

type UpdateFocusRequest struct {
	ID      string                 `json:"id"`
	Type    UpdateFocusRequestType `json:"type"`
	Quality int                    `json:"quality"`
}
