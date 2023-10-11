package logger

import (
	"encoding/base64"
	"log"
	"os"
	"path/filepath"
)

type Logger struct {
	LogDirectory string
	Filename     string
}

func New(logDirectory, ozoneWorkingDirectory, rootRunnable string) *Logger {
	ozoneWorkDir64 := base64.StdEncoding.EncodeToString([]byte(ozoneWorkingDirectory))

	return &Logger{
		LogDirectory: filepath.Join(logDirectory, ozoneWorkDir64),
		Filename:     rootRunnable + ".log",
	}
}

func (l *Logger) SetLogFile() {
	f, err := os.OpenFile(filepath.Join(l.LogDirectory, l.Filename), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error setting log file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
}
