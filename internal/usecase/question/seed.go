package question

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"

	"mentor-api/internal/infrastructure/config"
)

const (
	YearStart = 2023
	YearEnd   = 2023
	Limit     = 200
)

// DTOs para deserialização da API externa do ENEM
type APIResponse struct {
	Metadata  Metadata      `json:"metadata"`
	Questions []APIQuestion `json:"questions"`
}

type Metadata struct {
	Limit   int  `json:"limit"`
	Offset  int  `json:"offset"`
	Total   int  `json:"total"`
	HasMore bool `json:"hasMore"`
}

type APIQuestion struct {
	Title                    string           `json:"title"`
	Index                    int              `json:"index"`
	Discipline               string           `json:"discipline"`
	Language                 *string          `json:"language"`
	Year                     int              `json:"year"`
	Context                  string           `json:"context"`
	Files                    []string         `json:"files"`
	CorrectAlternative       string           `json:"correctAlternative"`
	AlternativesIntroduction string           `json:"alternativesIntroduction"`
	Alternatives             []APIAlternative `json:"alternatives"`
}

type APIAlternative struct {
	Letter    string  `json:"letter"`
	Text      *string `json:"text"`
	File      *string `json:"file"`
	IsCorrect bool    `json:"isCorrect"`
}

func (s *questionService) SeedDefaultQuestions(ctx context.Context, cfg *config.Config) error {
	data := make(map[int][]APIQuestion)

	for year := YearStart; year <= YearEnd; year++ {
		response, err := fetchQuestionsFromAPI(cfg, year, nil)
		if err != nil {
			log.Fatalf("Failed to fetch questions for year %d: %v", year, err)
			return err
		}
		data[year] = response.Questions

		if year >= 2017 {
			lang := "ingles"
			engResponse, err := fetchQuestionsFromAPI(cfg, year, &lang)
			if err != nil {
				log.Fatalf("Failed to fetch questions for year %d: %v", year, err)
				return err
			}
			data[year] = append(data[year], engResponse.Questions...)
		}
	}

	for year := YearStart; year <= YearEnd; year++ {
		questions := data[year]

		fmt.Printf("Seeding %d questions for year %d...\n", len(questions), year)

		for _, question := range questions {
			questionPayload := mapQuestionToEntity(question)
			if _, err := s.CreateQuestion(ctx, questionPayload); err != nil {
				return err
			}
		}
	}

	return nil
}

func fetchQuestionsFromAPI(cfg *config.Config, year int, language *string) (*APIResponse, error) {
	url := fmt.Sprintf("%s/exams/%d/questions?limit=%d", cfg.EnemApiURL, year, Limit)

	if language != nil {
		url = fmt.Sprintf("%s/exams/%d/questions?limit=5&language=%s", cfg.EnemApiURL, year, *language)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResponse APIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, err
	}

	return &apiResponse, nil
}

func mapQuestionToEntity(apiQuestion APIQuestion) CreateQuestionInput {
	alternatives := make([]CreateAlternativeInput, len(apiQuestion.Alternatives))
	for i, alt := range apiQuestion.Alternatives {
		isCorrect := alt.IsCorrect || (alt.Letter != "" && alt.Letter == apiQuestion.CorrectAlternative)

		imagePath, err := formatEnemURL(alt.File, "/assets")
		if err != nil {
			log.Fatalf("Error formatting URL for alternative: %v", err)
		}
		alternatives[i] = CreateAlternativeInput{
			Text:      alt.Text,
			ImagePath: imagePath,
			IsCorrect: isCorrect,
		}
	}

	formattedFiles := make([]string, len(apiQuestion.Files))
	for _, file := range apiQuestion.Files {
		imagePath, err := formatEnemURL(&file, "/assets")
		if err != nil {
			log.Fatalf("Error formatting URL for alternative: %v", err)
		}
		formattedFiles = append(formattedFiles, *imagePath)
	}

	return CreateQuestionInput{
		Title:                    stringToPtr(apiQuestion.Title),
		Index:                    apiQuestion.Index,
		Area:                     apiQuestion.Discipline, // Mapeado da disciplina da API
		Discipline:               stringToPtr(apiQuestion.Discipline),
		Language:                 apiQuestion.Language,
		Year:                     apiQuestion.Year,
		Context:                  stringToPtr(apiQuestion.Context),
		AlternativesIntroduction: stringToPtr(apiQuestion.AlternativesIntroduction),
		Images:                   formattedFiles,
		Alternatives:             alternatives,
	}
}

func stringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func formatEnemURL(urlString *string, basePath string) (*string, error) {
	if urlString == nil || *urlString == "" {
		return nil, nil
	}

	if basePath == "" {
		basePath = "/assets"
	}

	parsedURL, err := url.Parse(*urlString)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, errors.New("Invalid URL from image path.")
	}

	cleanPath := strings.TrimSuffix(basePath, "/")
	result := fmt.Sprintf("%s%s", cleanPath, parsedURL.Path)

	return &result, nil
}
