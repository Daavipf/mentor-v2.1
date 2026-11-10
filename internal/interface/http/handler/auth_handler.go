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

// Register godoc
// @Summary      Cadastro de novo usuário
// @Description  Cria uma nova conta na plataforma a partir dos dados informados
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      auth.RegisterInput  true  "Dados para cadastro do usuário"
// @Success      201      {object}  auth.AuthResponse   "Usuário criado com sucesso"
// @Failure      400      {object}  map[string]string   "Entrada inválida (ErrInvalidInput)"
// @Failure      409      {object}  map[string]string   "Conflito / E-mail já cadastrado"
// @Failure      500      {object}  map[string]string   "Erro interno do servidor"
// @Router       /auth/register [post]
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

	http.SetCookie(w, &http.Cookie{
		Name:     "mentor_token",
		Value:    res.Token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   30 * 24 * 3600, // 30 days
	})

	SendSuccess(w, http.StatusCreated, res)
}

// Login godoc
// @Summary      Autenticação de usuário
// @Description  Realiza login com e-mail e senha, retornando o token JWT de acesso
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      auth.LoginInput    true  "Credenciais de acesso"
// @Success      200      {object}  auth.AuthResponse   "Autenticado com sucesso"
// @Failure      400      {object}  map[string]string  "Entrada inválida (ErrInvalidInput)"
// @Failure      401      {object}  map[string]string  "Credenciais incorretas (ErrUnauthorized)"
// @Failure      500      {object}  map[string]string  "Erro interno do servidor"
// @Router       /auth/login [post]
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

	http.SetCookie(w, &http.Cookie{
		Name:     "mentor_token",
		Value:    res.Token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   30 * 24 * 3600, // 30 days
	})

	SendSuccess(w, http.StatusOK, res)
}

// Logout godoc
// @Summary      Logout de usuário
// @Description  Limpa o cookie de sessão http-only
// @Tags         Auth
// @Produce      json
// @Success      200      {object}  map[string]string   "Logout realizado com sucesso"
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "mentor_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	SendSuccess(w, http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}

// Me godoc
// @Summary      Obtém perfil do usuário logado
// @Description  Retorna as informações do perfil associado ao token JWT informado no header Authorization
// @Tags         Auth
// @Produce      json
// @Success      200      {object}  auth.UserResponse  "Dados do perfil do usuário"
// @Failure      401      {object}  map[string]string   "Não autorizado / Token inválido ou ausente"
// @Failure      404      {object}  map[string]string   "Usuário não encontrado"
// @Failure      500      {object}  map[string]string   "Erro interno do servidor"
// @Security     BearerAuth
// @Router       /auth/me [get]
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
