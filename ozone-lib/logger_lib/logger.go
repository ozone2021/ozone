package logger_lib

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/ozone2021/ozone/ozone-lib/utils"
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

func New(runId, ozoneWorkingDirectory, rootRunnable string, headless bool) (*Logger, error) {
	if headless {
		return &Logger{
			Logger:   log.New(os.Stdout, "", log.Ldate|log.Ltime),
			FileName: "stdout",
			File:     os.Stdout,
			closer:   nil,
		}, nil
	}

	filePath, workLogDir := getLogFilePathAndDir(ozoneWorkingDirectory, runId, rootRunnable)

	_, doesntExistErr := os.Stat(workLogDir)
	if os.IsNotExist(doesntExistErr) {
		err := os.MkdirAll(workLogDir, 0755)
		if err != nil {
			log.Fatalln(err)
		}
	}

	//os.Remove(filePath)

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}

	logger := &Logger{
		Logger:   log.New(file, "", 0),
		FileName: filePath,
		File:     file,
		closer: func() error {
			return file.Close()
		},
	}

	return logger, nil
}

func (l *Logger) GetLogFilePath() string {
	return filepath.Join(l.WorkLogDir, l.FileName)
}

func getLogFilePathAndDir(ozoneWorkingDirectory, runId, rootRunnable string) (string, string) {
	workLogDir := getWorkLogDir(ozoneWorkingDirectory, runId)

	fileName := rootRunnable + ".log"

	filePath := filepath.Join(workLogDir, fileName)

	return filePath, workLogDir
}

func getWorkLogDir(ozoneWorkingDirectory, runId string) string {
	return filepath.Join(utils.GetTmpDir(ozoneWorkingDirectory), runId, "logs")
}

func ClearLogs(ozoneWorkingDirectory, runId string) error {
	workLogDir := getWorkLogDir(ozoneWorkingDirectory, runId)

	err := filepath.Walk(workLogDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			err := os.Remove(path)
			if err != nil {
				return errors.New(fmt.Sprintf("Error deleting file:", path, err))
			}
		}
		return nil
	})

	if err != nil {
		return errors.New(fmt.Sprintf("Error walking through directory:", err))
	}
	return nil
}

func (l *Logger) Close() error {
	if l.closer != nil {
		return l.closer()
	}
	return nil
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
