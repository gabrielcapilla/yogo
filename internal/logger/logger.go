package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var Log *log.Logger

func init() {
	logPath := "/tmp/yogo.log"
	configDir, err := os.UserConfigDir()
	if err == nil {
		yogoDir := filepath.Join(configDir, "yogo")
		if err := os.MkdirAll(yogoDir, 0755); err == nil {
			logPath = filepath.Join(yogoDir, "yogo.log")
		}
	}

	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(fmt.Sprintf("Could not open log file: %v", err))
	}

	Log = log.New(file, "YOGO: ", log.LstdFlags|log.Lshortfile)
	Log.Println("Logger initialized. Writing to:", logPath)
}
