package compute

import (
	"log/slog"
	"os"
)

// InitLogging sets up structured key=value logging.
func InitLogging() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case "payload", "mask", "original":
				return slog.String(a.Key, "<REDACTED>")
			}
			return a
		},
	}))
}
