package get_profile

import (
	"errors"
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

	// Parse request
	var req GetProfileRequest
	if err := httputil.GetReqBodyWithLog(r, &req); err != nil {
		httputil.ErrWithMsgRaw(w, r, err, "failed to get request body")
		return
	}

	c := httputil.NewHttpUtilContext(w, r)

	logger.Info("getting user profile", "public_id", req.PublicID)

	// Query user by public ID
	// In a real implementation, you would use SQLC generated queries
	query := `
		SELECT public_id, email, username, display_name, created_at, updated_at
		FROM users
		WHERE public_id = $1 AND deleted_at IS NULL
	`

	var response GetProfileResponse
	err = dbconn.QueryRow(c.Ctx(), query, req.PublicID).Scan(
		&response.PublicID,
		&response.Email,
		&response.Username,
		&response.DisplayName,
		&response.CreatedAt,
		&response.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.ErrWithMsg(c, err, "user not found")
			return
		}
		httputil.ErrWithMsg(c, err, "failed to get user profile")
		return
	}

	logger.Info("user profile retrieved successfully", "username", response.Username)

	httputil.OkWithMsg(c,
		"user profile retrieved successfully",
		response)
}
