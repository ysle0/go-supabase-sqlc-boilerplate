package httputil

import (
	"io"
	"net/http"

	"github.com/your-org/go-monorepo-boilerplate/servers/internal/shared/jsonutil"
)

func GetReqBody[T any](r *http.Request, out *T) error {
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("failed to read request body", "error", err)
		return err
	}

	return jsonutil.Unmarshal(raw, out)
}
func GetReqBodyWithLog[T any](r *http.Request, out *T) error {
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("failed to read request body", "error", err)
		return err
	}

	return jsonutil.UnmarshalWithLog(raw, out)
}
