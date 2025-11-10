package helpers

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
)

// BaseIntegrationTestSuite provides common setup/teardown for integration tests
// Embed this in your test suites to automatically get container management
type BaseIntegrationTestSuite struct {
	suite.Suite
	Containers *TestContainers
	Fixtures   *TestFixtures
	Ctx        context.Context
	Redis      *redis.Client // Direct access to Redis client for test fixtures
}

// SetupSuite initializes test containers and fixtures before all tests in the suite
func (s *BaseIntegrationTestSuite) SetupSuite() {
	s.Ctx = context.Background()

	// Start containers
	tc, err := SetupTestContainers(s.Ctx)
	s.Require().NoError(err, "Failed to setup test containers")
	s.Containers = tc
	s.Fixtures = NewTestFixtures(tc.DBPool)
	s.Redis = tc.RedisClient

	// Set environment variables for singletons to use
	err = tc.SetEnvironmentVariables()
	s.Require().NoError(err, "Failed to set environment variables")
}

// TearDownSuite cleans up test containers after all tests in the suite
func (s *BaseIntegrationTestSuite) TearDownSuite() {
	if s.Containers != nil {
		err := s.Containers.Cleanup(s.Ctx)
		s.Require().NoError(err, "Failed to cleanup containers")
	}
}

// SetupTest cleans the database and Redis before each test
func (s *BaseIntegrationTestSuite) SetupTest() {
	err := s.Containers.TruncateAllTables(s.Ctx)
	s.Require().NoError(err, "Failed to truncate tables")

	// Clean Redis database
	err = s.Redis.FlushDB(s.Ctx).Err()
	s.Require().NoError(err, "Failed to flush Redis database")
}
