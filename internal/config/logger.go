package config

import (
	"log/slog"
	"os"
)

func LogInit() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				t := a.Value.Time()
				return slog.Attr{
					Key:   slog.TimeKey,
					Value: slog.StringValue(t.Format("2006-01-02 15:04:05")),
				}
			}
			return a
		},
	}))

	slog.SetDefault(logger)
}
