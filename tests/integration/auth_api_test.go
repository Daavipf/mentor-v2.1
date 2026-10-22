package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"mentor-api/internal/interface/http/response"
	"mentor-api/internal/usecase/auth"
)

func TestIntegration_Auth_Flow(t *testing.T) {
	app := SetupTestApp()

	// 1. Register User
	registerPayload := auth.RegisterInput{
		Name:     "Integration User",
		Email:    "integration@test.com",
		Password: "password123",
	}
	body, _ := json.Marshal(registerPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var regResp response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &regResp)
	require.NoError(t, err)
	assert.True(t, regResp.Success)

	dataMap := regResp.Data.(map[string]interface{})
	token := dataMap["token"].(string)
	assert.NotEmpty(t, token)

	// 2. Access /api/v1/auth/me without token -> 401 Unauthorized
	reqMeNoAuth := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	wMeNoAuth := httptest.NewRecorder()
	app.Router.ServeHTTP(wMeNoAuth, reqMeNoAuth)
	assert.Equal(t, http.StatusUnauthorized, wMeNoAuth.Code)

	// 3. Access /api/v1/auth/me with Bearer token -> 200 OK
	reqMe := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	reqMe.Header.Set("Authorization", "Bearer "+token)
	wMe := httptest.NewRecorder()
	app.Router.ServeHTTP(wMe, reqMe)
	require.Equal(t, http.StatusOK, wMe.Code)

	var meResp response.APIResponse
	err = json.Unmarshal(wMe.Body.Bytes(), &meResp)
	require.NoError(t, err)
	meData := meResp.Data.(map[string]interface{})
	assert.Equal(t, "integration@test.com", meData["email"])

	// 4. Login with correct credentials
	loginPayload := auth.LoginInput{
		Email:    "integration@test.com",
		Password: "password123",
	}
	loginBody, _ := json.Marshal(loginPayload)
	reqLogin := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(loginBody))
	reqLogin.Header.Set("Content-Type", "application/json")
	wLogin := httptest.NewRecorder()
	app.Router.ServeHTTP(wLogin, reqLogin)
	assert.Equal(t, http.StatusOK, wLogin.Code)

	// 5. Login with incorrect password -> 401 Unauthorized
	badLoginPayload := auth.LoginInput{
		Email:    "integration@test.com",
		Password: "wrong_password",
	}
	badBody, _ := json.Marshal(badLoginPayload)
	reqBadLogin := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(badBody))
	reqBadLogin.Header.Set("Content-Type", "application/json")
	wBadLogin := httptest.NewRecorder()
	app.Router.ServeHTTP(wBadLogin, reqBadLogin)
	assert.Equal(t, http.StatusUnauthorized, wBadLogin.Code)
}

