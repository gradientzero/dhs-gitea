package devpod

import (
	"bytes"
	"io/ioutil"
	"os/exec"
	"testing"
	"time"
)

func TestGetOutputCommand(t *testing.T) {
	// Test case 1: Successful execution with output
	t.Run("Successful execution with output", func(t *testing.T) {
		// Create a fake command
		cmd := exec.Command("echo", "Hello, world!")

		// Set up a buffer to capture the output
		var outputBuffer bytes.Buffer
		sendStream := func(s string) {
			outputBuffer.WriteString(s)
			outputBuffer.WriteString("\n")
		}

		err := getOutputCommand(cmd, sendStream)
		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}

		expectedOutput := "Hello, world!\n"
		actualOutput := outputBuffer.String()
		if actualOutput != expectedOutput {
			t.Errorf("Expected output: %s, but got: %s", expectedOutput, actualOutput)
		}
	})

	// Test case 2: Successful execution with no output
	t.Run("Successful execution with no output", func(t *testing.T) {
		// Create a fake command
		cmd := exec.Command("true")

		// Set up a buffer to capture the output
		var outputBuffer bytes.Buffer
		sendStream := func(s string) {
			outputBuffer.WriteString(s)
			outputBuffer.WriteString("\n")
		}

		err := getOutputCommand(cmd, sendStream)
		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}

		expectedOutput := ""
		actualOutput := outputBuffer.String()
		if actualOutput != expectedOutput {
			t.Errorf("Expected output: %s, but got: %s", expectedOutput, actualOutput)
		}
	})

	// Test case 3: Execution with error
	t.Run("Execution with error", func(t *testing.T) {
		// Create a fake command
		cmd := exec.Command("nonexistent-command")

		// Set up a buffer to capture the output
		var outputBuffer bytes.Buffer
		sendStream := func(s string) {
			outputBuffer.WriteString(s)
			outputBuffer.WriteString("\n")
		}

		err := getOutputCommand(cmd, sendStream)
		if err == nil {
			t.Error("Expected error, but got nil")
		}
	})

	// Test case 4: With file script as input
	t.Run("With file script as input", func(t *testing.T) {

		// Create file script
		tmpFile, err := ioutil.TempFile("", "testing")
		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}
		defer tmpFile.Close()

		// Write script to
		content := `#!/bin/bash
echo "Executing script 1"
sleep 1
echo "Executing script 2"
sleep 1
echo "Success"
`
		if _, err := tmpFile.Write([]byte(content)); err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}

		executdAt := []time.Time{}
		// execute the script
		cmd := exec.Command("bash", tmpFile.Name())

		// Set up a buffer to capture the output
		var outputBuffer bytes.Buffer
		sendStream := func(s string) {
			executdAt = append(executdAt, time.Now())
			outputBuffer.WriteString(s)
			outputBuffer.WriteString("\n")
		}

		err = getOutputCommand(cmd, sendStream)
		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}

		// check execution time with delay
		for i := 0; i < len(executdAt)-1; i++ {
			// check execution time with delay 1 second
			if executdAt[i+1].Sub(executdAt[i]).Seconds() < 1 {
				t.Errorf("Expected execution time with delay 1 second, but got: %v", executdAt[i+1].Sub(executdAt[i]).Seconds())
			}
		}

		expectedOutput := "Executing script 1\nExecuting script 2\nSuccess\n"
		actualOutput := outputBuffer.String()
		if actualOutput != expectedOutput {
			t.Errorf("Expected output: %s, but got: %s", expectedOutput, actualOutput)
		}

	})
}
