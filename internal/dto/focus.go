package dto

import (
	"attune/internal/models"
	"time"
)

type UpdateFocusRequestType string

const (
	UpdateFocusRequestTypePause   UpdateFocusRequestType = "Pause"
	UpdateFocusRequestTypeResume  UpdateFocusRequestType = "Resume"
	UpdateFocusRequestTypeStop    UpdateFocusRequestType = "Stop"
	UpdateFocusRequestTypeQuality UpdateFocusRequestType = "Quality"
)

type CreateFocusSessionRequest struct {
	VendorID string        `json:"vendorId"`
	Duration time.Duration `json:"duration"`
}

type UpdateFocusRequest struct {
	VendorID string                    `json:"id"`
	Type     UpdateFocusRequestType    `json:"type"`
	Status   models.FocusSessionStatus `json:"status"`
	Quality  int                       `json:"quality"`
}
