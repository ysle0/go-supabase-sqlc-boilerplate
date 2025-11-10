package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sqlc "github.com/your-org/go-monorepo-boilerplate/servers/internal/shared/database/sqlc/postgres"
)

// TestFixtures provides helper methods for creating test data
type TestFixtures struct {
	Pool    *pgxpool.Pool
	Queries *sqlc.Queries
}

// ChallengeModeResponse represents a challenge mode response with rewards
type ChallengeModeResponse struct {
	StageProgress int32
	CoinRewards   int32
	CrownRewards  int32
}

// NewTestFixtures creates a new TestFixtures instance
func NewTestFixtures(pool *pgxpool.Pool) *TestFixtures {
	return &TestFixtures{
		Pool:    pool,
		Queries: sqlc.New(pool),
	}
}

// CreateTestUser creates a test user with default values using raw SQL
func (f *TestFixtures) CreateTestUser(ctx context.Context, publicID string) (*sqlc.User, error) {
	var user sqlc.User

	// Generate a random UUID for oauth_id
	oauthID := uuid.New()
	// Parse publicID string to UUID
	publicUUID, err := uuid.Parse(publicID)
	if err != nil {
		return nil, fmt.Errorf("invalid public ID format: %w", err)
	}

	// Generate a test nickname using first 8 characters of publicID
	nickname := fmt.Sprintf("TestUser_%s", publicID[:8])

	query := `
		INSERT INTO users (oauth_id, public_id, nickname, crown_amount, coin_balance, country_code, gem_balance)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, oauth_id, public_id, nickname, profile_index, crown_amount, gem_balance, coin_balance,
		          hint_quantity, skip_quantity, country_code, is_deleted, created_at, updated_at, deleted_at
	`

	err = f.Pool.QueryRow(ctx, query, oauthID, publicUUID, nickname, 0, 100, "US", 0).Scan(
		&user.ID,
		&user.OauthID,
		&user.PublicID,
		&user.Nickname,
		&user.ProfileIndex,
		&user.CrownAmount,
		&user.GemBalance,
		&user.CoinBalance,
		&user.HintQuantity,
		&user.SkipQuantity,
		&user.CountryCode,
		&user.IsDeleted,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create test user: %w", err)
	}
	return &user, nil
}

// CreateTestCategory creates a test category using raw SQL
func (f *TestFixtures) CreateTestCategory(ctx context.Context, name string, displayOrder int32) (int32, error) {
	var categoryID int32
	query := `INSERT INTO categories (name, display_order) VALUES ($1, $2) RETURNING id`
	err := f.Pool.QueryRow(ctx, query, name, displayOrder).Scan(&categoryID)
	if err != nil {
		return 0, fmt.Errorf("failed to create test category: %w", err)
	}
	return categoryID, nil
}

// CreateTestGameType creates a test game type using raw SQL
func (f *TestFixtures) CreateTestGameType(ctx context.Context, name string, typeID int32) error {
	query := `INSERT INTO game_types (id, name, entrance_fee) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING`
	_, err := f.Pool.Exec(ctx, query, typeID, name, 10) // Default entrance fee of 10
	if err != nil {
		return fmt.Errorf("failed to create test game type: %w", err)
	}
	return nil
}

// CreateTestConfig creates a test default config using raw SQL
func (f *TestFixtures) CreateTestConfig(ctx context.Context, coinRewards int32, crownRewards int32) error {
	query := `INSERT INTO game_configs (
		id, coin_balance, gem_balance, coin_rewards, crown_rewards, max_stage_quantity,
		hint_quantity, skip_quantity, initial_game_time_left
	) OVERRIDING SYSTEM VALUE
	VALUES (1, 1000, 500, $1, $2, 4, 3, 3, 10)
	ON CONFLICT (id) DO UPDATE SET coin_rewards = $1, crown_rewards = $2, initial_game_time_left = 10`
	_, err := f.Pool.Exec(ctx, query, coinRewards, crownRewards)
	if err != nil {
		return fmt.Errorf("failed to create test config: %w", err)
	}
	return nil
}

// CreateTestStage creates a test stage for a user using SQLC
func (f *TestFixtures) CreateTestStage(ctx context.Context, userID int64, categoryID int32, stageNum int32, coinRewards int32, crownRewards int32, isClaimed bool) error {
	// Ensure classic mode game type exists (game_type_id = 0)
	_ = f.CreateTestGameType(ctx, "Classic", 0)

	// For classic mode (game_type_id = 0)
	_, err := f.Queries.CreateStage(ctx, sqlc.CreateStageParams{
		UserID:        userID,
		GameTypeID:    0, // Classic mode
		StageProgress: stageNum,
		CoinRewards:   coinRewards,
		CrownRewards:  crownRewards,
		CategoryID:    pgtype.Int4{Int32: categoryID, Valid: true},
		UpdatedAt:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to create test stage: %w", err)
	}

	// If the stage should be claimed, update it
	if isClaimed {
		_, err = f.Pool.Exec(ctx, `UPDATE stages SET is_claimed = true WHERE user_id = $1 AND category_id = $2`, userID, categoryID)
		if err != nil {
			return fmt.Errorf("failed to mark stage as claimed: %w", err)
		}
	}

	return nil
}

// CreateTestChallengeStage creates a test challenge mode stage using SQLC
// Uses the stages table with game_type_id to differentiate from classic mode
func (f *TestFixtures) CreateTestChallengeStage(
	ctx context.Context,
	userID int64,
	gameTypeID int32,
	stageNum int32,
	isClaimed bool,
) error {
	return f.CreateTestChallengeStageWithRewards(
		ctx,
		userID,
		gameTypeID,
		ChallengeModeResponse{
			StageProgress: stageNum,
		},
		isClaimed,
	)
}

// CreateTestChallengeStageWithRewards creates a test challenge mode stage with rewards using SQLC
func (f *TestFixtures) CreateTestChallengeStageWithRewards(
	ctx context.Context,
	userID int64,
	gameTypeID int32,
	reqBody ChallengeModeResponse,
	isClaimed bool,
) error {
	if _, err := f.Queries.CreateStagesWithoutCategory(ctx, sqlc.CreateStagesWithoutCategoryParams{
		UserID:        userID,
		GameTypeID:    gameTypeID,
		StageProgress: reqBody.StageProgress,
		CoinRewards:   reqBody.CoinRewards,
		CrownRewards:  reqBody.CrownRewards,
		UpdatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
	}); err != nil {
		return fmt.Errorf("failed to create test challenge stage with rewards: %w", err)
	}

	// If the stage should be claimed, update it
	if isClaimed {
		_, err := f.Pool.Exec(ctx, `UPDATE stages SET is_claimed = true WHERE user_id = $1 AND game_type_id = $2`, userID, gameTypeID)
		if err != nil {
			return fmt.Errorf("failed to mark stage as claimed: %w", err)
		}
	}

	return nil
}

// CreateTestQuestion creates a test question using raw SQL
func (f *TestFixtures) CreateTestQuestion(ctx context.Context, gameTypeID int32, categoryID int32, text string) (int32, error) {
	query := `INSERT INTO questions (game_type_id, category_id, text, explanation, choices, answer_index)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	// Set choices based on game type
	var choices []string
	switch gameTypeID {
	case 1: // True/False
		choices = []string{"True", "False"}
	default: // Classic Mode and other Challenge Modes (4 choices)
		choices = []string{"A", "B", "C", "D"}
	}

	var questionID int32
	err := f.Pool.QueryRow(ctx, query, gameTypeID, categoryID, text, "Test explanation", choices, 0).Scan(&questionID)
	if err != nil {
		return 0, fmt.Errorf("failed to create test question: %w", err)
	}
	return questionID, nil
}

// GetUserByID retrieves a user by ID using raw SQL
// Note: SQLC doesn't have a direct GetUserByID method, only GetUserByInternalId which takes an array
func (f *TestFixtures) GetUserByID(ctx context.Context, userID int64) (*sqlc.User, error) {
	var user sqlc.User
	query := `
		SELECT id, oauth_id, public_id, nickname, profile_index, crown_amount, gem_balance, coin_balance,
		       hint_quantity, skip_quantity, country_code, is_deleted, created_at, updated_at, deleted_at
		FROM users WHERE id = $1
	`
	err := f.Pool.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.OauthID,
		&user.PublicID,
		&user.Nickname,
		&user.ProfileIndex,
		&user.CrownAmount,
		&user.GemBalance,
		&user.CoinBalance,
		&user.HintQuantity,
		&user.SkipQuantity,
		&user.CountryCode,
		&user.IsDeleted,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return &user, nil
}

// GetChallengeStageClaimedStatus는 챌린지 모드 스테이지의 보상 수령 여부를 조회합니다.
func (f *TestFixtures) GetChallengeStageClaimedStatus(ctx context.Context, userID int64, gameTypeID int32) (bool, error) {
	stage, err := f.Queries.GetStageByUserIdAndGameTypeId(ctx, sqlc.GetStageByUserIdAndGameTypeIdParams{
		UserID:     userID,
		GameTypeID: gameTypeID,
	})
	if err != nil {
		return false, fmt.Errorf("failed to get challenge stage claimed status: %w", err)
	}
	return stage.IsClaimed, nil
}

// GetClassicStageClaimedStatus는 클래식 모드 스테이지의 보상 수령 여부를 조회합니다.
func (f *TestFixtures) GetClassicStageClaimedStatus(ctx context.Context, userID int64, categoryID int32) (bool, error) {
	stage, err := f.Queries.GetStageClassicMode(ctx, sqlc.GetStageClassicModeParams{
		UserID:     userID,
		GameTypeID: 0, // Classic mode
		CategoryID: pgtype.Int4{Int32: categoryID, Valid: true},
	})
	if err != nil {
		return false, fmt.Errorf("failed to get classic stage claimed status: %w", err)
	}
	return stage.IsClaimed, nil
}

// StageInfo는 스테이지 정보를 담는 구조체입니다.
type StageInfo struct {
	StageProgress int32
	CoinRewards   int32
	CrownRewards  int32
}

// GetChallengeStageInfo는 챌린지 모드 스테이지의 진행 정보를 조회합니다.
func (f *TestFixtures) GetChallengeStageInfo(ctx context.Context, userID int64, gameTypeID int32) (*StageInfo, error) {
	stage, err := f.Queries.GetStageByUserIdAndGameTypeId(ctx, sqlc.GetStageByUserIdAndGameTypeIdParams{
		UserID:     userID,
		GameTypeID: gameTypeID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get challenge stage info: %w", err)
	}
	return &StageInfo{
		StageProgress: stage.StageProgress,
		CoinRewards:   stage.CoinRewards,
		CrownRewards:  stage.CrownRewards,
	}, nil
}

// GetClassicStageInfo는 클래식 모드 스테이지의 진행 정보를 조회합니다.
func (f *TestFixtures) GetClassicStageInfo(ctx context.Context, userID int64, categoryID int32) (*StageInfo, error) {
	stage, err := f.Queries.GetStageClassicMode(ctx, sqlc.GetStageClassicModeParams{
		UserID:     userID,
		GameTypeID: 0, // Classic mode
		CategoryID: pgtype.Int4{Int32: categoryID, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get classic stage info: %w", err)
	}
	return &StageInfo{
		StageProgress: stage.StageProgress,
		CoinRewards:   stage.CoinRewards,
		CrownRewards:  stage.CrownRewards,
	}, nil
}

// UpdateUserCoinBalance는 사용자의 코인 잔액을 업데이트합니다.
func (f *TestFixtures) UpdateUserCoinBalance(ctx context.Context, userID int64, newBalance int32) error {
	err := f.Queries.UpdateCoinBalance(ctx, sqlc.UpdateCoinBalanceParams{
		ID:          userID,
		CoinBalance: newBalance,
	})
	if err != nil {
		return fmt.Errorf("failed to update user coin balance: %w", err)
	}
	return nil
}

// UpdateUserCrownAmount는 사용자의 크라운 수량을 업데이트합니다.
func (f *TestFixtures) UpdateUserCrownAmount(ctx context.Context, userID int64, newAmount int64) error {
	query := `UPDATE users SET crown_amount = $2 WHERE id = $1`
	_, err := f.Pool.Exec(ctx, query, userID, newAmount)
	if err != nil {
		return fmt.Errorf("failed to update user crown amount: %w", err)
	}
	return nil
}

// CountUsersByOAuthID는 특정 OAuth ID를 가진 사용자 수를 조회합니다.
func (f *TestFixtures) CountUsersByOAuthID(ctx context.Context, oauthID uuid.UUID) (int, error) {
	hasUser, err := f.Queries.HasUserByOAuthId(ctx, pgtype.UUID{Bytes: oauthID, Valid: true})
	if err != nil {
		return 0, fmt.Errorf("failed to count users by oauth_id: %w", err)
	}
	if hasUser {
		return 1, nil
	}
	return 0, nil
}

// UpdateUserGemBalance는 사용자의 젬 잔액을 업데이트합니다.
func (f *TestFixtures) UpdateUserGemBalance(ctx context.Context, userID int64, newBalance int32) error {
	err := f.Queries.UpdateGemBalance(ctx, sqlc.UpdateGemBalanceParams{
		ID:         userID,
		GemBalance: newBalance,
	})
	if err != nil {
		return fmt.Errorf("failed to update user gem balance: %w", err)
	}
	return nil
}

// UpdateUserHintQuantity는 사용자의 힌트 수량을 업데이트합니다.
func (f *TestFixtures) UpdateUserHintQuantity(ctx context.Context, userID int64, newQuantity int32) error {
	err := f.Queries.UpdateHintQuantity(ctx, sqlc.UpdateHintQuantityParams{
		ID:           userID,
		HintQuantity: newQuantity,
	})
	if err != nil {
		return fmt.Errorf("failed to update user hint quantity: %w", err)
	}
	return nil
}

// UpdateUserSkipQuantity는 사용자의 스킵 수량을 업데이트합니다.
func (f *TestFixtures) UpdateUserSkipQuantity(ctx context.Context, userID int64, newQuantity int32) error {
	err := f.Queries.UpdateSkipQuantity(ctx, sqlc.UpdateSkipQuantityParams{
		ID:           userID,
		SkipQuantity: newQuantity,
	})
	if err != nil {
		return fmt.Errorf("failed to update user skip quantity: %w", err)
	}
	return nil
}

// GetGameConfig는 기본 게임 설정을 조회합니다.
func (f *TestFixtures) GetGameConfig(ctx context.Context) (*sqlc.GameConfig, error) {
	config, err := f.Queries.GetDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get game config: %w", err)
	}
	return &config, nil
}

// CreateTestGameTypeWithFee는 특정 입장료를 가진 게임 타입을 생성합니다.
func (f *TestFixtures) CreateTestGameTypeWithFee(ctx context.Context, name string, typeID int32, entranceFee int32) error {
	query := `INSERT INTO game_types (id, name, entrance_fee) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING`
	_, err := f.Pool.Exec(ctx, query, typeID, name, entranceFee)
	if err != nil {
		return fmt.Errorf("failed to create test game type with fee: %w", err)
	}
	return nil
}

// CreateTestItem는 게임 아이템을 생성합니다 (힌트, 스킵 등).
func (f *TestFixtures) CreateTestItem(ctx context.Context, itemType string, price int32) error {
	query := `INSERT INTO items (name, item_type, price_type, price) VALUES ($1, $2, $3, $4)`
	name := itemType   // Use item type as name for test simplicity
	priceType := "gem" // Default to gem for all test items
	_, err := f.Pool.Exec(ctx, query, name, itemType, priceType, price)
	if err != nil {
		return fmt.Errorf("failed to create test item: %w", err)
	}
	return nil
}

// ===== Lucky Spin Fixtures =====

// CreateLSReward creates a lucky spin reward in the ls_rewards table
func (f *TestFixtures) CreateLSReward(ctx context.Context, rewardType string, quantity int32, spawnProbability int32) (int32, error) {
	var rewardID int32
	query := `INSERT INTO ls_rewards (type, quantity, spawn_probability)
	          VALUES ($1, $2, $3) RETURNING id`
	err := f.Pool.QueryRow(ctx, query, rewardType, quantity, spawnProbability).Scan(&rewardID)
	if err != nil {
		return 0, fmt.Errorf("failed to create ls_reward: %w", err)
	}
	return rewardID, nil
}

// CreateUserLSSpin creates a user lucky spin record in user_ls_spins table
func (f *TestFixtures) CreateUserLSSpin(ctx context.Context, userID int64, rewardID int32, wasAds bool, startedAt, endedAt time.Time) error {
	query := `INSERT INTO user_ls_spins (user_id, reward_id, was_ads, started_at, ended_at)
	          VALUES ($1, $2, $3, $4, $5)
	          ON CONFLICT (user_id) DO UPDATE
	          SET reward_id = $2, was_ads = $3, started_at = $4, ended_at = $5`
	_, err := f.Pool.Exec(ctx, query, userID, rewardID, wasAds, startedAt, endedAt)
	if err != nil {
		return fmt.Errorf("failed to create user_ls_spin: %w", err)
	}
	return nil
}

// GetUserLSSpin retrieves a user's lucky spin record
func (f *TestFixtures) GetUserLSSpin(ctx context.Context, userID int64) (*struct {
	ID        int64
	UserID    int64
	RewardID  int32
	WasAds    bool
	StartedAt time.Time
	EndedAt   time.Time
}, error) {
	var spin struct {
		ID        int64
		UserID    int64
		RewardID  int32
		WasAds    bool
		StartedAt time.Time
		EndedAt   time.Time
	}
	query := `SELECT id, user_id, reward_id, was_ads, started_at, ended_at
	          FROM user_ls_spins WHERE user_id = $1`
	err := f.Pool.QueryRow(ctx, query, userID).Scan(
		&spin.ID, &spin.UserID, &spin.RewardID, &spin.WasAds, &spin.StartedAt, &spin.EndedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user_ls_spin: %w", err)
	}
	return &spin, nil
}

// GetLSRewardByID retrieves a lucky spin reward by ID
func (f *TestFixtures) GetLSRewardByID(ctx context.Context, rewardID int32) (*struct {
	ID               int32
	Type             string
	Quantity         int32
	SpawnProbability int32
}, error) {
	var reward struct {
		ID               int32
		Type             string
		Quantity         int32
		SpawnProbability int32
	}
	query := `SELECT id, type, quantity, spawn_probability FROM ls_rewards WHERE id = $1`
	err := f.Pool.QueryRow(ctx, query, rewardID).Scan(
		&reward.ID, &reward.Type, &reward.Quantity, &reward.SpawnProbability,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get ls_reward: %w", err)
	}
	return &reward, nil
}

// GetAllLSRewards retrieves all lucky spin rewards
func (f *TestFixtures) GetAllLSRewards(ctx context.Context) ([]struct {
	ID               int32
	Type             string
	Quantity         int32
	SpawnProbability int32
}, error) {
	query := `SELECT id, type, quantity, spawn_probability FROM ls_rewards ORDER BY id`
	rows, err := f.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all ls_rewards: %w", err)
	}
	defer rows.Close()

	var rewards []struct {
		ID               int32
		Type             string
		Quantity         int32
		SpawnProbability int32
	}

	for rows.Next() {
		var reward struct {
			ID               int32
			Type             string
			Quantity         int32
			SpawnProbability int32
		}
		if err := rows.Scan(&reward.ID, &reward.Type, &reward.Quantity, &reward.SpawnProbability); err != nil {
			return nil, fmt.Errorf("failed to scan ls_reward: %w", err)
		}
		rewards = append(rewards, reward)
	}

	return rewards, nil
}

// GetLSClaimHistory retrieves all claim records for a user
func (f *TestFixtures) GetLSClaimHistory(ctx context.Context, userID int64) ([]struct {
	ID              int64
	UserID          int64
	RewardID        int32
	QuantityAtClaim int32
	TypeAtClaim     string
	ClaimedAt       time.Time
}, error) {
	query := `SELECT id, user_id, reward_id, quantity_at_claim, type_at_claim, claimed_at
	          FROM user_ls_claims WHERE user_id = $1 ORDER BY claimed_at DESC`
	rows, err := f.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ls_claim_history: %w", err)
	}
	defer rows.Close()

	var claims []struct {
		ID              int64
		UserID          int64
		RewardID        int32
		QuantityAtClaim int32
		TypeAtClaim     string
		ClaimedAt       time.Time
	}

	for rows.Next() {
		var claim struct {
			ID              int64
			UserID          int64
			RewardID        int32
			QuantityAtClaim int32
			TypeAtClaim     string
			ClaimedAt       time.Time
		}
		if err := rows.Scan(&claim.ID, &claim.UserID, &claim.RewardID, &claim.QuantityAtClaim, &claim.TypeAtClaim, &claim.ClaimedAt); err != nil {
			return nil, fmt.Errorf("failed to scan ls_claim: %w", err)
		}
		claims = append(claims, claim)
	}

	return claims, nil
}

// ===== Ranking Fixtures =====

// CreateRankingDate creates a ranking date record
func (f *TestFixtures) CreateRankingDate(ctx context.Context, dailyStart, dailyEnd, weeklyStart, weeklyEnd time.Time) error {
	query := `INSERT INTO ranking_dates (daily_started_at, daily_ended_at, weekly_started_at, weekly_ended_at)
	          VALUES ($1, $2, $3, $4)
	          ON CONFLICT (id) DO UPDATE
	          SET daily_started_at = $1, daily_ended_at = $2, weekly_started_at = $3, weekly_ended_at = $4`
	_, err := f.Pool.Exec(ctx, query, dailyStart, dailyEnd, weeklyStart, weeklyEnd)
	if err != nil {
		return fmt.Errorf("failed to create ranking_date: %w", err)
	}
	return nil
}

// GetRankingDate retrieves the current ranking date record
func (f *TestFixtures) GetRankingDate(ctx context.Context) (*struct {
	ID              int32
	DailyStartedAt  time.Time
	DailyEndedAt    time.Time
	WeeklyStartedAt time.Time
	WeeklyEndedAt   time.Time
}, error) {
	var rankingDate struct {
		ID              int32
		DailyStartedAt  time.Time
		DailyEndedAt    time.Time
		WeeklyStartedAt time.Time
		WeeklyEndedAt   time.Time
	}
	query := `SELECT id, daily_started_at, daily_ended_at, weekly_started_at, weekly_ended_at
	          FROM ranking_dates LIMIT 1`
	err := f.Pool.QueryRow(ctx, query).Scan(
		&rankingDate.ID, &rankingDate.DailyStartedAt, &rankingDate.DailyEndedAt,
		&rankingDate.WeeklyStartedAt, &rankingDate.WeeklyEndedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get ranking_date: %w", err)
	}
	return &rankingDate, nil
}

// ===== Daily Rewards Fixtures =====

// CreateDailyReward creates a daily_reward record for a user
func (f *TestFixtures) CreateDailyReward(ctx context.Context, userID int64, currentDay int32, startedAt, endedAt time.Time) error {
	query := `INSERT INTO daily_rewards (user_id, current_day, current_card_index, cards_start_index, started_at, ended_at)
	          VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := f.Pool.Exec(ctx, query, userID, currentDay, currentDay, 0, startedAt, endedAt)
	if err != nil {
		return fmt.Errorf("failed to create daily_reward: %w", err)
	}
	return nil
}

// GetDailyReward retrieves a user's daily reward record
func (f *TestFixtures) GetDailyReward(ctx context.Context, userID int64) (*struct {
	UserID           int64
	CurrentDay       int32
	CurrentCardIndex int32
	CardsStartIndex  int32
	StartedAt        time.Time
	EndedAt          time.Time
}, error) {
	var dr struct {
		UserID           int64
		CurrentDay       int32
		CurrentCardIndex int32
		CardsStartIndex  int32
		StartedAt        time.Time
		EndedAt          time.Time
	}
	query := `SELECT user_id, current_day, current_card_index, cards_start_index, started_at, ended_at
	          FROM daily_rewards WHERE user_id = $1`
	err := f.Pool.QueryRow(ctx, query, userID).Scan(
		&dr.UserID, &dr.CurrentDay, &dr.CurrentCardIndex, &dr.CardsStartIndex, &dr.StartedAt, &dr.EndedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily_reward: %w", err)
	}
	return &dr, nil
}

// CreateDRCardClaim creates a daily reward card claim for a user
func (f *TestFixtures) CreateDRCardClaim(ctx context.Context, userID int64, cardID int32) error {
	query := `INSERT INTO user_dr_card_claims (user_id, card_id) VALUES ($1, $2)`
	_, err := f.Pool.Exec(ctx, query, userID, cardID)
	if err != nil {
		return fmt.Errorf("failed to create dr_card_claim: %w", err)
	}
	return nil
}

// CreateDRBoxClaim creates a daily reward box claim for a user
func (f *TestFixtures) CreateDRBoxClaim(ctx context.Context, userID int64, boxID int32) error {
	query := `INSERT INTO user_dr_box_claims (user_id, box_id) VALUES ($1, $2)`
	_, err := f.Pool.Exec(ctx, query, userID, boxID)
	if err != nil {
		return fmt.Errorf("failed to create dr_box_claim: %w", err)
	}
	return nil
}

// GetDRCardClaimCount retrieves the count of card claims for a user
func (f *TestFixtures) GetDRCardClaimCount(ctx context.Context, userID int64) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM user_dr_card_claims WHERE user_id = $1`
	err := f.Pool.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count dr_card_claims: %w", err)
	}
	return count, nil
}

// GetDRBoxClaimCount retrieves the count of box claims for a user
func (f *TestFixtures) GetDRBoxClaimCount(ctx context.Context, userID int64) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM user_dr_box_claims WHERE user_id = $1`
	err := f.Pool.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count dr_box_claims: %w", err)
	}
	return count, nil
}

// ===== Event Package Fixtures =====

// CreateEPDuration creates an event package duration timer for a user
func (f *TestFixtures) CreateEPDuration(ctx context.Context, userID int64, eventPackageID int32, startedAt, expiredAt time.Time) error {
	query := `INSERT INTO ep_durations (user_id, event_package_id, started_at, expired_at)
	          VALUES ($1, $2, $3, $4)
	          ON CONFLICT (user_id, event_package_id) DO UPDATE
	          SET started_at = $3, expired_at = $4`
	_, err := f.Pool.Exec(ctx, query, userID, eventPackageID, startedAt, expiredAt)
	if err != nil {
		return fmt.Errorf("failed to create ep_duration: %w", err)
	}
	return nil
}

// GetEPDuration retrieves an event package duration timer
func (f *TestFixtures) GetEPDuration(ctx context.Context, userID int64, eventPackageID int32) (*struct {
	UserID         int64
	EventPackageID int32
	StartedAt      time.Time
	ExpiredAt      time.Time
}, error) {
	var epd struct {
		UserID         int64
		EventPackageID int32
		StartedAt      time.Time
		ExpiredAt      time.Time
	}
	query := `SELECT user_id, event_package_id, started_at, expired_at
	          FROM ep_durations WHERE user_id = $1 AND event_package_id = $2`
	err := f.Pool.QueryRow(ctx, query, userID, eventPackageID).Scan(
		&epd.UserID, &epd.EventPackageID, &epd.StartedAt, &epd.ExpiredAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get ep_duration: %w", err)
	}
	return &epd, nil
}

// CreateEPTransaction creates an event package purchase transaction
func (f *TestFixtures) CreateEPTransaction(ctx context.Context, userID int64, eventPackageID int32) (int32, error) {
	var transactionID int32
	query := `INSERT INTO ep_transactions (user_id, event_package_id, purchased_at)
	          VALUES ($1, $2, NOW()) RETURNING id`
	err := f.Pool.QueryRow(ctx, query, userID, eventPackageID).Scan(&transactionID)
	if err != nil {
		return 0, fmt.Errorf("failed to create ep_transaction: %w", err)
	}
	return transactionID, nil
}

// GetEPTransactionCount retrieves the count of purchases for a user and package
func (f *TestFixtures) GetEPTransactionCount(ctx context.Context, userID int64, eventPackageID int32) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM ep_transactions WHERE user_id = $1 AND event_package_id = $2`
	err := f.Pool.QueryRow(ctx, query, userID, eventPackageID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count ep_transactions: %w", err)
	}
	return count, nil
}

// GetAllEPTransactions retrieves all purchase transactions for a user
func (f *TestFixtures) GetAllEPTransactions(ctx context.Context, userID int64) ([]struct {
	ID             int32
	UserID         int64
	EventPackageID int32
	PurchasedAt    time.Time
}, error) {
	query := `SELECT id, user_id, event_package_id, purchased_at
	          FROM ep_transactions WHERE user_id = $1 ORDER BY purchased_at DESC`
	rows, err := f.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ep_transactions: %w", err)
	}
	defer rows.Close()

	var transactions []struct {
		ID             int32
		UserID         int64
		EventPackageID int32
		PurchasedAt    time.Time
	}

	for rows.Next() {
		var tx struct {
			ID             int32
			UserID         int64
			EventPackageID int32
			PurchasedAt    time.Time
		}
		if err := rows.Scan(&tx.ID, &tx.UserID, &tx.EventPackageID, &tx.PurchasedAt); err != nil {
			return nil, fmt.Errorf("failed to scan ep_transaction: %w", err)
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// CreateUserQuestionStat creates a user question stat record for testing
func (f *TestFixtures) CreateUserQuestionStat(
	ctx context.Context,
	userID int64,
	categoryName string,
	answeredCount, totalCount, correctCount int32,
) error {
	query := `
		INSERT INTO user_question_stats (user_id, category_name, answered_question_count, total_question_count, correct_question_count)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, category_name) DO UPDATE
		SET answered_question_count = EXCLUDED.answered_question_count,
		    total_question_count = EXCLUDED.total_question_count,
		    correct_question_count = EXCLUDED.correct_question_count
	`
	_, err := f.Pool.Exec(ctx, query, userID, categoryName, answeredCount, totalCount, correctCount)
	if err != nil {
		return fmt.Errorf("failed to create user question stat: %w", err)
	}
	return nil
}

// GetUserQuestionStats retrieves all user question stats for a user
func (f *TestFixtures) GetUserQuestionStats(ctx context.Context, userID int64) ([]sqlc.UserQuestionStat, error) {
	query := `
		SELECT id, user_id, category_name, answered_question_count, total_question_count, correct_question_count, created_at, updated_at
		FROM user_question_stats
		WHERE user_id = $1
		ORDER BY category_name
	`
	rows, err := f.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user question stats: %w", err)
	}
	defer rows.Close()

	var stats []sqlc.UserQuestionStat
	for rows.Next() {
		var stat sqlc.UserQuestionStat
		if err := rows.Scan(
			&stat.ID,
			&stat.UserID,
			&stat.CategoryName,
			&stat.AnsweredQuestionCount,
			&stat.TotalQuestionCount,
			&stat.CorrectQuestionCount,
			&stat.CreatedAt,
			&stat.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user question stat: %w", err)
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

// SetupClassicModeGameData sets up all necessary data for a classic mode game test.
// This includes: user, config, game type, category, question, user progress, and stage.
// Returns the created user, question ID, and category ID.
type ClassicModeGameData struct {
	User       *sqlc.User
	QuestionID int32
	CategoryID int32
}

func (f *TestFixtures) SetupClassicModeGameData(ctx context.Context, categoryName string, questionText string) (*ClassicModeGameData, error) {
	// Create a test user
	user, err := f.CreateTestUser(ctx, uuid.New().String())
	if err != nil {
		return nil, fmt.Errorf("failed to create test user: %w", err)
	}

	// Create game config (may already exist from SetupTest, so ignore conflicts)
	if err := f.CreateTestConfig(ctx, 10, 5); err != nil {
		return nil, fmt.Errorf("failed to create test config: %w", err)
	}

	// Create a game type (0 = Classic Mode)
	if err := f.CreateTestGameType(ctx, "Classic Mode", 0); err != nil {
		return nil, fmt.Errorf("failed to create game type: %w", err)
	}

	// Create a category with display_order=0
	categoryID, err := f.CreateTestCategory(ctx, categoryName, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to create test category: %w", err)
	}

	// Create a question (game_type: 0 = Classic Mode)
	questionID, err := f.CreateTestQuestion(ctx, 0, categoryID, questionText)
	if err != nil {
		return nil, fmt.Errorf("failed to create test question: %w", err)
	}

	// Create user_progress entry (required for GetNextClassicModeQuestion to work)
	// Note: We use raw query here because CreateNewUserProgress is :batchexec in sqlc
	_, err = f.Pool.Exec(ctx,
		"INSERT INTO user_progresses (user_id, question_id, status, updated_at) VALUES ($1, $2, 'unsolved', now())",
		user.ID, questionID)
	if err != nil {
		return nil, fmt.Errorf("failed to create user progress: %w", err)
	}

	// Create stage entry (required for SelectAnswer to update progress)
	if err := f.CreateTestStage(ctx, user.ID, categoryID, 1, 10, 5, false); err != nil {
		return nil, fmt.Errorf("failed to create stage: %w", err)
	}

	return &ClassicModeGameData{
		User:       user,
		QuestionID: questionID,
		CategoryID: categoryID,
	}, nil
}

// SetupClassicModeGameDataWithUser sets up all necessary data for a classic mode game test with an existing user.
// This is useful when you need to set specific user properties (like item quantities) before starting the game.
// Returns the created question ID and category ID.
func (f *TestFixtures) SetupClassicModeGameDataWithUser(ctx context.Context, user *sqlc.User, categoryName string, questionText string) (*ClassicModeGameData, error) {
	// Create game config (may already exist from SetupTest, so ignore conflicts)
	if err := f.CreateTestConfig(ctx, 10, 5); err != nil {
		return nil, fmt.Errorf("failed to create test config: %w", err)
	}

	// Create a game type (0 = Classic Mode)
	if err := f.CreateTestGameType(ctx, "Classic Mode", 0); err != nil {
		return nil, fmt.Errorf("failed to create game type: %w", err)
	}

	// Create a category with display_order=0
	categoryID, err := f.CreateTestCategory(ctx, categoryName, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to create test category: %w", err)
	}

	// Create a question (game_type: 0 = Classic Mode)
	questionID, err := f.CreateTestQuestion(ctx, 0, categoryID, questionText)
	if err != nil {
		return nil, fmt.Errorf("failed to create test question: %w", err)
	}

	// Create user_progress entry (required for GetNextClassicModeQuestion to work)
	_, err = f.Pool.Exec(ctx,
		"INSERT INTO user_progresses (user_id, question_id, status, updated_at) VALUES ($1, $2, 'unsolved', now())",
		user.ID, questionID)
	if err != nil {
		return nil, fmt.Errorf("failed to create user progress: %w", err)
	}

	// Create stage entry (required for SelectAnswer to update progress)
	if err := f.CreateTestStage(ctx, user.ID, categoryID, 1, 10, 5, false); err != nil {
		return nil, fmt.Errorf("failed to create stage: %w", err)
	}

	return &ClassicModeGameData{
		User:       user,
		QuestionID: questionID,
		CategoryID: categoryID,
	}, nil
}

// GetStageByUserAndCategory retrieves a stage by user ID and category ID
func (f *TestFixtures) GetStageByUserAndCategory(ctx context.Context, userID int64, categoryID int32) (*sqlc.Stage, error) {
	var stage sqlc.Stage
	query := `
		SELECT id, user_id, category_id, game_type_id, stage_progress,
		       coin_rewards, crown_rewards, is_claimed, is_quit, created_at, updated_at, quited_at
		FROM stages
		WHERE user_id = $1 AND category_id = $2
	`
	err := f.Pool.QueryRow(ctx, query, userID, categoryID).Scan(
		&stage.ID,
		&stage.UserID,
		&stage.CategoryID,
		&stage.GameTypeID,
		&stage.StageProgress,
		&stage.CoinRewards,
		&stage.CrownRewards,
		&stage.IsClaimed,
		&stage.IsQuit,
		&stage.CreatedAt,
		&stage.UpdatedAt,
		&stage.QuitedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get stage: %w", err)
	}
	return &stage, nil
}

// ChallengeModeGameData holds test data for Challenge Mode games
type ChallengeModeGameData struct {
	User       *sqlc.User
	QuestionID int32
	CategoryID int32
	GameTypeID int32
}

// SetupChallengeModeGameData sets up all necessary data for a challenge mode game test.
// This includes: user, config, game type, category, question, user progress, and challenge stage.
// Returns the created user, question ID, category ID, and game type ID.
func (f *TestFixtures) SetupChallengeModeGameData(ctx context.Context, gameTypeID int32, categoryName string, questionText string) (*ChallengeModeGameData, error) {
	// Create a test user
	user, err := f.CreateTestUser(ctx, uuid.New().String())
	if err != nil {
		return nil, fmt.Errorf("failed to create test user: %w", err)
	}

	// Create game config (may already exist from SetupTest, so ignore conflicts)
	if err := f.CreateTestConfig(ctx, 10, 5); err != nil {
		return nil, fmt.Errorf("failed to create test config: %w", err)
	}

	// Create game type name based on ID
	gameTypeName := ""
	switch gameTypeID {
	case 1:
		gameTypeName = "True or False"
	case 2:
		gameTypeName = "Would You Rather"
	case 3:
		gameTypeName = "Which Came First"
	case 4:
		gameTypeName = "Odd One Out"
	case 5:
		gameTypeName = "Idiot Test"
	default:
		return nil, fmt.Errorf("unsupported game type ID: %d", gameTypeID)
	}

	// Create game type
	if err := f.CreateTestGameType(ctx, gameTypeName, gameTypeID); err != nil {
		return nil, fmt.Errorf("failed to create game type: %w", err)
	}

	// Create a category with display_order=0
	categoryID, err := f.CreateTestCategory(ctx, categoryName, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to create test category: %w", err)
	}

	// Create a question for the specified game type
	questionID, err := f.CreateTestQuestion(ctx, gameTypeID, categoryID, questionText)
	if err != nil {
		return nil, fmt.Errorf("failed to create test question: %w", err)
	}

	// Create user_progress entry (required for GetNextQuestion to work)
	_, err = f.Pool.Exec(ctx,
		"INSERT INTO user_progresses (user_id, question_id, status, updated_at) VALUES ($1, $2, 'unsolved', now())",
		user.ID, questionID)
	if err != nil {
		return nil, fmt.Errorf("failed to create user progress: %w", err)
	}

	// Create challenge stage entry (required for SelectAnswerChallenge to update progress)
	// Using CreateTestChallengeStage with initial stage_progress=1
	if err := f.CreateTestChallengeStage(ctx, user.ID, gameTypeID, 1, false); err != nil {
		return nil, fmt.Errorf("failed to create challenge stage: %w", err)
	}

	return &ChallengeModeGameData{
		User:       user,
		QuestionID: questionID,
		CategoryID: categoryID,
		GameTypeID: gameTypeID,
	}, nil
}

// ============================================================================
// User Profile Fixture Methods (for boilerplate schema)
// ============================================================================

// User represents a user in the boilerplate schema
type User struct {
	ID          int64
	PublicID    pgtype.UUID
	Email       string
	Username    string
	DisplayName string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   pgtype.Timestamptz
}

// CreateUser creates a user with flexible parameters for the boilerplate schema
// Accepts a map with optional fields: public_id, email, username, display_name
func (f *TestFixtures) CreateUser(ctx context.Context, params map[string]interface{}) (*User, error) {
	// Set defaults
	publicID := pgtype.UUID{}
	if pid, ok := params["public_id"].(pgtype.UUID); ok {
		publicID = pid
	} else {
		// Generate random UUID if not provided
		uid := uuid.New()
		if err := publicID.Scan(uid.String()); err != nil {
			return nil, fmt.Errorf("failed to scan UUID: %w", err)
		}
	}

	email := "test@example.com"
	if e, ok := params["email"].(string); ok {
		email = e
	}

	username := "testuser"
	if u, ok := params["username"].(string); ok {
		username = u
	}

	displayName := "Test User"
	if d, ok := params["display_name"].(string); ok {
		displayName = d
	}

	query := `
		INSERT INTO users (public_id, email, username, display_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id, public_id, email, username, display_name, created_at, updated_at, deleted_at
	`

	var user User
	err := f.Pool.QueryRow(ctx, query, publicID, email, username, displayName).Scan(
		&user.ID,
		&user.PublicID,
		&user.Email,
		&user.Username,
		&user.DisplayName,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// GetUserByPublicID retrieves a user by their public ID
func (f *TestFixtures) GetUserByPublicID(ctx context.Context, publicID pgtype.UUID) (*User, error) {
	query := `
		SELECT id, public_id, email, username, display_name, created_at, updated_at, deleted_at
		FROM users
		WHERE public_id = $1
	`

	var user User
	err := f.Pool.QueryRow(ctx, query, publicID).Scan(
		&user.ID,
		&user.PublicID,
		&user.Email,
		&user.Username,
		&user.DisplayName,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get user by public ID: %w", err)
	}

	return &user, nil
}

// SoftDeleteUser soft deletes a user by setting deleted_at timestamp
func (f *TestFixtures) SoftDeleteUser(ctx context.Context, publicID pgtype.UUID) error {
	query := `
		UPDATE users
		SET deleted_at = NOW()
		WHERE public_id = $1
	`

	_, err := f.Pool.Exec(ctx, query, publicID)
	if err != nil {
		return fmt.Errorf("failed to soft delete user: %w", err)
	}

	return nil
}
