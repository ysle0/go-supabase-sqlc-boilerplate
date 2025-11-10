package shared

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// Closer is an interface for gracefully closing resources
type Closer interface {
	Close(ctx context.Context) error
}

type ClassicMode struct {
	Text        string   `json:"questionText"`
	Choices     []string `json:"choices"`
	AnswerIndex int      `json:"answerIndex"`
}

type TrueOrFalse struct {
	QuestionText      string `json:"question_text"`
	QuestionNumber    int    `json:"question_number"`
	AnswerExplanation string `json:"answer_explanation"`
	AnswerIndex       int    `json:"answer_index"`
}

type HasNewVersionParsedResult struct {
	ResultType string    `json:"@type"`
	UpdatedAt  time.Time `json:"result"`
}

// FromStringToUUID converts a string to pgtype.UUID with proper error handling
func FromStringToUUID(s string) (pgtype.UUID, error) {
	var uuid pgtype.UUID
	if err := uuid.Scan(s); err != nil {
		return pgtype.UUID{}, WrapError(err, "failed to convert string to pgtype.UUID")
	}
	return uuid, nil
}
