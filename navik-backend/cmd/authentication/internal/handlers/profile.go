// internal/handlers/profile_handler.go
package handlers

import (
    "encoding/json"
    "net/http"
    
    "authentication/internal/domain"
    "authentication/internal/service"
)

type ProfileHandler struct {
    profileService *service.ProfileService
}

func NewProfileHandler(profileService *service.ProfileService) *ProfileHandler {
    return &ProfileHandler{
        profileService: profileService,
    }
}

// GetCustomerProfile handles requests to get a customer profile
func (h *ProfileHandler) GetCustomerProfile(w http.ResponseWriter, r *http.Request) {
    // Get user ID from request context (set by auth middleware)
    userID, ok := r.Context().Value("user_id").(string)
    if !ok {
        respondWithError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }
    
    // Get profile from service
    profile, err := h.profileService.GetCustomerProfile(r.Context(), userID)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Failed to retrieve profile")
        return
    }
    
    respondWithJSON(w, http.StatusOK, profile)
}

// UpdateCustomerProfile handles requests to update a customer profile
func (h *ProfileHandler) UpdateCustomerProfile(w http.ResponseWriter, r *http.Request) {
    // Get user ID from request context
    userID, ok := r.Context().Value("user_id").(string)
    if !ok {
        respondWithError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }
    
    // Decode request body
    var update domain.CustomerProfileUpdate
    if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid request payload")
        return
    }
    
    // Update profile
    if err := h.profileService.UpdateCustomerProfile(r.Context(), userID, update); err != nil {
        respondWithError(w, http.StatusInternalServerError, "Failed to update profile")
        return
    }
    
    // Get updated profile
    profile, err := h.profileService.GetCustomerProfile(r.Context(), userID)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Profile updated but failed to retrieve updated data")
        return
    }
    
    respondWithJSON(w, http.StatusOK, profile)
}

// GetDriverProfile handles requests to get a driver profile
func (h *ProfileHandler) GetDriverProfile(w http.ResponseWriter, r *http.Request) {
    // Get user ID from request context
    userID, ok := r.Context().Value("user_id").(string)
    if !ok {
        respondWithError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }
    
    // Get profile from service
    profile, err := h.profileService.GetDriverProfile(r.Context(), userID)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Failed to retrieve profile")
        return
    }
    
    respondWithJSON(w, http.StatusOK, profile)
}

// UpdateDriverProfile handles requests to update a driver profile
func (h *ProfileHandler) UpdateDriverProfile(w http.ResponseWriter, r *http.Request) {
    // Get user ID from request context
    userID, ok := r.Context().Value("user_id").(string)
    if !ok {
        respondWithError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }
    
    // Decode request body
    var update domain.DriverProfileUpdate
    if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid request payload")
        return
    }
    
    // Update profile
    if err := h.profileService.UpdateDriverProfile(r.Context(), userID, update); err != nil {
        respondWithError(w, http.StatusInternalServerError, "Failed to update profile")
        return
    }
    
    // Get updated profile
    profile, err := h.profileService.GetDriverProfile(r.Context(), userID)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Profile updated but failed to retrieve updated data")
        return
    }
    
    respondWithJSON(w, http.StatusOK, profile)
}
