package test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AuthTestSuite struct {
	BaseTestSuite
}

func (suite *AuthTestSuite) TestUserRegistration() {
	suite.Run("successful registration", func() {
		user := suite.AuthHelper.RegisterUser(
			"test@example.com",
			"testuser",
			"Test User",
			"password123",
		)

		assert.NotZero(suite.T(), user.ID)
		assert.Equal(suite.T(), "test@example.com", user.Email)
		assert.Equal(suite.T(), "testuser", user.Username)
		assert.Equal(suite.T(), "Test User", user.FullName)
		assert.NotEmpty(suite.T(), user.Token)
	})

	suite.Run("duplicate email registration", func() {

		suite.AuthHelper.RegisterUser(
			"duplicate@example.com",
			"user1",
			"User One",
			"password123",
		)

		reqBody := map[string]string{
			"email":     "duplicate@example.com",
			"username":  "user2",
			"full_name": "User Two",
			"password":  "password123",
		}

		w := suite.AuthHelper.MakeAuthenticatedRequest("POST", "/api/v1/auth/register", "", reqBody)
		assert.Equal(suite.T(), http.StatusConflict, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(suite.T(), err)
		assert.False(suite.T(), response["success"].(bool))
	})

	suite.Run("invalid email format", func() {
		reqBody := map[string]string{
			"email":     "invalid-email",
			"username":  "testuser",
			"full_name": "Test User",
			"password":  "password123",
		}

		w := suite.AuthHelper.MakeAuthenticatedRequest("POST", "/api/v1/auth/register", "", reqBody)
		assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	})
}

func (suite *AuthTestSuite) TestUserLogin() {

	email := "login@example.com"
	password := "password123"
	suite.AuthHelper.RegisterUser(email, "loginuser", "Login User", password)

	suite.Run("successful login", func() {
		user := suite.AuthHelper.LoginUser(email, password)

		assert.NotZero(suite.T(), user.ID)
		assert.Equal(suite.T(), email, user.Email)
		assert.NotEmpty(suite.T(), user.Token)
	})

	suite.Run("invalid credentials", func() {
		reqBody := map[string]string{
			"email":    email,
			"password": "wrongpassword",
		}

		w := suite.AuthHelper.MakeAuthenticatedRequest("POST", "/api/v1/auth/login", "", reqBody)
		assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(suite.T(), err)
		assert.False(suite.T(), response["success"].(bool))
	})

	suite.Run("missing credentials", func() {
		reqBody := map[string]string{
			"email": email,
		}

		w := suite.AuthHelper.MakeAuthenticatedRequest("POST", "/api/v1/auth/login", "", reqBody)
		assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	})
}

func (suite *AuthTestSuite) TestProtectedEndpoints() {

	user := suite.AuthHelper.RegisterUser(
		"protected@example.com",
		"protecteduser",
		"Protected User",
		"password123",
	)

	suite.Run("access with valid token", func() {
		w := suite.AuthHelper.MakeAuthenticatedRequest("GET", "/api/v1/users/profile", user.Token, nil)
		assert.Equal(suite.T(), http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(suite.T(), err)
		assert.True(suite.T(), response["success"].(bool))

		data := response["data"].(map[string]interface{})
		assert.Equal(suite.T(), user.Email, data["email"])
	})

	suite.Run("access without token", func() {
		w := suite.AuthHelper.MakeAuthenticatedRequest("GET", "/api/v1/users/profile", "", nil)
		assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	})

	suite.Run("access with invalid token", func() {
		w := suite.AuthHelper.MakeAuthenticatedRequest("GET", "/api/v1/users/profile", "invalid-token", nil)
		assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	})
}

func (suite *AuthTestSuite) TestTokenRefresh() {

	user := suite.AuthHelper.RegisterUser(
		"refresh@example.com",
		"refreshuser",
		"Refresh User",
		"password123",
	)

	suite.Run("valid token refresh", func() {
		reqBody := map[string]string{
			"token": user.Token,
		}

		w := suite.AuthHelper.MakeAuthenticatedRequest("POST", "/api/v1/auth/refresh", "", reqBody)
		assert.Equal(suite.T(), http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(suite.T(), err)
		assert.True(suite.T(), response["success"].(bool))

		data := response["data"].(map[string]interface{})
		newToken := data["token"].(string)
		assert.NotEmpty(suite.T(), newToken)
		assert.NotEqual(suite.T(), user.Token, newToken)
	})

	suite.Run("invalid token refresh", func() {
		reqBody := map[string]string{
			"token": "invalid-token",
		}

		w := suite.AuthHelper.MakeAuthenticatedRequest("POST", "/api/v1/auth/refresh", "", reqBody)
		assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	})
}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}
