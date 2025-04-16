package domain

type CustomerProfileResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Address   string `json:"address,omitempty"`
	LoyaltyID string `json:"loyalty_id,omitempty"`
	CreatedAt string `json:"created_at"`
}

type DriverProfileResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Phone         string `json:"phone"`
	LicenseNumber string `json:"license_number,omitempty"`
	LicenseExpiry string `json:"license_expiry,omitempty"`
	VehicleInfo   string `json:"vehicle_info,omitempty"`
	IsVerified    bool   `json:"is_verified"`
	CreatedAt     string `json:"created_at"`
}

type CustomerProfileUpdate struct {
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Address   string `json:"address,omitempty"`
}

type DriverProfileUpdate struct {
	FirstName     string `json:"first_name,omitempty"`
	LastName      string `json:"last_name,omitempty"`
	Phone         string `json:"phone,omitempty"`
	VehicleInfo   string `json:"vehicle_info,omitempty"`
	LicenseNumber string `json:"license_number,omitempty"`
	LicenseExpiry string `json:"license_expiry,omitempty"`
}
