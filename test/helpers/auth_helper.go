package helpers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type AuthHelper struct {
	router *gin.Engine
	t      *testing.T
}

func NewAuthHelper(router *gin.Engine, t *testing.T) *AuthHelper {
	return &AuthHelper{
		router: router,
		t:      t,
	}
}

type TestUser struct {
	ID       uint   `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Token    string `json:"token"`
}

func (ah *AuthHelper) RegisterUser(email, username, fullName, password string) *TestUser {
	reqBody := map[string]string{
		"email":     email,
		"username":  username,
		"full_name": fullName,
		"password":  password,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ah.router.ServeHTTP(w, req)

	require.Equal(ah.t, http.StatusCreated, w.Code, "Registration should succeed")

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			User struct {
				ID       uint   `json:"id"`
				Email    string `json:"email"`
				Username string `json:"username"`
				FullName string `json:"full_name"`
			} `json:"user"`
			Token     string `json:"token"`
			ExpiresAt string `json:"expires_at"`
		} `json:"data"`
	}

	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(ah.t, err, "Should unmarshal response")
	require.True(ah.t, response.Success, "Response should be successful")

	return &TestUser{
		ID:       response.Data.User.ID,
		Email:    response.Data.User.Email,
		Username: response.Data.User.Username,
		FullName: response.Data.User.FullName,
		Token:    response.Data.Token,
	}
}

func (ah *AuthHelper) LoginUser(email, password string) *TestUser {
	reqBody := map[string]string{
		"email":    email,
		"password": password,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ah.router.ServeHTTP(w, req)

	require.Equal(ah.t, http.StatusOK, w.Code, "Login should succeed")

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			User struct {
				ID       uint   `json:"id"`
				Email    string `json:"email"`
				Username string `json:"username"`
				FullName string `json:"full_name"`
			} `json:"user"`
			Token string `json:"token"`
		} `json:"data"`
	}

	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(ah.t, err)
	require.True(ah.t, response.Success)

	return &TestUser{
		ID:       response.Data.User.ID,
		Email:    response.Data.User.Email,
		Username: response.Data.User.Username,
		FullName: response.Data.User.FullName,
		Token:    response.Data.Token,
	}
}

func (ah *AuthHelper) MakeAuthenticatedRequest(method, path, token string, body interface{}) *httptest.ResponseRecorder {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}

	req := httptest.NewRequest(method, path, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	ah.router.ServeHTTP(w, req)
	return w
}
