package handler

import (
	"encoding/json"
	"net/http"

	"mentor-api/internal/domain"
	"mentor-api/internal/interface/http/middleware"
	"mentor-api/internal/usecase/exam"

	"github.com/go-chi/chi/v5"
)

type ExamHandler struct {
	examUseCase exam.UseCase
}

func NewExamHandler(examUseCase exam.UseCase) *ExamHandler {
	return &ExamHandler{
		examUseCase: examUseCase,
	}
}

// Create godoc
// @Summary      Cria um novo simulado
// @Description  Gera um novo simulado para o usuário autenticado com base nas configurações enviadas
// @Tags         Exams
// @Accept       json
// @Produce      json
// @Param        request  body      exam.CreateExamInput  true  "Parâmetros de configuração do simulado"
// @Success      201      {object}  exam.ExamResponse       "Simulado criado com sucesso"
// @Failure      400      {object}  map[string]string     "Entrada inválida (ErrInvalidInput)"
// @Failure      401      {object}  map[string]string     "Não autorizado / Token JWT ausente ou inválido"
// @Failure      500      {object}  map[string]string     "Erro interno do servidor"
// @Security     BearerAuth
// @Router       /exams [post]
func (h *ExamHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		SendError(w, domain.ErrUnauthorized)
		return
	}

	var input exam.CreateExamInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		SendError(w, domain.ErrInvalidInput)
		return
	}

	res, err := h.examUseCase.CreateExam(r.Context(), userID, input)
	if err != nil {
		SendError(w, err)
		return
	}

	SendSuccess(w, http.StatusCreated, res)
}

// List godoc
// @Summary      Lista os simulados do usuário
// @Description  Retorna o histórico e a lista de simulados pertencentes ao usuário autenticado
// @Tags         Exams
// @Produce      json
// @Success      200  {array}   exam.ExamSummaryResponse  "Lista de simulados retornada com sucesso"
// @Failure      401  {object}  map[string]string       "Não autorizado / Token JWT ausente ou inválido"
// @Failure      500  {object}  map[string]string       "Erro interno do servidor"
// @Security     BearerAuth
// @Router       /exams [get]
func (h *ExamHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		SendError(w, domain.ErrUnauthorized)
		return
	}

	res, err := h.examUseCase.ListUserExams(r.Context(), userID)
	if err != nil {
		SendError(w, err)
		return
	}

	SendSuccess(w, http.StatusOK, res)
}

// GetByID godoc
// @Summary      Busca um simulado por ID
// @Description  Retorna os detalhes completos de um simulado específico pertencente ao usuário autenticado
// @Tags         Exams
// @Produce      json
// @Param        id   path      string             true  "ID do Simulado (UUID)"
// @Success      200  {object}  exam.ExamResponse    "Simulado encontrado com sucesso"
// @Failure      400  {object}  map[string]string  "ID inválido ou ausente"
// @Failure      401  {object}  map[string]string  "Não autorizado / Token JWT ausente ou inválido"
// @Failure      404  {object}  map[string]string  "Simulado não encontrado"
// @Failure      500  {object}  map[string]string  "Erro interno do servidor"
// @Security     BearerAuth
// @Router       /exams/{id} [get]
func (h *ExamHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		SendError(w, domain.ErrUnauthorized)
		return
	}

	examID := chi.URLParam(r, "id")
	if examID == "" {
		SendError(w, domain.ErrInvalidInput)
		return
	}

	res, err := h.examUseCase.GetExamByID(r.Context(), userID, examID)
	if err != nil {
		SendError(w, err)
		return
	}

	SendSuccess(w, http.StatusOK, res)
}

// SubmitExam godoc
// @Summary      Submete todas as respostas de um simulado
// @Description  Envia o lote completo de respostas de um simulado de uma só vez para avaliação
// @Tags         Exams
// @Accept       json
// @Produce      json
// @Param        id       path      string                     true  "ID do Simulado (UUID)"
// @Param        answers  body      []exam.AnswerQuestionInput true  "Lista de respostas das questões"
// @Success      200      {object}  exam.FinishExamResponse      "Simulado submetido e avaliado com sucesso"
// @Failure      400      {object}  map[string]string          "Entrada ou ID do simulado inválido"
// @Failure      401      {object}  map[string]string          "Não autorizado / Token JWT ausente ou inválido"
// @Failure      404      {object}  map[string]string          "Simulado não encontrado"
// @Failure      500      {object}  map[string]string          "Erro interno do servidor"
// @Security     BearerAuth
// @Router       /exams/{id}/submit [post]
func (h *ExamHandler) SubmitExam(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		SendError(w, domain.ErrUnauthorized)
		return
	}

	examID := chi.URLParam(r, "id")
	if examID == "" {
		SendError(w, domain.ErrInvalidInput)
		return
	}

	var answers []exam.AnswerQuestionInput
	if err := json.NewDecoder(r.Body).Decode(&answers); err != nil {
		SendError(w, domain.ErrInvalidInput)
		return
	}

	response, err := h.examUseCase.SubmitExam(r.Context(), examID, userID, answers)
	if err != nil {
		SendError(w, err)
		return
	}

	SendSuccess(w, http.StatusOK, response)
}

// AnswerQuestion godoc
// @Summary      Responde uma questão pontual do simulado
// @Description  Registra a resposta individual de uma questão no simulado em andamento
// @Tags         Exams
// @Accept       json
// @Produce      json
// @Param        id       path      string                   true  "ID do Simulado (UUID)"
// @Param        request  body      exam.AnswerQuestionInput true  "Dados da resposta da questão"
// @Success      200      {object}  map[string]string        "Resposta registrada com sucesso"
// @Failure      400      {object}  map[string]string        "Dados de resposta inválidos"
// @Failure      401      {object}  map[string]string        "Não autorizado / Token JWT ausente ou inválido"
// @Failure      404      {object}  map[string]string        "Simulado ou questão não encontrada"
// @Failure      500      {object}  map[string]string        "Erro interno do servidor"
// @Security     BearerAuth
// @Router       /exams/{id}/answers [post]
func (h *ExamHandler) AnswerQuestion(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		SendError(w, domain.ErrUnauthorized)
		return
	}

	examID := chi.URLParam(r, "id")
	if examID == "" {
		SendError(w, domain.ErrInvalidInput)
		return
	}

	var input exam.AnswerQuestionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		SendError(w, domain.ErrInvalidInput)
		return
	}

	if err := input.Validate(); err != nil {
		SendError(w, err)
		return
	}

	err := h.examUseCase.AnswerQuestion(r.Context(), userID, examID, input)
	if err != nil {
		SendError(w, err)
		return
	}

	SendSuccess(w, http.StatusOK, map[string]string{
		"message": "Answer recorded successfully",
	})
}

// Finish godoc
// @Summary      Finaliza um simulado
// @Description  Encerra o tempo do simulado e calcula a pontuação/resultado final
// @Tags         Exams
// @Produce      json
// @Param        id   path      string                 true  "ID do Simulado (UUID)"
// @Success      200  {object}  exam.FinishExamResponse  "Simulado finalizado com sucesso"
// @Failure      400  {object}  map[string]string      "ID do simulado inválido"
// @Failure      401  {object}  map[string]string      "Não autorizado / Token JWT ausente ou inválido"
// @Failure      404  {object}  map[string]string      "Simulado não encontrado"
// @Failure      500  {object}  map[string]string      "Erro interno do servidor"
// @Security     BearerAuth
// @Router       /exams/{id}/finish [post]
func (h *ExamHandler) Finish(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		SendError(w, domain.ErrUnauthorized)
		return
	}

	examID := chi.URLParam(r, "id")
	if examID == "" {
		SendError(w, domain.ErrInvalidInput)
		return
	}

	res, err := h.examUseCase.FinishExam(r.Context(), userID, examID)
	if err != nil {
		SendError(w, err)
		return
	}

	SendSuccess(w, http.StatusOK, res)
}

// Delete godoc
// @Summary      Remove um simulado
// @Description  Exclui um simulado específico pertencente ao usuário autenticado
// @Tags         Exams
// @Produce      json
// @Param        id   path      string             true  "ID do Simulado (UUID)"
// @Success      200  {object}  map[string]string  "Simulado excluído com sucesso"
// @Failure      400  {object}  map[string]string  "ID do simulado inválido"
// @Failure      401  {object}  map[string]string  "Não autorizado / Token JWT ausente ou inválido"
// @Failure      404  {object}  map[string]string  "Simulado não encontrado"
// @Failure      500  {object}  map[string]string  "Erro interno do servidor"
// @Security     BearerAuth
// @Router       /exams/{id} [delete]
func (h *ExamHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		SendError(w, domain.ErrUnauthorized)
		return
	}

	examID := chi.URLParam(r, "id")
	if examID == "" {
		SendError(w, domain.ErrInvalidInput)
		return
	}

	err := h.examUseCase.DeleteExam(r.Context(), userID, examID)
	if err != nil {
		SendError(w, err)
		return
	}

	SendSuccess(w, http.StatusOK, map[string]string{
		"message": "Exam deleted successfully",
	})
}
