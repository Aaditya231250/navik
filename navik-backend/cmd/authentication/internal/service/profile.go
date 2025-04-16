// internal/service/profile_service.go
package service

import (
	"context"

	"authentication/internal/domain"
	"authentication/internal/repository"
)

type ProfileService struct {
	userRepo *repository.UserRepository
}

func NewProfileService(userRepo *repository.UserRepository) *ProfileService {
	return &ProfileService{
		userRepo: userRepo,
	}
}

// GetCustomerProfile retrieves a customer's profile
func (s *ProfileService) GetCustomerProfile(ctx context.Context, userID string) (*domain.CustomerProfileResponse, error) {
	customer, err := s.userRepo.GetCustomerByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &domain.CustomerProfileResponse{
		ID:        customer.ID,
		Email:     customer.Email,
		FirstName: customer.FirstName,
		LastName:  customer.LastName,
		Phone:     customer.Phone,
		Address:   customer.Address,
		LoyaltyID: customer.LoyaltyID,
		CreatedAt: customer.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// UpdateCustomerProfile updates a customer's profile
func (s *ProfileService) UpdateCustomerProfile(ctx context.Context, userID string, update domain.CustomerProfileUpdate) error {
	return s.userRepo.UpdateCustomerProfile(ctx, userID, update)
}

// GetDriverProfile retrieves a driver's profile
func (s *ProfileService) GetDriverProfile(ctx context.Context, userID string) (*domain.DriverProfileResponse, error) {
	driver, err := s.userRepo.GetDriverByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	licenseExpiry := ""
	if !driver.LicenseExpiry.IsZero() {
		licenseExpiry = driver.LicenseExpiry.Format("2006-01-02T15:04:05Z07:00")
	}

	return &domain.DriverProfileResponse{
		ID:            driver.ID,
		Email:         driver.Email,
		FirstName:     driver.FirstName,
		LastName:      driver.LastName,
		Phone:         driver.Phone,
		LicenseNumber: driver.LicenseNumber,
		LicenseExpiry: licenseExpiry,
		VehicleInfo:   driver.VehicleInfo,
		IsVerified:    driver.IsVerified,
		CreatedAt:     driver.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// UpdateDriverProfile updates a driver's profile
func (s *ProfileService) UpdateDriverProfile(ctx context.Context, userID string, update domain.DriverProfileUpdate) error {
	return s.userRepo.UpdateDriverProfile(ctx, userID, update)
}
