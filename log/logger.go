package log

import (
	"github.com/woshikedayaa/boxtray/config"
	"log/slog"
	"os"
)

type Logger = slog.Logger

var (
	globalLogger *Logger
)

const FieldKey = "logger"

func Init(config config.LogConfig) error {
	lev := slog.Level(0)
	err := lev.UnmarshalText([]byte(config.Level))
	if err != nil {
		return err
	}

	globalLogger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: lev == slog.LevelDebug,
		Level:     lev,
	}))

	return nil
}

func Get(field string) *Logger {
	return globalLogger.With(slog.String(FieldKey, field))
}
