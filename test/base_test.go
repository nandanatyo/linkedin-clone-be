package test

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"linked-clone/internal/config/server"
	"linked-clone/pkg/logger"
	testConfig "linked-clone/test/config"
	testDB "linked-clone/test/database"
	"linked-clone/test/helpers"
	"os"
)

type BaseTestSuite struct {
	suite.Suite
	Router     *gin.Engine
	TestDB     *testDB.TestDB
	AuthHelper *helpers.AuthHelper
}

func (suite *BaseTestSuite) SetupSuite() {

	os.Setenv("APP_ENV", "test")

	gin.SetMode(gin.TestMode)

	cfg := testConfig.LoadTestConfig()

	db, err := testDB.NewTestDB(cfg)
	suite.Require().NoError(err, "Failed to setup test database")
	suite.TestDB = db

	structuredLogger := logger.NewStructuredLogger()

	srv, err := server.NewServer(cfg, db.DB, structuredLogger)
	suite.Require().NoError(err, "Failed to create test server")

	suite.Router = srv.GetRouter()
	suite.AuthHelper = helpers.NewAuthHelper(suite.Router, suite.T())
}

func (suite *BaseTestSuite) TearDownSuite() {
	if suite.TestDB != nil {
		suite.TestDB.Close()
	}
}

func (suite *BaseTestSuite) SetupTest() {

	err := suite.TestDB.Clean()
	suite.Require().NoError(err, "Failed to clean test database")
}
