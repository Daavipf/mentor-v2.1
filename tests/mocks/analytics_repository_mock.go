package mocks

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"mentor-api/internal/domain/entity"
	"mentor-api/internal/domain/repository"
)

type MockAnalyticsRepository struct {
	mu           sync.RWMutex
	examRepo     *MockExamRepository
	historyRepo  *MockHistoryRepository
	questionRepo *MockQuestionRepository

	CustomOverview    *entity.OverviewMetrics
	CustomDisciplines []entity.DisciplinePerformance
	CustomTopics      *entity.TopicsRanking
	CustomEvolution   []entity.TimePeriodMetric
	CustomRhythm      []entity.DailyStudyRhythm
	CustomExams       []entity.ExamPerformanceHistory
}

func NewMockAnalyticsRepository(
	examRepo *MockExamRepository,
	historyRepo *MockHistoryRepository,
	questionRepo *MockQuestionRepository,
) *MockAnalyticsRepository {
	return &MockAnalyticsRepository{
		examRepo:     examRepo,
		historyRepo:  historyRepo,
		questionRepo: questionRepo,
	}
}

func (m *MockAnalyticsRepository) GetOverview(ctx context.Context, userID string) (*entity.OverviewMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.CustomOverview != nil {
		ov := *m.CustomOverview
		return &ov, nil
	}

	totalAnswered := 0
	totalCorrect := 0

	if m.examRepo != nil {
		exams, _ := m.examRepo.ListByUserID(ctx, userID)
		for _, e := range exams {
			for _, qe := range e.Questions {
				if qe.SelectedAlternativeID != nil || qe.GotRight != nil {
					totalAnswered++
					if qe.GotRight != nil && *qe.GotRight {
						totalCorrect++
					}
				}
			}
		}
	}

	overallAccuracy := 0.0
	if totalAnswered > 0 {
		overallAccuracy = math.Round((float64(totalCorrect)/float64(totalAnswered))*100*100) / 100
	}

	completedExams := 0
	scoreSum := 0
	if m.historyRepo != nil {
		histories, _ := m.historyRepo.ListByUserID(ctx, userID)
		for _, h := range histories {
			if h.FinishedAt != nil {
				completedExams++
				if h.Score != nil {
					scoreSum += *h.Score
				}
			}
		}
	}

	avgScore := 0.0
	if completedExams > 0 {
		avgScore = math.Round((float64(scoreSum)/float64(completedExams))*100) / 100
	}

	areas, _ := m.GetPerformanceByArea(ctx, userID)
	rhythm, _ := m.GetStudyRhythm(ctx, userID, 30)

	return &entity.OverviewMetrics{
		TotalQuestionsAnswered: totalAnswered,
		TotalCorrectAnswers:    totalCorrect,
		OverallAccuracyRate:    overallAccuracy,
		TotalCompletedExams:    completedExams,
		AverageExamScore:       avgScore,
		PerformanceByArea:      areas,
		RecentStudyRhythm:      rhythm,
	}, nil
}

func (m *MockAnalyticsRepository) GetPerformanceByArea(ctx context.Context, userID string) ([]entity.AreaPerformance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	type stats struct {
		area    string
		total   int
		correct int
	}
	areaMap := make(map[string]*stats)

	if m.examRepo != nil {
		exams, _ := m.examRepo.ListByUserID(ctx, userID)
		for _, e := range exams {
			for _, qe := range e.Questions {
				if qe.SelectedAlternativeID != nil || qe.GotRight != nil {
					area := "Geral"
					if qe.Question != nil && qe.Question.Area != "" {
						area = qe.Question.Area
					}
					st, exists := areaMap[area]
					if !exists {
						st = &stats{area: area}
						areaMap[area] = st
					}
					st.total++
					if qe.GotRight != nil && *qe.GotRight {
						st.correct++
					}
				}
			}
		}
	}

	var result []entity.AreaPerformance
	for _, st := range areaMap {
		acc := 0.0
		if st.total > 0 {
			acc = math.Round((float64(st.correct)/float64(st.total))*100*100) / 100
		}
		result = append(result, entity.AreaPerformance{
			Area:           st.area,
			TotalQuestions: st.total,
			CorrectAnswers: st.correct,
			AccuracyRate:   acc,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalQuestions > result[j].TotalQuestions
	})

	return result, nil
}

func (m *MockAnalyticsRepository) GetPerformanceByDiscipline(
	ctx context.Context,
	userID string,
	filter repository.DisciplineFilter,
) ([]entity.DisciplinePerformance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.CustomDisciplines != nil {
		return m.CustomDisciplines, nil
	}

	type stats struct {
		disc    string
		area    string
		total   int
		correct int
	}
	discMap := make(map[string]*stats)

	if m.examRepo != nil {
		exams, _ := m.examRepo.ListByUserID(ctx, userID)
		for _, e := range exams {
			if filter.StartDate != nil && e.Date.Before(*filter.StartDate) {
				continue
			}
			if filter.EndDate != nil && e.Date.After(*filter.EndDate) {
				continue
			}

			for _, qe := range e.Questions {
				if qe.SelectedAlternativeID != nil || qe.GotRight != nil {
					area := "Geral"
					disc := "Geral"
					if qe.Question != nil {
						if qe.Question.Area != "" {
							area = qe.Question.Area
						}
						if qe.Question.Discipline != nil && *qe.Question.Discipline != "" {
							disc = *qe.Question.Discipline
						}
					}

					if filter.Area != "" && filter.Area != area {
						continue
					}

					key := area + ":" + disc
					st, exists := discMap[key]
					if !exists {
						st = &stats{disc: disc, area: area}
						discMap[key] = st
					}
					st.total++
					if qe.GotRight != nil && *qe.GotRight {
						st.correct++
					}
				}
			}
		}
	}

	var result []entity.DisciplinePerformance
	for _, st := range discMap {
		acc := 0.0
		if st.total > 0 {
			acc = math.Round((float64(st.correct)/float64(st.total))*100*100) / 100
		}
		result = append(result, entity.DisciplinePerformance{
			Discipline:     st.disc,
			Area:           st.area,
			TotalQuestions: st.total,
			CorrectAnswers: st.correct,
			AccuracyRate:   acc,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalQuestions > result[j].TotalQuestions
	})

	return result, nil
}

func (m *MockAnalyticsRepository) GetTopicsRanking(
	ctx context.Context,
	userID string,
	minQuestions int,
) (*entity.TopicsRanking, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.CustomTopics != nil {
		return m.CustomTopics, nil
	}

	type stats struct {
		topic   string
		total   int
		correct int
	}
	topicMap := make(map[string]*stats)

	if m.examRepo != nil {
		exams, _ := m.examRepo.ListByUserID(ctx, userID)
		for _, e := range exams {
			for _, qe := range e.Questions {
				if qe.SelectedAlternativeID != nil || qe.GotRight != nil {
					if qe.Question != nil {
						for _, t := range qe.Question.Topics {
							st, exists := topicMap[t.Topic]
							if !exists {
								st = &stats{topic: t.Topic}
								topicMap[t.Topic] = st
							}
							st.total++
							if qe.GotRight != nil && *qe.GotRight {
								st.correct++
							}
						}
					}
				}
			}
		}
	}

	var allTopics []entity.TopicPerformance
	for _, st := range topicMap {
		acc := 0.0
		if st.total > 0 {
			acc = math.Round((float64(st.correct)/float64(st.total))*100*100) / 100
		}
		allTopics = append(allTopics, entity.TopicPerformance{
			Topic:          st.topic,
			TotalQuestions: st.total,
			CorrectAnswers: st.correct,
			AccuracyRate:   acc,
		})
	}

	// Strong topics: total >= minQuestions, highest accuracy rate
	var strongCandidates []entity.TopicPerformance
	for _, t := range allTopics {
		if t.TotalQuestions >= minQuestions {
			strongCandidates = append(strongCandidates, t)
		}
	}
	sort.Slice(strongCandidates, func(i, j int) bool {
		if strongCandidates[i].AccuracyRate == strongCandidates[j].AccuracyRate {
			return strongCandidates[i].TotalQuestions > strongCandidates[j].TotalQuestions
		}
		return strongCandidates[i].AccuracyRate > strongCandidates[j].AccuracyRate
	})
	if len(strongCandidates) > 5 {
		strongCandidates = strongCandidates[:5]
	}

	// Weak topics: lowest accuracy rate
	weakCandidates := make([]entity.TopicPerformance, len(allTopics))
	copy(weakCandidates, allTopics)
	sort.Slice(weakCandidates, func(i, j int) bool {
		if weakCandidates[i].AccuracyRate == weakCandidates[j].AccuracyRate {
			return weakCandidates[i].TotalQuestions > weakCandidates[j].TotalQuestions
		}
		return weakCandidates[i].AccuracyRate < weakCandidates[j].AccuracyRate
	})
	if len(weakCandidates) > 5 {
		weakCandidates = weakCandidates[:5]
	}

	return &entity.TopicsRanking{
		StrongTopics: strongCandidates,
		WeakTopics:   weakCandidates,
	}, nil
}

func (m *MockAnalyticsRepository) GetEvolution(
	ctx context.Context,
	userID string,
	periodType string,
	months int,
) ([]entity.TimePeriodMetric, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.CustomEvolution != nil {
		return m.CustomEvolution, nil
	}

	type stats struct {
		period     string
		periodDate time.Time
		total      int
		correct    int
	}
	periodMap := make(map[string]*stats)

	if m.examRepo != nil {
		exams, _ := m.examRepo.ListByUserID(ctx, userID)
		for _, e := range exams {
			for _, qe := range e.Questions {
				if qe.SelectedAlternativeID != nil || qe.GotRight != nil {
					t := qe.UpdatedAt
					if t.IsZero() {
						t = e.Date
					}

					var periodLabel string
					var periodDate time.Time
					if periodType == "monthly" {
						periodLabel = t.Format("2006-01")
						periodDate = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
					} else {
						year, week := t.ISOWeek()
						periodLabel = fmt.Sprintf("%d-W%02d", year, week)
						periodDate = t
					}

					st, exists := periodMap[periodLabel]
					if !exists {
						st = &stats{period: periodLabel, periodDate: periodDate}
						periodMap[periodLabel] = st
					}
					st.total++
					if qe.GotRight != nil && *qe.GotRight {
						st.correct++
					}
				}
			}
		}
	}

	var result []entity.TimePeriodMetric
	for _, st := range periodMap {
		acc := 0.0
		if st.total > 0 {
			acc = math.Round((float64(st.correct)/float64(st.total))*100*100) / 100
		}
		result = append(result, entity.TimePeriodMetric{
			Period:         st.period,
			PeriodDate:     st.periodDate,
			TotalQuestions: st.total,
			CorrectAnswers: st.correct,
			AccuracyRate:   acc,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].PeriodDate.Before(result[j].PeriodDate)
	})

	return result, nil
}

func (m *MockAnalyticsRepository) GetStudyRhythm(ctx context.Context, userID string, days int) ([]entity.DailyStudyRhythm, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.CustomRhythm != nil {
		return m.CustomRhythm, nil
	}

	type stats struct {
		day     string
		total   int
		correct int
	}
	dayMap := make(map[string]*stats)

	if m.examRepo != nil {
		exams, _ := m.examRepo.ListByUserID(ctx, userID)
		for _, e := range exams {
			for _, qe := range e.Questions {
				if qe.SelectedAlternativeID != nil || qe.GotRight != nil {
					t := qe.UpdatedAt
					if t.IsZero() {
						t = e.Date
					}
					dayStr := t.Format("2006-01-02")
					st, exists := dayMap[dayStr]
					if !exists {
						st = &stats{day: dayStr}
						dayMap[dayStr] = st
					}
					st.total++
					if qe.GotRight != nil && *qe.GotRight {
						st.correct++
					}
				}
			}
		}
	}

	var result []entity.DailyStudyRhythm
	for _, st := range dayMap {
		result = append(result, entity.DailyStudyRhythm{
			Date:           st.day,
			TotalQuestions: st.total,
			CorrectAnswers: st.correct,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Date < result[j].Date
	})

	return result, nil
}

func (m *MockAnalyticsRepository) GetExamsHistory(ctx context.Context, userID string) ([]entity.ExamPerformanceHistory, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.CustomExams != nil {
		return m.CustomExams, nil
	}

	var result []entity.ExamPerformanceHistory

	if m.historyRepo != nil && m.examRepo != nil {
		histories, _ := m.historyRepo.ListByUserID(ctx, userID)
		for _, h := range histories {
			if h.FinishedAt != nil {
				exam, err := m.examRepo.GetByID(ctx, h.ExamID)
				if err == nil && exam != nil {
					scoreVal := 0
					if h.Score != nil {
						scoreVal = *h.Score
					}
					correct := 0
					for _, qe := range exam.Questions {
						if qe.GotRight != nil && *qe.GotRight {
							correct++
						}
					}
					result = append(result, entity.ExamPerformanceHistory{
						ExamID:         exam.ID,
						ExamTitle:      exam.Title,
						Date:           exam.Date,
						FinishedAt:     h.FinishedAt,
						Score:          scoreVal,
						TotalQuestions: len(exam.Questions),
						CorrectAnswers: correct,
					})
				}
			}
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].FinishedAt != nil && result[j].FinishedAt != nil {
			return result[i].FinishedAt.Before(*result[j].FinishedAt)
		}
		return result[i].Date.Before(result[j].Date)
	})

	return result, nil
}

