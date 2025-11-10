package user_profile_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/your-org/go-monorepo-boilerplate/servers/internal/feature/user_profile/get_profile"
	"github.com/your-org/go-monorepo-boilerplate/servers/test/helpers"
)

// TestGetProfile_Success는 사용자 프로필 조회가 성공적으로 수행되는지 검증합니다.
//
// 엔드포인트: POST /v1/user-profile/get
// 관련 파일: internal/feature/user_profile/get_profile/
//
// 테스트 의도:
//   - 유효한 public_id로 사용자 프로필을 조회할 수 있는지 확인
//   - 반환된 프로필 정보가 데이터베이스 데이터와 일치하는지 검증
//
// 테스트 시나리오:
//  1. users 테이블에 테스트 사용자 생성
//  2. public_id를 사용하여 get_profile 요청 전송
//  3. 응답 데이터 확인
//
// 기대 결과:
//   - HTTP 200 OK 응답
//   - 응답에 올바른 사용자 정보 포함 (email, username, display_name)
//   - created_at, updated_at 필드가 존재함
func (s *UserProfileTestSuite) TestGetProfile_Success() {
	// Given: Create a test user
	publicID := pgtype.UUID{}
	err := publicID.Scan("550e8400-e29b-41d4-a716-446655440000")
	s.Require().NoError(err)

	user, err := s.Fixtures.CreateUser(s.Ctx, map[string]interface{}{
		"public_id":    publicID,
		"email":        "test@example.com",
		"username":     "testuser",
		"display_name": "Test User",
	})
	s.Require().NoError(err)

	// When: Make get profile request
	reqBody := map[string]interface{}{
		"public_id": publicID,
	}
	req := helpers.MustCreateJSONRequest(http.MethodPost, "/v1/user-profile/get", reqBody)
	w := httptest.NewRecorder()
	get_profile.Map(w, req)

	// Then: Verify response
	s.Equal(http.StatusOK, w.Code, "Expected 200 OK status")

	response, err := helpers.DecodeStandardResponse[get_profile.GetProfileResponse](w)
	s.Require().NoError(err)

	// Verify profile data matches created user
	s.Equal(user.Email, response.Data.Email, "email should match")
	s.Equal(user.Username, response.Data.Username, "username should match")
	s.Equal(user.DisplayName, response.Data.DisplayName, "display_name should match")
	s.NotZero(response.Data.CreatedAt, "created_at should not be zero")
	s.NotZero(response.Data.UpdatedAt, "updated_at should not be zero")
}

// TestGetProfile_UserNotFound는 존재하지 않는 public_id로 조회 시 에러를 반환하는지 검증합니다.
//
// 엔드포인트: POST /v1/user-profile/get
// 관련 파일: internal/feature/user_profile/get_profile/
//
// 테스트 의도:
//   - 존재하지 않는 사용자 조회 시 적절한 에러 응답을 반환하는지 확인
//   - 에러 메시지가 명확한지 검증
//
// 테스트 시나리오:
//  1. 데이터베이스에 존재하지 않는 public_id 생성
//  2. 존재하지 않는 public_id로 get_profile 요청 전송
//  3. 에러 응답 확인
//
// 기대 결과:
//   - HTTP 4xx 또는 5xx 에러 응답
//   - 에러 메시지에 "not found" 포함
func (s *UserProfileTestSuite) TestGetProfile_UserNotFound() {
	// Given: Non-existent public_id
	publicID := pgtype.UUID{}
	err := publicID.Scan("550e8400-e29b-41d4-a716-446655440001")
	s.Require().NoError(err)

	// When: Make get profile request with non-existent user
	reqBody := map[string]interface{}{
		"public_id": publicID,
	}
	req := helpers.MustCreateJSONRequest(http.MethodPost, "/v1/user-profile/get", reqBody)
	w := httptest.NewRecorder()
	get_profile.Map(w, req)

	// Then: Verify error response
	s.NotEqual(http.StatusOK, w.Code, "Should return error status")

	// Verify error message contains "not found"
	response, err := helpers.DecodeErrorResponse(w)
	s.Require().NoError(err)
	s.Contains(response.Message, "not found", "error message should indicate user not found")
}

// TestGetProfile_SoftDeletedUser는 soft delete된 사용자가 조회되지 않는지 검증합니다.
//
// 엔드포인트: POST /v1/user-profile/get
// 관련 파일: internal/feature/user_profile/get_profile/
//
// 테스트 의도:
//   - deleted_at이 설정된 사용자는 조회되지 않아야 함
//   - Soft delete 로직이 올바르게 동작하는지 확인
//
// 테스트 시나리오:
//  1. users 테이블에 사용자 생성
//  2. 사용자를 soft delete 처리 (deleted_at 설정)
//  3. public_id로 get_profile 요청 전송
//  4. 에러 응답 확인
//
// 기대 결과:
//   - HTTP 4xx 에러 응답
//   - 사용자가 조회되지 않음
func (s *UserProfileTestSuite) TestGetProfile_SoftDeletedUser() {
	// Given: Create and soft delete a user
	publicID := pgtype.UUID{}
	err := publicID.Scan("550e8400-e29b-41d4-a716-446655440002")
	s.Require().NoError(err)

	_, err = s.Fixtures.CreateUser(s.Ctx, map[string]interface{}{
		"public_id":    publicID,
		"email":        "deleted@example.com",
		"username":     "deleteduser",
		"display_name": "Deleted User",
	})
	s.Require().NoError(err)

	// Soft delete the user
	err = s.Fixtures.SoftDeleteUser(s.Ctx, publicID)
	s.Require().NoError(err)

	// When: Try to get soft deleted user
	reqBody := map[string]interface{}{
		"public_id": publicID,
	}
	req := helpers.MustCreateJSONRequest(http.MethodPost, "/v1/user-profile/get", reqBody)
	w := httptest.NewRecorder()
	get_profile.Map(w, req)

	// Then: Verify user is not found
	s.NotEqual(http.StatusOK, w.Code, "Should return error for soft deleted user")
}
