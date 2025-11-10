package user_profile_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/your-org/go-monorepo-boilerplate/servers/internal/feature/user_profile/update_profile"
	"github.com/your-org/go-monorepo-boilerplate/servers/test/helpers"
)

// TestUpdateProfile_Success는 사용자 프로필 수정이 성공적으로 수행되는지 검증합니다.
//
// 엔드포인트: POST /v1/user-profile/update
// 관련 파일: internal/feature/user_profile/update_profile/
//
// 테스트 의도:
//   - username과 display_name을 수정할 수 있는지 확인
//   - 수정된 데이터가 데이터베이스에 올바르게 저장되는지 검증
//   - updated_at이 자동으로 업데이트되는지 확인
//
// 테스트 시나리오:
//  1. users 테이블에 테스트 사용자 생성
//  2. username과 display_name을 새로운 값으로 update_profile 요청 전송
//  3. 응답 데이터 확인
//  4. 데이터베이스에서 직접 조회하여 변경사항 확인
//
// 기대 결과:
//   - HTTP 200 OK 응답
//   - 응답에 수정된 username과 display_name 포함
//   - updated_at이 변경됨
//   - 데이터베이스 데이터와 일치
func (s *UserProfileTestSuite) TestUpdateProfile_Success() {
	// Given: Create a test user
	publicID := pgtype.UUID{}
	err := publicID.Scan("550e8400-e29b-41d4-a716-446655440000")
	s.Require().NoError(err)

	user, err := s.Fixtures.CreateUser(s.Ctx, map[string]any{
		"public_id":    publicID,
		"email":        "test@example.com",
		"username":     "oldusername",
		"display_name": "Old Name",
	})
	s.Require().NoError(err)

	// When: Update profile
	newUsername := "newusername"
	newDisplayName := "New Display Name"
	reqBody := map[string]any{
		"public_id":    publicID,
		"username":     newUsername,
		"display_name": newDisplayName,
	}
	req := helpers.MustCreateJSONRequest(http.MethodPost, "/v1/user-profile/update", reqBody)
	w := httptest.NewRecorder()
	update_profile.Map(w, req)

	// Then: Verify response
	s.Equal(http.StatusOK, w.Code, "Expected 200 OK status")

	response, err := helpers.DecodeStandardResponse[update_profile.UpdateProfileResponse](w)
	s.Require().NoError(err)

	// Verify updated data
	s.Equal(newUsername, response.Data.Username, "username should be updated")
	s.Equal(newDisplayName, response.Data.DisplayName, "display_name should be updated")
	s.Equal(user.Email, response.Data.Email, "email should remain unchanged")
	s.True(response.Data.UpdatedAt.After(user.UpdatedAt.Time), "updated_at should be newer")

	// Verify database was actually updated
	updatedUser, err := s.Fixtures.GetUserByPublicID(s.Ctx, publicID)
	s.Require().NoError(err)
	s.Equal(newUsername, updatedUser.Username, "database username should be updated")
	s.Equal(newDisplayName, updatedUser.DisplayName.String, "database display_name should be updated")
}

// TestUpdateProfile_PartialUpdate는 일부 필드만 수정할 수 있는지 검증합니다.
//
// 엔드포인트: POST /v1/user-profile/update
// 관련 파일: internal/feature/user_profile/update_profile/
//
// 테스트 의도:
//   - username만 수정 시 display_name은 그대로 유지되는지 확인
//   - COALESCE 로직이 올바르게 동작하는지 검증
//
// 테스트 시나리오:
//  1. users 테이블에 테스트 사용자 생성
//  2. username만 포함한 update_profile 요청 전송
//  3. display_name이 변경되지 않았는지 확인
//
// 기대 결과:
//   - HTTP 200 OK 응답
//   - username만 변경됨
//   - display_name은 이전 값 유지
func (s *UserProfileTestSuite) TestUpdateProfile_PartialUpdate() {
	// Given: Create a test user
	publicID := pgtype.UUID{}
	err := publicID.Scan("550e8400-e29b-41d4-a716-446655440001")
	s.Require().NoError(err)

	user, err := s.Fixtures.CreateUser(s.Ctx, map[string]interface{}{
		"public_id":    publicID,
		"email":        "partial@example.com",
		"username":     "oldusername",
		"display_name": "Original Display Name",
	})
	s.Require().NoError(err)

	// When: Update only username
	newUsername := "updatedusername"
	reqBody := map[string]interface{}{
		"public_id": publicID,
		"username":  newUsername,
		// display_name is intentionally omitted
	}
	req := helpers.MustCreateJSONRequest(http.MethodPost, "/v1/user-profile/update", reqBody)
	w := httptest.NewRecorder()
	update_profile.Map(w, req)

	// Then: Verify response
	s.Equal(http.StatusOK, w.Code, "Expected 200 OK status")

	response, err := helpers.DecodeStandardResponse[update_profile.UpdateProfileResponse](w)
	s.Require().NoError(err)

	// Verify only username was updated
	s.Equal(newUsername, response.Data.Username, "username should be updated")
	s.Equal(user.DisplayName.String, response.Data.DisplayName, "display_name should remain unchanged")
}

// TestUpdateProfile_TransactionRollback는 트랜잭션 롤백이 올바르게 동작하는지 검증합니다.
//
// 엔드포인트: POST /v1/user-profile/update
// 관련 파일: internal/feature/user_profile/update_profile/
//
// 테스트 의도:
//   - 데이터베이스 제약 조건 위반 시 트랜잭션이 롤백되는지 확인
//   - 이전 데이터가 그대로 유지되는지 검증
//
// 테스트 시나리오:
//  1. users 테이블에 두 명의 사용자 생성
//  2. 첫 번째 사용자를 두 번째 사용자의 username으로 변경 시도 (unique 제약 위반)
//  3. 에러 응답 확인
//  4. 첫 번째 사용자의 데이터가 변경되지 않았는지 확인
//
// 기대 결과:
//   - HTTP 4xx 또는 5xx 에러 응답
//   - 첫 번째 사용자의 username이 변경되지 않음 (트랜잭션 롤백)
func (s *UserProfileTestSuite) TestUpdateProfile_TransactionRollback() {
	// Given: Create two users
	publicID1 := pgtype.UUID{}
	err := publicID1.Scan("550e8400-e29b-41d4-a716-446655440002")
	s.Require().NoError(err)

	publicID2 := pgtype.UUID{}
	err = publicID2.Scan("550e8400-e29b-41d4-a716-446655440003")
	s.Require().NoError(err)

	user1, err := s.Fixtures.CreateUser(s.Ctx, map[string]interface{}{
		"public_id":    publicID1,
		"email":        "user1@example.com",
		"username":     "user1",
		"display_name": "User One",
	})
	s.Require().NoError(err)

	_, err = s.Fixtures.CreateUser(s.Ctx, map[string]interface{}{
		"public_id":    publicID2,
		"email":        "user2@example.com",
		"username":     "user2",
		"display_name": "User Two",
	})
	s.Require().NoError(err)

	// When: Try to update user1's username to user2's username (unique constraint violation)
	reqBody := map[string]interface{}{
		"public_id": publicID1,
		"username":  "user2", // This should fail due to unique constraint
	}
	req := helpers.MustCreateJSONRequest(http.MethodPost, "/v1/user-profile/update", reqBody)
	w := httptest.NewRecorder()
	update_profile.Map(w, req)

	// Then: Verify error response
	s.NotEqual(http.StatusOK, w.Code, "Should return error status")

	// Verify user1's data was not changed (transaction rollback)
	unchangedUser, err := s.Fixtures.GetUserByPublicID(s.Ctx, publicID1)
	s.Require().NoError(err)
	s.Equal(user1.Username, unchangedUser.Username, "username should not be changed due to rollback")
	s.Equal(user1.DisplayName.String, unchangedUser.DisplayName.String, "display_name should not be changed")
}
