package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
)

var Log zerolog.Logger

func Setup(debug bool) {
	if !debug {
		Log = zerolog.Nop()
		return
	}

	logPath := "/tmp"
	configDir, err := os.UserConfigDir()
	if err == nil {
		yogoDir := filepath.Join(configDir, "yogo")
		if err := os.MkdirAll(yogoDir, 0755); err == nil {
			logPath = yogoDir
		}
	}

	logFileName := fmt.Sprintf("yogo-%s.log", time.Now().Format("20060102-150405"))
	logFilePath := filepath.Join(logPath, logFileName)

	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(fmt.Sprintf("Could not open log file: %v", err))
	}

	Log = zerolog.New(file).With().Timestamp().Caller().Logger()
	Log.Info().Msgf("Logger initialized. Writing to: %s", logFilePath)
}
