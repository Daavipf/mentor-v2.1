package analytics

import (
	"time"
)

// AreaPerformanceDTO represents performance in a specific knowledge area.
type AreaPerformanceDTO struct {
	Area           string  `json:"area"`
	TotalQuestions int     `json:"total_questions"`
	CorrectAnswers int     `json:"correct_answers"`
	AccuracyRate   float64 `json:"accuracy_rate"`
}

// DailyStudyRhythmDTO represents question volume answered per day.
type DailyStudyRhythmDTO struct {
	Date           string `json:"date"`
	TotalQuestions int    `json:"total_questions"`
	CorrectAnswers int    `json:"correct_answers"`
}

// OverviewResponse represents the general analytics summary.
type OverviewResponse struct {
	TotalQuestionsAnswered int                   `json:"total_questions_answered"`
	TotalCorrectAnswers    int                   `json:"total_correct_answers"`
	OverallAccuracyRate    float64               `json:"overall_accuracy_rate"`
	TotalCompletedExams    int                   `json:"total_completed_exams"`
	AverageExamScore       float64               `json:"average_exam_score"`
	PerformanceByArea      []AreaPerformanceDTO  `json:"performance_by_area"`
	RecentStudyRhythm      []DailyStudyRhythmDTO `json:"recent_study_rhythm"`
}

// ByDisciplineFilterInput holds query parameters for discipline performance.
type ByDisciplineFilterInput struct {
	StartDate *time.Time
	EndDate   *time.Time
	Area      string
}

// DisciplinePerformanceDTO represents performance in a single discipline.
type DisciplinePerformanceDTO struct {
	Discipline     string  `json:"discipline"`
	Area           string  `json:"area"`
	TotalQuestions int     `json:"total_questions"`
	CorrectAnswers int     `json:"correct_answers"`
	AccuracyRate   float64 `json:"accuracy_rate"`
}

// ByDisciplineResponse wraps discipline performance results.
type ByDisciplineResponse struct {
	Disciplines      []DisciplinePerformanceDTO `json:"disciplines"`
	TotalDisciplines int                        `json:"total_disciplines"`
}

// TopicPerformanceDTO represents performance in a specific topic.
type TopicPerformanceDTO struct {
	Topic          string  `json:"topic"`
	TotalQuestions int     `json:"total_questions"`
	CorrectAnswers int     `json:"correct_answers"`
	AccuracyRate   float64 `json:"accuracy_rate"`
}

// TopicsRankingResponse wraps top strong topics and top weak topics.
type TopicsRankingResponse struct {
	MinQuestionsThreshold int                   `json:"min_questions_threshold"`
	StrongTopics          []TopicPerformanceDTO `json:"strong_topics"`
	WeakTopics            []TopicPerformanceDTO `json:"weak_topics"`
}

// TimePeriodMetricDTO represents aggregated performance point in time.
type TimePeriodMetricDTO struct {
	Period         string    `json:"period"`
	PeriodDate     time.Time `json:"period_date"`
	TotalQuestions int       `json:"total_questions"`
	CorrectAnswers int       `json:"correct_answers"`
	AccuracyRate   float64   `json:"accuracy_rate"`
}

// EvolutionResponse represents temporal evolution over time.
type EvolutionResponse struct {
	PeriodType  string                `json:"period_type"`
	Points      []TimePeriodMetricDTO `json:"points"`
	StudyRhythm []DailyStudyRhythmDTO `json:"study_rhythm"`
}

// ExamPerformanceHistoryDTO represents a completed exam record.
type ExamPerformanceHistoryDTO struct {
	ExamID         string     `json:"exam_id"`
	ExamTitle      string     `json:"exam_title"`
	Date           time.Time  `json:"date"`
	FinishedAt     *time.Time `json:"finished_at,omitempty"`
	Score          int        `json:"score"`
	TotalQuestions int        `json:"total_questions"`
	CorrectAnswers int        `json:"correct_answers"`
}

// ExamsHistoryResponse represents comparative exam history and score evolution.
type ExamsHistoryResponse struct {
	Exams                    []ExamPerformanceHistoryDTO `json:"exams"`
	TotalFinishedExams       int                         `json:"total_finished_exams"`
	OverallAverageScore      float64                     `json:"overall_average_score"`
	RecentAverageScore       float64                     `json:"recent_average_score"`
	ScoreEvolutionPercentage float64                     `json:"score_evolution_percentage"`
}

