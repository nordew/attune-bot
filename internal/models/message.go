package models

import (
	"attune/pkg/apperrors"
	"time"
)

type MessageType string

const (
	MessageTypeText  MessageType = "text"
	MessageTypeImage MessageType = "image"
	MessageTypeVideo MessageType = "video"
	MessageTypeAudio MessageType = "audio"
	MessageTypeFile  MessageType = "file"
	MessageTypeLink  MessageType = "link"
)

type Message struct {
	VendorID string      `json:"vendorId"`
	Type     MessageType `json:"type"`
	Text     string      `json:"message"`
	IssuedAt time.Time   `json:"issuedAt"`
}

func NewMessage(
	vendorID string,
	messageType MessageType,
	text string,
) (Message, error) {
	if vendorID == "" {
		return Message{}, apperrors.NewBadRequest().WithDescription("vendorId is required")
	}
	if text == "" {
		return Message{}, apperrors.NewBadRequest().WithDescription("text is required")
	}

	return Message{
		VendorID: vendorID,
		Type:     messageType,
		Text:     text,
		IssuedAt: time.Now(),
	}, nil
}
