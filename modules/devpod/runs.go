package devpod

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const RunsFolder = "devpod-runs"

// CreateRunFile initializes the run log file with the start time and status
func CreateRunFile(filename string) error {
	// Ensure the runs directory exists
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return err
	}

	// Create the run file with initial content
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	startTime := time.Now().Format(time.RFC3339)
	initialContent := fmt.Sprintf("Start Time: %s\nStatus: running\nLog:\n", startTime)
	_, err = file.WriteString(initialContent)
	return err
}

// DeleteRunFile removes the specified run log file if it exists
func DeleteRunFile(filename string) error {
	// Check if the file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", filename)
	}

	// Delete the file
	err := os.Remove(filename)
	if err != nil {
		return fmt.Errorf("failed to delete file %s: %v", filename, err)
	}

	fmt.Printf("File %s deleted successfully\n", filename)
	return nil
}

type logFileEntry struct {
	file       *os.File
	writer     *bufio.Writer
	closeTimer *time.Timer
}

var (
	logFiles       = make(map[string]*logFileEntry)
	logMutex       sync.Mutex
	inactivityTime = 10 * time.Second
)

// LogRunMessageToFile appends a log entry to the run file
func LogRunMessageToFile(filename, message string) error {
	logMutex.Lock()
	defer logMutex.Unlock()

	if _, exists := logFiles[filename]; !exists {
		file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		logFiles[filename] = &logFileEntry{
			file:      file,
			writer:    bufio.NewWriter(file),
			closeTimer: nil,
		}
	}

	entry := logFiles[filename]

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	line := fmt.Sprintf("[%s] %s\n", timestamp, message)
	if _, err := entry.writer.WriteString(line); err != nil {
		return fmt.Errorf("failed to write to log file: %w", err)
	}

	if err := entry.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush buffer: %w", err)
	}

	if entry.closeTimer != nil {
		entry.closeTimer.Stop()
	}
	entry.closeTimer = time.AfterFunc(inactivityTime, func() {
		closeLogFile(filename)
	})

	return nil
}

func closeLogFile(filename string) {
	logMutex.Lock()
	defer logMutex.Unlock()

	entry, exists := logFiles[filename]
	if !exists {
		return
	}

	entry.writer.Flush()
	entry.file.Close()

	delete(logFiles, filename)
}

type LogFile struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

// ReadLogFiles reads all files in the given directory and returns their content as a JSON array
func ReadLogFiles(dir string) ([]LogFile, error) {
	var logFiles []LogFile

	// Walk through the directory to find all files
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Read the file content
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		// Add file info and content to the list
		logFiles = append(logFiles, LogFile{
			Filename: filepath.Base(path),
			Content:  string(content),
		})

		return nil
	})
	if err != nil {
		return nil, err
	}
	return logFiles, nil
}
