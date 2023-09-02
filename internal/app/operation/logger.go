package operation

import (
	"os"

	"github.com/kinneko-de/restaurant-document-generate-svc/build"
	"github.com/rs/zerolog"
)

var Logger zerolog.Logger

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	Logger = zerolog.New(os.Stdout).With().
		Timestamp().
		Caller().
		Str("version", build.Version).
		Logger()
}
