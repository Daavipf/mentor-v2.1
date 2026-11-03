package entity

import "time"

// OverviewMetrics represents executive summary metrics for a given user.
type OverviewMetrics struct {
	TotalQuestionsAnswered int                `json:"total_questions_answered"`
	TotalCorrectAnswers    int                `json:"total_correct_answers"`
	OverallAccuracyRate    float64            `json:"overall_accuracy_rate"`
	TotalCompletedExams    int                `json:"total_completed_exams"`
	AverageExamScore       float64            `json:"average_exam_score"`
	PerformanceByArea      []AreaPerformance  `json:"performance_by_area,omitempty"`
	RecentStudyRhythm      []DailyStudyRhythm `json:"recent_study_rhythm,omitempty"`
	AverageQuestionTime    float64            `json:"average_question_time"`
}

// AreaPerformance represents accuracy and question volume per knowledge area.
type AreaPerformance struct {
	Area           string  `json:"area"`
	TotalQuestions int     `json:"total_questions"`
	CorrectAnswers int     `json:"correct_answers"`
	AccuracyRate   float64 `json:"accuracy_rate"`
}

// DisciplinePerformance represents performance grouped by discipline.
type DisciplinePerformance struct {
	Discipline     string  `json:"discipline"`
	Area           string  `json:"area"`
	TotalQuestions int     `json:"total_questions"`
	CorrectAnswers int     `json:"correct_answers"`
	AccuracyRate   float64 `json:"accuracy_rate"`
}

// TopicPerformance represents performance grouped by specific question topics.
type TopicPerformance struct {
	Topic          string  `json:"topic"`
	TotalQuestions int     `json:"total_questions"`
	CorrectAnswers int     `json:"correct_answers"`
	AccuracyRate   float64 `json:"accuracy_rate"`
}

// TopicsRanking aggregates top performing and weakest topics for the user.
type TopicsRanking struct {
	StrongTopics []TopicPerformance `json:"strong_topics"`
	WeakTopics   []TopicPerformance `json:"weak_topics"`
}

// TimePeriodMetric represents aggregated temporal metric for a specific week or month.
type TimePeriodMetric struct {
	Period         string    `json:"period"`
	PeriodDate     time.Time `json:"period_date"`
	TotalQuestions int       `json:"total_questions"`
	CorrectAnswers int       `json:"correct_answers"`
	AccuracyRate   float64   `json:"accuracy_rate"`
}

// DailyStudyRhythm represents question count answered per day for heatmaps/bar charts.
type DailyStudyRhythm struct {
	Date           string `json:"date"`
	TotalQuestions int    `json:"total_questions"`
	CorrectAnswers int    `json:"correct_answers"`
}

// ExamPerformanceHistory holds comparative exam record data over time.
type ExamPerformanceHistory struct {
	ExamID         string     `json:"exam_id"`
	ExamTitle      string     `json:"exam_title"`
	Date           time.Time  `json:"date"`
	FinishedAt     *time.Time `json:"finished_at,omitempty"`
	Score          int        `json:"score"`
	TotalQuestions int        `json:"total_questions"`
	CorrectAnswers int        `json:"correct_answers"`
}

// ExamsHistoryAnalytics presents comparative exam performance and percentage score evolution.
type ExamsHistoryAnalytics struct {
	Exams                    []ExamPerformanceHistory `json:"exams"`
	TotalFinishedExams       int                      `json:"total_finished_exams"`
	OverallAverageScore      float64                  `json:"overall_average_score"`
	RecentAverageScore       float64                  `json:"recent_average_score"`
	ScoreEvolutionPercentage float64                  `json:"score_evolution_percentage"`
}
