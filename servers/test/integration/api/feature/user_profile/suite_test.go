package user_profile_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/your-org/go-monorepo-boilerplate/servers/test/helpers"
)

// UserProfileTestSuite is the integration test suite for user profile features
type UserProfileTestSuite struct {
	helpers.BaseIntegrationTestSuite
}

// TestUserProfileSuite runs the user profile test suite
func TestUserProfileSuite(t *testing.T) {
	suite.Run(t, new(UserProfileTestSuite))
}
