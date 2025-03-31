package dto

import "attune/internal/models"

type CreateUserRequest struct {
	Name       string        `json:"name"`
	VendorType models.Vendor `json:"vendorType"`
	VendorID   string        `json:"vendorId"`
}
