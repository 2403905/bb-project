package main

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLogger set the global configuration
// The logLevel definition
//	-1 TraceLevel defines trace log level.
//	0 DebugLevel defines debug log level.
//	1 InfoLevel defines info log level.
//	2 WarnLevel defines warn log level.
//	3 ErrorLevel defines error log level.
//	4 FatalLevel defines fatal log level.
//	5 PanicLevel defines panic log level.
//	6 NoLevel defines an absent log level.
//	7 Disabled disables the logger.
func InitLogger(logLevel int, pretty bool) {
	// Init logger
	zerolog.SetGlobalLevel(zerolog.Level(logLevel))
	zerolog.TimeFieldFormat = time.RFC3339
	if pretty {
		log.Logger = log.With().Caller().Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	} else {
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()
	}
}
