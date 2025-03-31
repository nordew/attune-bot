package models

import (
	"attune/internal/consts"
	"attune/pkg/apperrors"
	"github.com/google/uuid"
	"time"
	"unicode/utf8"
)

type Vendor string

const (
	VendorTelegram Vendor = "telegram"
)

type User struct {
	ID             string    `json:"id"`
	VendorID       string    `json:"vendorId"`
	VendorType     Vendor    `json:"vendorType"`
	Name           string    `json:"name"`
	LastActivityAt time.Time `json:"lastActivityAt"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func NewUser(
	id string,
	vendorID string,
	vendorType Vendor,
	name string,
) (User, error) {
	_, err := uuid.Parse(id)
	if err != nil {
		return User{}, apperrors.NewBadRequest().WithDescription("invalid user id")
	}
	if utf8.RuneCountInString(name) > consts.MaxNameLength {
		return User{}, apperrors.NewBadRequest().WithDescription("name is too long")
	}

	now := time.Now()

	return User{
		ID:             id,
		VendorID:       vendorID,
		VendorType:     vendorType,
		Name:           name,
		LastActivityAt: now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

func (u *User) UpdateLastActivity() {
	u.LastActivityAt = time.Now()
}

func (u *User) UpdateName(name string) error {
	if utf8.RuneCountInString(name) > consts.MaxNameLength {
		return apperrors.NewBadRequest().WithDescription("name is too long")
	}

	u.Name = name
	return nil
}
