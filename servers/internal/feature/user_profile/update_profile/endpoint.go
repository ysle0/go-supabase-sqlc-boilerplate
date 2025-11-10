package update_profile

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/your-org/go-monorepo-boilerplate/servers/internal/shared/database/supabase_postgres"
	"github.com/your-org/go-monorepo-boilerplate/servers/internal/shared/httputil"
)

func Map(w http.ResponseWriter, r *http.Request) {
	logger := r.Context().Value("logger").(*slog.Logger)
	pooler := supabase_postgres.GetDBPooler()

	// Acquire database connection
	dbconn, err := pooler.Pool.Acquire(r.Context())
	if err != nil {
		httputil.ErrWithMsgRaw(w, r, err, "failed to get db pooler")
		return
	}
	defer dbconn.Release()

	// Begin transaction
	tx, err := dbconn.Begin(r.Context())
	if err != nil {
		httputil.ErrWithMsgRaw(w, r, err, "failed to begin transaction")
		return
	}

	defer func() {
		if err = tx.Rollback(r.Context()); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			logger.Error("failed rollback", "error", err)
		}
	}()

	// Parse request
	var req UpdateProfileRequest
	if err := httputil.GetReqBodyWithLog(r, &req); err != nil {
		httputil.ErrWithMsgRaw(w, r, err, "failed to get request body")
		return
	}

	c := httputil.NewHttpUtilContext(w, r)

	logger.Info("updating user profile", "public_id", req.PublicID)

	// Update user profile
	// In a real implementation, you would use SQLC generated queries
	query := `
		UPDATE users
		SET
			username = COALESCE($2, username),
			display_name = COALESCE($3, display_name),
			updated_at = NOW()
		WHERE public_id = $1 AND deleted_at IS NULL
		RETURNING public_id, email, username, display_name, updated_at
	`

	var response UpdateProfileResponse
	err = tx.QueryRow(c.Ctx(), query,
		req.PublicID,
		req.Username,
		req.DisplayName,
	).Scan(
		&response.PublicID,
		&response.Email,
		&response.Username,
		&response.DisplayName,
		&response.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.ErrWithMsg(c, err, "user not found")
			return
		}
		httputil.ErrWithMsg(c, fmt.Errorf("failed to update user profile: %w", err), "failed to update profile")
		return
	}

	// Commit transaction
	if err = tx.Commit(c.Ctx()); err != nil {
		httputil.ErrWithMsg(c, err, "failed to commit transaction")
		return
	}

	logger.Info("user profile updated successfully", "username", response.Username)

	httputil.OkWithMsg(c,
		"user profile updated successfully",
		response)
}
