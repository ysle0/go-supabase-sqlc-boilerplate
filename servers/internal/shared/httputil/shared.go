package httputil

import (
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
