package jsonutil

import (
	"encoding/json"
	"log/slog"
	"os"
	"time"

	"github.com/MatusOllah/slogcolor"
	"github.com/fatih/color"
)

var (
	logger = slog.New(slogcolor.NewHandler(os.Stdout, &slogcolor.Options{
		Level:       slog.LevelDebug,
		TimeFormat:  time.DateTime,
		SrcFileMode: slogcolor.ShortFile,
		MsgColor:    color.New(),
	}))
)

func Marshal(v interface{}) ([]byte, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		logger.Error("failed to marshal json", "error", err)
		return nil, err
	}
	return bytes, nil
}

func MarshalJson(v interface{}) (string, error) {
	bytes, err := Marshal(v)
	if err != nil {
		return "<error>", err
	}
	return string(bytes), nil
}

func Unmarshal(data []byte, v interface{}) error {
	if err := json.Unmarshal(data, v); err != nil {
		logger.Error("failed to unmarshal json", "error", err)
		return err
	}
	return nil
}

func UnmarshalWithLog(data []byte, v interface{}) error {
	if err := json.Unmarshal(data, v); err != nil {
		logger.Error("failed to unmarshal json", "error", err)
		return err
	}

	if os.Getenv("RUN_INTEGRATION_TESTS") == "false" {
		logger.Debug("json", slog.Any("json", v))
	}
	return nil
}
