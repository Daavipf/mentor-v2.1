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
	"mentor-api/internal/usecase/auth"
	"mentor-api/internal/usecase/exam"
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

	firstQuestion := questions[0].(map[string]interface{})
	firstQuestionID := firstQuestion["question_id"].(string)
	firstAlts := firstQuestion["alternatives"].([]interface{})
	firstAltID := firstAlts[0].(map[string]interface{})["id"].(string)

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

	// 4. User 2 attempts to answer User 1's question -> MUST BE 403 Forbidden
	answerPayload := exam.AnswerQuestionInput{
		QuestionID:            firstQuestionID,
		SelectedAlternativeID: firstAltID,
	}
	ansBody, _ := json.Marshal(answerPayload)
	reqAnsU2 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/exams/%s/answers", examID), bytes.NewBuffer(ansBody))
	reqAnsU2.Header.Set("Authorization", "Bearer "+tokenUser2)
	reqAnsU2.Header.Set("Content-Type", "application/json")
	wAnsU2 := httptest.NewRecorder()
	app.Router.ServeHTTP(wAnsU2, reqAnsU2)
	assert.Equal(t, http.StatusForbidden, wAnsU2.Code)

	// 5. User 1 submits an answer -> 200 OK
	reqAnsU1 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/exams/%s/answers", examID), bytes.NewBuffer(ansBody))
	reqAnsU1.Header.Set("Authorization", "Bearer "+tokenUser1)
	reqAnsU1.Header.Set("Content-Type", "application/json")
	wAnsU1 := httptest.NewRecorder()
	app.Router.ServeHTTP(wAnsU1, reqAnsU1)
	assert.Equal(t, http.StatusOK, wAnsU1.Code)

	// 6. User 2 attempts to finish User 1's exam -> MUST BE 403 Forbidden
	reqFinishU2 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/exams/%s/finish", examID), nil)
	reqFinishU2.Header.Set("Authorization", "Bearer "+tokenUser2)
	wFinishU2 := httptest.NewRecorder()
	app.Router.ServeHTTP(wFinishU2, reqFinishU2)
	assert.Equal(t, http.StatusForbidden, wFinishU2.Code)

	// 7. User 1 finishes the exam -> 200 OK
	reqFinishU1 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/exams/%s/finish", examID), nil)
	reqFinishU1.Header.Set("Authorization", "Bearer "+tokenUser1)
	wFinishU1 := httptest.NewRecorder()
	app.Router.ServeHTTP(wFinishU1, reqFinishU1)
	assert.Equal(t, http.StatusOK, wFinishU1.Code)

	var finishResp response.APIResponse
	_ = json.Unmarshal(wFinishU1.Body.Bytes(), &finishResp)
	finishData := finishResp.Data.(map[string]interface{})
	assert.Equal(t, examID, finishData["exam_id"])
	assert.NotNil(t, finishData["score"])

	// 8. User 1 lists histories -> contains finished exam
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

	// 9. Verify that Question and Alternative endpoints DO NOT EXIST on HTTP
	reqQuestionHTTP := httptest.NewRequest(http.MethodGet, "/api/v1/questions", nil)
	wQuestionHTTP := httptest.NewRecorder()
	app.Router.ServeHTTP(wQuestionHTTP, reqQuestionHTTP)
	assert.Equal(t, http.StatusNotFound, wQuestionHTTP.Code, "Questions must NOT be exposed via HTTP endpoints")

	reqAltHTTP := httptest.NewRequest(http.MethodGet, "/api/v1/alternatives", nil)
	wAltHTTP := httptest.NewRecorder()
	app.Router.ServeHTTP(wAltHTTP, reqAltHTTP)
	assert.Equal(t, http.StatusNotFound, wAltHTTP.Code, "Alternatives must NOT be exposed via HTTP endpoints")
}

