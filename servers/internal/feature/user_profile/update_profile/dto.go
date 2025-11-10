package update_profile

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type UpdateProfileRequest struct {
	PublicID    pgtype.UUID `json:"public_id"`
	Username    *string     `json:"username,omitempty"`
	DisplayName *string     `json:"display_name,omitempty"`
}

type UpdateProfileResponse struct {
	PublicID    string    `json:"public_id"`
	Email       string    `json:"email"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	UpdatedAt   time.Time `json:"updated_at"`
}
