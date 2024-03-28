package devpod

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
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

func TestGitCredential(t *testing.T) {

	// Check git must be installed
	gitPath, err := exec.LookPath("git")
	if err != nil {
		t.Errorf("Expected git to be installed, but got: %v", err)
	}

	if gitPath == "" {
		t.Errorf("Expected git to be installed, but got: %v", gitPath)
	}

	// Create a test GitCredential
	workspaceID := "test_workspace"
	username := "test_username"
	token := "test_token"

	gitCred := NewGitCredential(workspaceID, username, token)
	// Test Init method
	gitCred.Init()

	// check env ar set
	if gitCred.envUser != "GIT_USER_"+workspaceID {
		t.Errorf("Expected envUser to be %s, got %s", "GIT_USER_"+workspaceID, gitCred.envUser)
	}
	if gitCred.envToken != "GIT_TOKEN_"+workspaceID {
		t.Errorf("Expected envToken to be %s, got %s", "GIT_TOKEN_"+workspaceID, gitCred.envToken)
	}

	// check if scriptPath and scriptContent are set
	if !strings.Contains(gitCred.scriptPath, workspaceID) {
		t.Errorf("Expected scriptPath to not contain %s, got %s", workspaceID, gitCred.scriptPath)
	}

	if !strings.Contains(gitCred.scriptContent, workspaceID) {
		t.Errorf("Expected scriptContent to not contain %s, got %s", workspaceID, gitCred.scriptContent)
	}

	// Test Set method
	if err := gitCred.Set(); err != nil {
		t.Errorf("Set method returned an error: %v", err)
	}

	// Test if the script file is created
	if _, err := os.Stat(gitCred.scriptPath); os.IsNotExist(err) {
		t.Errorf("Expected script file to exist, but it does not")
	}

	// check env ar set
	if os.Getenv(gitCred.envUser) != username {
		t.Errorf("Expected envUser to be %s, got %s", username, os.Getenv(gitCred.envUser))
	}

	if os.Getenv(gitCred.envToken) != token {
		t.Errorf("Expected envToken to be %s, got %s", token, os.Getenv(gitCred.envToken))
	}

	// Test if the script file contains the expected content
	scriptContent, err := ioutil.ReadFile(gitCred.scriptPath)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	if string(scriptContent) != gitCred.scriptContent {
		t.Errorf("Expected script content to be %s, got %s", gitCred.scriptContent, string(scriptContent))
	}

	// check git credential is set
	contentS, err := exec.Command("git", "config", "--get", "--global", "credential.helper").Output()
	if err != nil {
		t.Errorf("Expected git credential is set, but got: %v", err)
	}

	if !strings.Contains(string(contentS), gitCred.scriptPath) {
		t.Errorf("Expected git credential is set, but got: %s", string(contentS))
	}

	// Test Unset method
	err = gitCred.Remove()
	if err != nil {
		t.Errorf("Unset method returned an error: %v", err)
	}

	// Test if the script file is removed
	if _, err := os.Stat(gitCred.scriptPath); !os.IsNotExist(err) {
		t.Errorf("Expected script file to not exist, but it does")
	}

	// check env ar set
	if os.Getenv(gitCred.envUser) != "" {
		t.Errorf("Expected envUser to be empty, got %s", os.Getenv(gitCred.envUser))
	}

	if os.Getenv(gitCred.envToken) != "" {
		t.Errorf("Expected envToken to be empty, got %s", os.Getenv(gitCred.envToken))
	}

	// check git credential is unset
	contentS, err = exec.Command("git", "config", "--get", "--global", "credential.helper").Output()
	if err == nil {
		t.Errorf("Expected git credential is unset, but got: %s", string(contentS))
	}

	if string(contentS) != "" {
		t.Errorf("Expected git credential is unset, but got: %s", string(contentS))
	}

}
