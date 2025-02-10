package devpod

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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

// LogRunMessageToFile appends a log entry to the run file
func LogRunMessageToFile(filename, message string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	_, err = file.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, message))
	return err
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
