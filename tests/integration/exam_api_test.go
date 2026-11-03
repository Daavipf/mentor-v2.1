package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mentor-api/internal/interface/http/response"
	"mentor-api/internal/usecase/auth"
	"mentor-api/internal/usecase/exam"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func registerAndGetToken(t *testing.T, app *TestApp, name, email string) string {
	payload := auth.RegisterInput{
		Name:     name,
		Email:    email,
		Password: "password123",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var resp response.APIResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	dataMap := resp.Data.(map[string]interface{})
	return dataMap["token"].(string)
}

func TestIntegration_Exam_And_Security_Isolation(t *testing.T) {
	app := SetupTestApp()

	tokenUser1 := registerAndGetToken(t, app, "User One", "user1@test.com")
	tokenUser2 := registerAndGetToken(t, app, "User Two", "user2@test.com")

	// 1. User 1 creates an exam
	createPayload := exam.CreateExamInput{
		Title:         "Simulado Ciências da Natureza",
		Areas:         []string{"Ciências da Natureza"},
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
	require.NotEmpty(t, examID)
	require.NotEmpty(t, questions)

	// Monta a lista de respostas iterando sobre as questões do exame criado
	var answersPayload []exam.AnswerQuestionInput
	now := time.Now()
	sat := now.Add(-10 * time.Minute)
	for _, qItem := range questions {
		qMap := qItem.(map[string]interface{})
		qID := qMap["question_id"].(string)
		alts := qMap["alternatives"].([]interface{})
		altID := alts[0].(map[string]interface{})["id"].(string)

		answersPayload = append(answersPayload, exam.AnswerQuestionInput{
			QuestionID:            qID,
			SelectedAlternativeID: altID,
			StartedAt:             &sat,
			FinishedAt:            &now,
		})
	}

	// 2. User 1 can view their exam
	reqGet := httptest.NewRequest(http.MethodGet, "/api/v1/exams/"+examID, nil)
	reqGet.Header.Set("Authorization", "Bearer "+tokenUser1)
	wGet := httptest.NewRecorder()
	app.Router.ServeHTTP(wGet, reqGet)
	assert.Equal(t, http.StatusOK, wGet.Code)

	// 3. User 2 attempts to view User 1's exam -> MUST BE 403 Forbidden
	reqGetU2 := httptest.NewRequest(http.MethodGet, "/api/v1/exams/"+examID, nil)
	reqGetU2.Header.Set("Authorization", "Bearer "+tokenUser2)
	wGetU2 := httptest.NewRecorder()
	app.Router.ServeHTTP(wGetU2, reqGetU2)
	assert.Equal(t, http.StatusForbidden, wGetU2.Code)

	// 4. User 2 attempts to submit User 1's exam -> MUST BE 403 Forbidden
	submitBody, _ := json.Marshal(answersPayload)
	reqSubU2 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/exams/%s/submit", examID), bytes.NewBuffer(submitBody))
	reqSubU2.Header.Set("Authorization", "Bearer "+tokenUser2)
	reqSubU2.Header.Set("Content-Type", "application/json")
	wSubU2 := httptest.NewRecorder()
	app.Router.ServeHTTP(wSubU2, reqSubU2)
	assert.Equal(t, http.StatusForbidden, wSubU2.Code)

	// 5. User 1 submits the exam -> 200 OK
	reqSubU1 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/exams/%s/submit", examID), bytes.NewBuffer(submitBody))
	reqSubU1.Header.Set("Authorization", "Bearer "+tokenUser1)
	reqSubU1.Header.Set("Content-Type", "application/json")
	wSubU1 := httptest.NewRecorder()
	app.Router.ServeHTTP(wSubU1, reqSubU1)
	assert.Equal(t, http.StatusOK, wSubU1.Code)

	var submitResp response.APIResponse
	_ = json.Unmarshal(wSubU1.Body.Bytes(), &submitResp)
	submitData := submitResp.Data.(map[string]interface{})
	assert.Equal(t, examID, submitData["exam_id"])
	assert.NotNil(t, submitData["score"])

	// 6. User 1 attempts to submit the same exam again -> MUST BE 400 Bad Request or 409 Conflict
	reqSubAgain := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/exams/%s/submit", examID), bytes.NewBuffer(submitBody))
	reqSubAgain.Header.Set("Authorization", "Bearer "+tokenUser1)
	reqSubAgain.Header.Set("Content-Type", "application/json")
	wSubAgain := httptest.NewRecorder()
	app.Router.ServeHTTP(wSubAgain, reqSubAgain)
	assert.NotEqual(t, http.StatusOK, wSubAgain.Code, "Re-submitting an already finished exam should fail")

	// 7. User 1 lists histories -> contains finished exam
	reqHistU1 := httptest.NewRequest(http.MethodGet, "/api/v1/histories", nil)
	reqHistU1.Header.Set("Authorization", "Bearer "+tokenUser1)
	wHistU1 := httptest.NewRecorder()
	app.Router.ServeHTTP(wHistU1, reqHistU1)
	assert.Equal(t, http.StatusOK, wHistU1.Code)

	var histResp response.APIResponse
	_ = json.Unmarshal(wHistU1.Body.Bytes(), &histResp)
	histList := histResp.Data.([]interface{})
	assert.Len(t, histList, 1)

	// User 2 lists histories -> empty
	reqHistU2 := httptest.NewRequest(http.MethodGet, "/api/v1/histories", nil)
	reqHistU2.Header.Set("Authorization", "Bearer "+tokenUser2)
	wHistU2 := httptest.NewRecorder()
	app.Router.ServeHTTP(wHistU2, reqHistU2)
	assert.Equal(t, http.StatusOK, wHistU2.Code)

	var histU2Resp response.APIResponse
	_ = json.Unmarshal(wHistU2.Body.Bytes(), &histU2Resp)
	histU2List := histU2Resp.Data.([]interface{})
	assert.Empty(t, histU2List)

	// 8. Verify that Question and Alternative endpoints DO NOT EXIST on HTTP
	reqQuestionHTTP := httptest.NewRequest(http.MethodGet, "/api/v1/questions", nil)
	wQuestionHTTP := httptest.NewRecorder()
	app.Router.ServeHTTP(wQuestionHTTP, reqQuestionHTTP)
	assert.Equal(t, http.StatusNotFound, wQuestionHTTP.Code, "Questions must NOT be exposed via HTTP endpoints")

	reqAltHTTP := httptest.NewRequest(http.MethodGet, "/api/v1/alternatives", nil)
	wAltHTTP := httptest.NewRecorder()
	app.Router.ServeHTTP(wAltHTTP, reqAltHTTP)
	assert.Equal(t, http.StatusNotFound, wAltHTTP.Code, "Alternatives must NOT be exposed via HTTP endpoints")
}
