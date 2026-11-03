package exam

import (
	"context"
	"strings"
	"time"

	"mentor-api/internal/domain"
	"mentor-api/internal/domain/entity"
	"mentor-api/internal/domain/repository"
	"mentor-api/internal/usecase/question"

	"github.com/google/uuid"
)

type UseCase interface {
	CreateExam(ctx context.Context, userID string, input CreateExamInput) (*ExamResponse, error)
	GetExamByID(ctx context.Context, userID string, examID string) (*ExamResponse, error)
	ListUserExams(ctx context.Context, userID string) ([]ExamSummaryResponse, error)
	AnswerQuestion(ctx context.Context, userID string, examID string, input AnswerQuestionInput) error
	FinishExam(ctx context.Context, userID string, examID string) (*FinishExamResponse, error)
	DeleteExam(ctx context.Context, userID string, examID string) error
	SubmitExam(ctx context.Context, examID, userID string, answers []AnswerQuestionInput) (*FinishExamResponse, error)
}

type examUseCase struct {
	examRepo        repository.ExamRepository
	historyRepo     repository.HistoryRepository
	questionService question.Service
}

func NewExamUseCase(
	examRepo repository.ExamRepository,
	historyRepo repository.HistoryRepository,
	questionService question.Service,
) UseCase {
	return &examUseCase{
		examRepo:        examRepo,
		historyRepo:     historyRepo,
		questionService: questionService,
	}
}

func (u *examUseCase) CreateExam(ctx context.Context, userID string, input CreateExamInput) (*ExamResponse, error) {
	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		return nil, domain.ErrInvalidInput
	}

	examDate := time.Now()
	if input.Date != nil {
		examDate = *input.Date
	}

	// Select questions internally using question service
	selectedQuestions, err := u.questionService.SelectQuestionsForExam(ctx, input.Areas, input.QuestionCount)
	if err != nil {
		return nil, err
	}

	examID := uuid.New().String()
	now := time.Now()

	examAreas := make([]entity.ExamArea, 0, len(input.Areas))
	for _, a := range input.Areas {
		examAreas = append(examAreas, entity.ExamArea{
			ID:      uuid.New().String(),
			ExamsID: examID,
			Area:    a,
		})
	}

	questionsOnExam := make([]entity.QuestionsOnExam, 0, len(selectedQuestions))
	for _, q := range selectedQuestions {
		qCopy := q
		questionsOnExam = append(questionsOnExam, entity.QuestionsOnExam{
			QuestionID: q.ID,
			ExamID:     examID,
			CreatedAt:  now,
			UpdatedAt:  now,
			Question:   &qCopy,
		})
	}

	exam := &entity.Exam{
		ID:        examID,
		UserID:    userID,
		Date:      examDate,
		Title:     input.Title,
		Areas:     examAreas,
		Questions: questionsOnExam,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := u.examRepo.Create(ctx, exam); err != nil {
		return nil, err
	}

	return u.mapToExamResponse(exam, false, nil), nil
}

func (u *examUseCase) GetExamByID(ctx context.Context, userID string, examID string) (*ExamResponse, error) {
	exam, err := u.examRepo.GetByID(ctx, examID)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	// Strict User Ownership Check: Only owner can access their exam
	if exam.UserID != userID {
		return nil, domain.ErrForbidden
	}

	// Check if exam is already finished
	history, err := u.historyRepo.GetByExamID(ctx, examID)
	isFinished := err == nil && history != nil && history.FinishedAt != nil
	var score *int
	if isFinished {
		score = history.Score
	}

	return u.mapToExamResponse(exam, isFinished, score), nil
}

func (u *examUseCase) ListUserExams(ctx context.Context, userID string) ([]ExamSummaryResponse, error) {
	exams, err := u.examRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	histories, err := u.historyRepo.ListByUserID(ctx, userID)
	historyByExam := make(map[string]*entity.History)
	if err == nil {
		for i := range histories {
			historyByExam[histories[i].ExamID] = &histories[i]
		}
	}

	summaries := make([]ExamSummaryResponse, 0, len(exams))
	for _, e := range exams {
		h := historyByExam[e.ID]
		isFinished := h != nil && h.FinishedAt != nil
		var score *int
		if isFinished {
			score = h.Score
		}

		areas := make([]string, 0, len(e.Areas))
		for _, a := range e.Areas {
			areas = append(areas, a.Area)
		}

		summaries = append(summaries, ExamSummaryResponse{
			ID:             e.ID,
			UserID:         e.UserID,
			Title:          e.Title,
			Date:           e.Date,
			Areas:          areas,
			TotalQuestions: len(e.Questions),
			IsFinished:     isFinished,
			Score:          score,
			CreatedAt:      e.CreatedAt,
			UpdatedAt:      e.UpdatedAt,
		})
	}

	return summaries, nil
}

func (u *examUseCase) AnswerQuestion(ctx context.Context, userID string, examID string, input AnswerQuestionInput) error {
	exam, err := u.examRepo.GetByID(ctx, examID)
	if err != nil {
		return domain.ErrNotFound
	}

	// Strict User Ownership Check
	if exam.UserID != userID {
		return domain.ErrForbidden
	}

	// Check if already finished
	history, err := u.historyRepo.GetByExamID(ctx, examID)
	if err == nil && history != nil && history.FinishedAt != nil {
		return domain.ErrExamAlreadyFinished
	}

	// Verify question belongs to this exam
	qOnExam, err := u.examRepo.GetQuestionOnExam(ctx, examID, input.QuestionID)
	if err != nil || qOnExam == nil {
		return domain.ErrQuestionNotInExam
	}

	// Validate alternative internally using question service
	isCorrect, err := u.questionService.ValidateAnswer(ctx, input.QuestionID, input.SelectedAlternativeID)
	if err != nil {
		return err
	}

	return u.examRepo.AnswerQuestion(ctx, examID, input.QuestionID, input.SelectedAlternativeID, *input.StartedAt, *input.FinishedAt, isCorrect)
}

func (u *examUseCase) FinishExam(ctx context.Context, userID string, examID string) (*FinishExamResponse, error) {
	exam, err := u.examRepo.GetByID(ctx, examID)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	// Strict User Ownership Check
	if exam.UserID != userID {
		return nil, domain.ErrForbidden
	}

	// Check if already finished
	existingHistory, err := u.historyRepo.GetByExamID(ctx, examID)
	if err == nil && existingHistory != nil && existingHistory.FinishedAt != nil {
		return nil, domain.ErrExamAlreadyFinished
	}

	// Calculate score
	totalQuestions := len(exam.Questions)
	correctCount := 0
	for _, qe := range exam.Questions {
		if qe.GotRight != nil && *qe.GotRight {
			correctCount++
		}
	}

	score := 0
	if totalQuestions > 0 {
		score = (correctCount * 100) / totalQuestions
	}

	now := time.Now()
	history := &entity.History{
		ID:         uuid.New().String(),
		UserID:     userID,
		ExamID:     examID,
		Score:      &score,
		CreatedAt:  now,
		UpdatedAt:  now,
		FinishedAt: &now,
	}

	if err := u.historyRepo.Create(ctx, history); err != nil {
		return nil, err
	}

	return &FinishExamResponse{
		ExamID:         examID,
		TotalQuestions: totalQuestions,
		CorrectAnswers: correctCount,
		Score:          score,
		FinishedAt:     now,
	}, nil
}

func (u *examUseCase) DeleteExam(ctx context.Context, userID string, examID string) error {
	exam, err := u.examRepo.GetByID(ctx, examID)
	if err != nil {
		return domain.ErrNotFound
	}

	// Strict User Ownership Check
	if exam.UserID != userID {
		return domain.ErrForbidden
	}

	return u.examRepo.Delete(ctx, examID)
}

func (u *examUseCase) mapToExamResponse(exam *entity.Exam, isFinished bool, score *int) *ExamResponse {
	areas := make([]string, 0, len(exam.Areas))
	for _, a := range exam.Areas {
		areas = append(areas, a.Area)
	}

	questionsDTO := make([]ExamQuestionDTO, 0, len(exam.Questions))
	for _, qe := range exam.Questions {
		var qDTO ExamQuestionDTO
		qDTO.QuestionID = qe.QuestionID
		qDTO.SelectedAlternativeID = qe.SelectedAlternativeID

		if isFinished {
			qDTO.GotRight = qe.GotRight
		}

		if qe.Question != nil {
			q := qe.Question
			qDTO.Context = q.Context
			qDTO.AlternativesIntroduction = q.AlternativesIntroduction
			qDTO.Area = q.Area
			qDTO.Discipline = q.Discipline
			qDTO.Year = q.Year
			qDTO.Language = q.Language
			qDTO.Index = q.Index
			qDTO.Title = q.Title

			topics := make([]string, 0, len(q.Topics))
			for _, t := range q.Topics {
				topics = append(topics, t.Topic)
			}
			qDTO.Topics = topics

			images := make([]string, 0, len(q.Images))
			for _, img := range q.Images {
				images = append(images, img.Path)
			}
			qDTO.Images = images

			alts := make([]ExamAlternativeDTO, 0, len(q.Alternatives))
			for _, alt := range q.Alternatives {
				altDTO := ExamAlternativeDTO{
					ID:        alt.ID,
					Text:      alt.Text,
					ImagePath: alt.ImagePath,
				}
				// Crucial security rule: only reveal correct answer after exam is finished
				if isFinished {
					isCorrectCopy := alt.IsCorrect
					altDTO.IsCorrect = &isCorrectCopy
				}
				alts = append(alts, altDTO)
			}
			qDTO.Alternatives = alts
		}

		questionsDTO = append(questionsDTO, qDTO)
	}

	return &ExamResponse{
		ID:             exam.ID,
		UserID:         exam.UserID,
		Title:          exam.Title,
		Date:           exam.Date,
		Areas:          areas,
		Questions:      questionsDTO,
		TotalQuestions: len(questionsDTO),
		IsFinished:     isFinished,
		Score:          score,
		CreatedAt:      exam.CreatedAt,
		UpdatedAt:      exam.UpdatedAt,
	}
}
