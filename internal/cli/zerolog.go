package cli

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func SetupZerolog(levelName, formatName string) {
	if strings.EqualFold(formatName, "pretty") {
		writer := zerolog.ConsoleWriter{Out: os.Stderr}
		if noColor, ok := os.LookupEnv("NO_COLOR"); ok {
			if noColor != "" {
				writer.NoColor = true
			}
		}
		log.Logger = log.Output(writer)
	}
	var level zerolog.Level
	switch strings.ToLower(levelName) {
	case "error":
		level = zerolog.ErrorLevel
	case "warn":
		level = zerolog.WarnLevel
	case "info":
		level = zerolog.InfoLevel
	case "debug":
		level = zerolog.DebugLevel
	case "trace":
		level = zerolog.TraceLevel
	default:
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	log.Info().Stringer("level", level).Msg("log level set")
}
