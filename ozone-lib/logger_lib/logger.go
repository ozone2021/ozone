package logger_lib

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type Logger struct {
	*log.Logger
	File       *os.File
	FileName   string
	WorkLogDir string
	closer     func() error
}

func New(ozoneWorkingDirectory, rootRunnable string, headless bool) (*Logger, error) {
	if headless {
		log.New(os.Stdout, "", log.Ldate|log.Ltime)
	}

	filePath, workLogDir := getLogFilePathAndDir(ozoneWorkingDirectory, rootRunnable)

	_, err := os.Stat(workLogDir)
	if os.IsNotExist(err) {
		err := os.Mkdir(workLogDir, 0755)
		if err != nil {
			log.Fatalln(err)
		}
	}

	var file *os.File
	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		file, err = os.Create(filePath)
		if err != nil && err != os.ErrExist {
			log.Fatalln(fmt.Sprintf("Error creating logger: %s", err))
		}
	} else {
		file, err = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal(err)
		}
	}

	return &Logger{
		Logger:   log.New(file, "", 0),
		FileName: filePath,
		File:     file,
		closer: func() error {
			return file.Close()
		},
	}, nil
}

func (l *Logger) GetLogFilePath() string {
	return filepath.Join(l.WorkLogDir, l.FileName)
}

func getLogFilePathAndDir(ozoneWorkingDirectory string, rootRunnable string) (string, string) {
	ozoneWorkDir64 := base64.StdEncoding.EncodeToString([]byte(ozoneWorkingDirectory))
	fileName := rootRunnable + ".log"

	workLogDir := filepath.Join("/tmp/ozone", ozoneWorkDir64)

	filePath := filepath.Join(workLogDir, fileName)

	return filePath, workLogDir
}

func (l *Logger) Close() error {
	return l.closer()
}

func (l *Logger) TailFile(n int) ([]string, error) {
	file, err := os.Open(l.GetLogFilePath())
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := make([]string, 0, n)

	// Read all lines of the file into a slice
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Calculate the starting index to extract the last n lines
	startIndex := len(lines) - n
	if startIndex < 0 {
		startIndex = 0
	}

	return lines[startIndex:], nil
}
