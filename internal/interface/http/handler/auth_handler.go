package handler

import (
	"encoding/json"
	"net/http"

	"mentor-api/internal/domain"
	"mentor-api/internal/interface/http/middleware"
	"mentor-api/internal/usecase/auth"
)

type AuthHandler struct {
	authUseCase auth.UseCase
}

func NewAuthHandler(authUseCase auth.UseCase) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input auth.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		SendError(w, domain.ErrInvalidInput)
		return
	}

	res, err := h.authUseCase.Register(r.Context(), input)
	if err != nil {
		SendError(w, err)
		return
	}

	SendSuccess(w, http.StatusCreated, res)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input auth.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		SendError(w, domain.ErrInvalidInput)
		return
	}

	res, err := h.authUseCase.Login(r.Context(), input)
	if err != nil {
		SendError(w, err)
		return
	}

	SendSuccess(w, http.StatusOK, res)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		SendError(w, domain.ErrUnauthorized)
		return
	}

	res, err := h.authUseCase.GetProfile(r.Context(), userID)
	if err != nil {
		SendError(w, err)
		return
	}

	SendSuccess(w, http.StatusOK, res)
}
