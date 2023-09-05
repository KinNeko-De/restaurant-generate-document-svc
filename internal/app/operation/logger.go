package operation

import (
	"os"

	"github.com/kinneko-de/restaurant-document-generate-svc/build"
	"github.com/rs/zerolog"
)

var Logger zerolog.Logger = zerolog.New(os.Stdout).With().
	Timestamp().
	Caller().
	Str("version", build.Version).
	Logger()

func SetDefaultLoggingLevel() {
	SetLoggingLevel(zerolog.InfoLevel)
}

func SetLoggingLevel(level zerolog.Level) {
	zerolog.SetGlobalLevel(level)
}
