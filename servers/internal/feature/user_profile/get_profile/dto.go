package get_profile

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type GetProfileRequest struct {
	PublicID pgtype.UUID `json:"public_id"`
}

type GetProfileResponse struct {
	PublicID    string    `json:"public_id"`
	Email       string    `json:"email"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
