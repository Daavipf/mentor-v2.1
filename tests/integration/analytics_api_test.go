package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mentor-api/internal/interface/http/response"
	"mentor-api/internal/usecase/exam"
)

func TestIntegration_Analytics_Unauthorized(t *testing.T) {
	app := SetupTestApp()

	endpoints := []string{
		"/api/v1/analytics/overview",
		"/api/v1/analytics/by-discipline",
		"/api/v1/analytics/topics-ranking",
		"/api/v1/analytics/evolution",
		"/api/v1/analytics/exams-history",
	}

	for _, ep := range endpoints {
		req := httptest.NewRequest(http.MethodGet, ep, nil)
		w := httptest.NewRecorder()
		app.Router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code, "Expected 401 for unauthenticated request to %s", ep)
	}
}

func TestIntegration_Analytics_FullFlow_And_UserIsolation(t *testing.T) {
	app := SetupTestApp()

	tokenUser1 := registerAndGetToken(t, app, "Analytics User 1", "analytics1@test.com")
	tokenUser2 := registerAndGetToken(t, app, "Analytics User 2", "analytics2@test.com")

	// 1. User 1 creates an exam
	createPayload := exam.CreateExamInput{
		Title:         "Simulado de Ciências Humanas",
		Areas:         []string{"Ciências Humanas"},
		QuestionCount: 2,
	}
	body, _ := json.Marshal(createPayload)
	reqCreate := httptest.NewRequest(http.MethodPost, "/api/v1/exams", bytes.NewBuffer(body))
	reqCreate.Header.Set("Authorization", "Bearer "+tokenUser1)
	reqCreate.Header.Set("Content-Type", "application/json")
	wCreate := httptest.NewRecorder()
	app.Router.ServeHTTP(wCreate, reqCreate)
	require.Equal(t, http.StatusCreated, wCreate.Code)

	var createResp response.APIResponse
	err := json.Unmarshal(wCreate.Body.Bytes(), &createResp)
	require.NoError(t, err)

	examData := createResp.Data.(map[string]interface{})
	examID := examData["id"].(string)
	questions := examData["questions"].([]interface{})
	require.NotEmpty(t, questions)

	// User 1 answers all questions
	for _, qItem := range questions {
		qMap := qItem.(map[string]interface{})
		qID := qMap["question_id"].(string)
		alts := qMap["alternatives"].([]interface{})
		firstAltID := alts[0].(map[string]interface{})["id"].(string)

		ansPayload := exam.AnswerQuestionInput{
			QuestionID:            qID,
			SelectedAlternativeID: firstAltID,
		}
		ansBody, _ := json.Marshal(ansPayload)
		reqAns := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/exams/%s/answers", examID), bytes.NewBuffer(ansBody))
		reqAns.Header.Set("Authorization", "Bearer "+tokenUser1)
		reqAns.Header.Set("Content-Type", "application/json")
		wAns := httptest.NewRecorder()
		app.Router.ServeHTTP(wAns, reqAns)
		assert.Equal(t, http.StatusOK, wAns.Code)
	}

	// User 1 finishes the exam
	reqFinish := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/exams/%s/finish", examID), nil)
	reqFinish.Header.Set("Authorization", "Bearer "+tokenUser1)
	wFinish := httptest.NewRecorder()
	app.Router.ServeHTTP(wFinish, reqFinish)
	assert.Equal(t, http.StatusOK, wFinish.Code)

	// 2. User 1 calls GET /api/v1/analytics/overview
	reqOverview := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/overview", nil)
	reqOverview.Header.Set("Authorization", "Bearer "+tokenUser1)
	wOverview := httptest.NewRecorder()
	app.Router.ServeHTTP(wOverview, reqOverview)
	require.Equal(t, http.StatusOK, wOverview.Code)

	var overviewResp response.APIResponse
	err = json.Unmarshal(wOverview.Body.Bytes(), &overviewResp)
	require.NoError(t, err)
	assert.True(t, overviewResp.Success)
	overviewData := overviewResp.Data.(map[string]interface{})
	assert.Equal(t, float64(len(questions)), overviewData["total_questions_answered"])
	assert.Equal(t, float64(1), overviewData["total_completed_exams"])

	// 3. User 1 calls GET /api/v1/analytics/by-discipline
	reqDiscipline := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/by-discipline", nil)
	reqDiscipline.Header.Set("Authorization", "Bearer "+tokenUser1)
	wDiscipline := httptest.NewRecorder()
	app.Router.ServeHTTP(wDiscipline, reqDiscipline)
	require.Equal(t, http.StatusOK, wDiscipline.Code)

	var discResp response.APIResponse
	err = json.Unmarshal(wDiscipline.Body.Bytes(), &discResp)
	require.NoError(t, err)
	discData := discResp.Data.(map[string]interface{})
	assert.NotEmpty(t, discData["disciplines"])

	// 4. User 1 calls GET /api/v1/analytics/topics-ranking
	reqTopics := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/topics-ranking?min_questions=1", nil)
	reqTopics.Header.Set("Authorization", "Bearer "+tokenUser1)
	wTopics := httptest.NewRecorder()
	app.Router.ServeHTTP(wTopics, reqTopics)
	require.Equal(t, http.StatusOK, wTopics.Code)

	var topicsResp response.APIResponse
	err = json.Unmarshal(wTopics.Body.Bytes(), &topicsResp)
	require.NoError(t, err)
	topicsData := topicsResp.Data.(map[string]interface{})
	assert.Equal(t, float64(1), topicsData["min_questions_threshold"])

	// 5. User 1 calls GET /api/v1/analytics/evolution
	reqEvolution := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/evolution?period=weekly", nil)
	reqEvolution.Header.Set("Authorization", "Bearer "+tokenUser1)
	wEvolution := httptest.NewRecorder()
	app.Router.ServeHTTP(wEvolution, reqEvolution)
	require.Equal(t, http.StatusOK, wEvolution.Code)

	var evolResp response.APIResponse
	err = json.Unmarshal(wEvolution.Body.Bytes(), &evolResp)
	require.NoError(t, err)
	evolData := evolResp.Data.(map[string]interface{})
	assert.Equal(t, "weekly", evolData["period_type"])
	assert.NotEmpty(t, evolData["points"])

	// 6. User 1 calls GET /api/v1/analytics/exams-history
	reqExamsHist := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/exams-history", nil)
	reqExamsHist.Header.Set("Authorization", "Bearer "+tokenUser1)
	wExamsHist := httptest.NewRecorder()
	app.Router.ServeHTTP(wExamsHist, reqExamsHist)
	require.Equal(t, http.StatusOK, wExamsHist.Code)

	var histResp response.APIResponse
	err = json.Unmarshal(wExamsHist.Body.Bytes(), &histResp)
	require.NoError(t, err)
	histData := histResp.Data.(map[string]interface{})
	assert.Equal(t, float64(1), histData["total_finished_exams"])

	// 7. USER ISOLATION CHECK: User 2 requests overview and exams history
	// User 2 MUST have 0 answered questions and 0 completed exams
	reqU2Overview := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/overview", nil)
	reqU2Overview.Header.Set("Authorization", "Bearer "+tokenUser2)
	wU2Overview := httptest.NewRecorder()
	app.Router.ServeHTTP(wU2Overview, reqU2Overview)
	require.Equal(t, http.StatusOK, wU2Overview.Code)

	var u2OverviewResp response.APIResponse
	err = json.Unmarshal(wU2Overview.Body.Bytes(), &u2OverviewResp)
	require.NoError(t, err)
	u2OverviewData := u2OverviewResp.Data.(map[string]interface{})
	assert.Equal(t, float64(0), u2OverviewData["total_questions_answered"], "User 2 must not see User 1's answered questions")
	assert.Equal(t, float64(0), u2OverviewData["total_completed_exams"], "User 2 must not see User 1's completed exams")

	reqU2Hist := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/exams-history", nil)
	reqU2Hist.Header.Set("Authorization", "Bearer "+tokenUser2)
	wU2Hist := httptest.NewRecorder()
	app.Router.ServeHTTP(wU2Hist, reqU2Hist)
	require.Equal(t, http.StatusOK, wU2Hist.Code)

	var u2HistResp response.APIResponse
	err = json.Unmarshal(wU2Hist.Body.Bytes(), &u2HistResp)
	require.NoError(t, err)
	u2HistData := u2HistResp.Data.(map[string]interface{})
	assert.Equal(t, float64(0), u2HistData["total_finished_exams"], "User 2 must have 0 finished exams")
	u2ExamsList := u2HistData["exams"].([]interface{})
	assert.Empty(t, u2ExamsList)
}

