package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"

	"authentication/internal/domain"
	"authentication/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
	validate    *validator.Validate
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validate:    validator.New(),
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.Registration

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		respondWithError(w, http.StatusBadRequest, formatValidationErrors(validationErrors))
		return
	}

	var (
		authResp *domain.AuthResponse
		err      error
	)

	switch req.UserType {
	case domain.UserTypeCustomer:
		authResp, err = h.authService.RegisterCustomer(r.Context(), req)
	case domain.UserTypeDriver:
		authResp, err = h.authService.RegisterDriver(r.Context(), req)
	default:
		respondWithError(w, http.StatusBadRequest, "Invalid user type")
		return
	}

	if err != nil {
		switch {
		case errors.Is(err, errors.New("email already exists")):
			respondWithError(w, http.StatusConflict, "Email already in use")
		default:
			respondWithError(w, http.StatusInternalServerError, "Failed to register user")
		}
		return
	}

	respondWithJSON(w, http.StatusCreated, authResp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.Login

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		respondWithError(w, http.StatusBadRequest, formatValidationErrors(validationErrors))
		return
	}

	authResp, err := h.authService.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			respondWithError(w, http.StatusUnauthorized, "Invalid email or password")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Login failed")
		return
	}

	respondWithJSON(w, http.StatusOK, authResp)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	refreshToken := reqBody.RefreshToken

	if refreshToken == "" {
		refreshToken = r.FormValue("refresh_token")
	}

	if refreshToken == "" {
		// Try to get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			refreshToken = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}
	fmt.Println("Refresh token:", refreshToken)
	if refreshToken == "" {
		respondWithError(w, http.StatusBadRequest, "Refresh token is required")
		return
	}

	// Generate new tokens
	authResp, err := h.authService.RefreshToken(r.Context(), refreshToken)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, service.ErrInvalidToken):
			statusCode = http.StatusUnauthorized
		case errors.Is(err, service.ErrExpiredToken):
			statusCode = http.StatusUnauthorized
		}

		respondWithError(w, statusCode, "Invalid or expired refresh token")
		return
	}

	respondWithJSON(w, http.StatusOK, authResp)
}

func (h *AuthHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req domain.PasswordReset

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validate.StructPartial(req, "Email"); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		respondWithError(w, http.StatusBadRequest, formatValidationErrors(validationErrors))
		return
	}

	_, err := h.authService.RequestPasswordReset(r.Context(), req.Email)
	if err != nil {
		// Don't expose details of failure, just return a generic message
		respondWithError(w, http.StatusInternalServerError, "Failed to process request")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "If your email exists in our system, you will receive password reset instructions",
	})
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func formatValidationErrors(errs validator.ValidationErrors) string {
	var errMsgs []string
	for _, err := range errs {
		errMsgs = append(errMsgs, err.Field()+" "+err.Tag())
	}
	return strings.Join(errMsgs, ", ")
}
