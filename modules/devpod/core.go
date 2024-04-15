package devpod

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	org_model "code.gitea.io/gitea/models/organization"
	"code.gitea.io/gitea/modules/log"
	"github.com/rs/xid"
)

// TODO: 1. need to forward S3 credentials
// TODO: 2. need to find git credential handling for pushing back
// not sure concurrent safety for devpod for now

func Execute(privateKey, user, host string, port int32,
	gitUrl, gitBranch, gitUser, gitEmail, gitToken string,
	config map[string][]org_model.OrgDevpodCredential,
	sendStream func(string),
) error {
	// gitUrl := "git@gitlab.com:grz1/aqua-research.git" // NOTE: make sure to trim, will get surprise if there is space

	providerId := xid.New().String()
	workSpaceId := xid.New().String()

	// git credential handling
	// if gitToken is available then create git credential file and set it
	if gitToken != "" {
		log.Info("Creating git creds for devpod: %v", workSpaceId)
		g := NewGitCredential(workSpaceId, gitUser, gitToken)
		err := g.Set()
		if err != nil {
			log.Error("Failing creating git creds file: %v", err)
		}
		defer g.Remove()
	}

	// setting up private key for use
	privateKeyFile := "/tmp/" + providerId + ".key"
	fmt.Println(privateKeyFile + " will created")

	// TODO: make sure append new line end of key

	// Important for line ending for correct
	privateKey = strings.Replace(privateKey, "\r\n", "\n", -1)
	// also make sure there is new line in end of file
	err := os.WriteFile(privateKeyFile, []byte(privateKey+"\n"), 0o600)
	if err != nil {
		log.Error("Failing creating private key file: %v", err)
	}

	defer func() {
		err := os.Remove(privateKeyFile)
		if err != nil {
			log.Error("Failing to remove private key file: %v", err)
		}
	}()

	// devpod provider add ssh --name <provider-id> -o HOST=vagrant@localhost -o PORT=2222 -o EXTRA_FLAGS="-i /tmp/private_key"
	cmd := exec.Command("devpod", "provider", "add", "ssh", "--name", providerId,
		"-o", "HOST="+user+"@"+host, "-o", "PORT="+strconv.Itoa(int(port)),
		"-o", "EXTRA_FLAGS=-i "+privateKeyFile) // Don't use \" in EXTRA_FLAGS for some reason it not working
	cmd.Env = os.Environ()
	cmd.Dir = "/tmp"

	err = getOutputCommand(cmd, sendStream)
	if err != nil {
		log.Error("Failing to add provider: %v", err)
	}

	// add branch to gitUrl
	gitUrl = gitUrl + "@" + gitBranch
	cmd = exec.Command("devpod", "up", workSpaceId, "--source=git:"+gitUrl, "--provider", providerId, "--ide", "none", "--id", workSpaceId)
	err = getOutputCommand(cmd, sendStream)
	if err != nil {
		log.Error("Failing to up devpod: %v", err)
	}

	for key, v := range config {
		cmd = exec.Command("devpod", "ssh", workSpaceId, "--command", "echo ['remote \""+key+"\"'] >> .dvc/config.local")
		err := getOutputCommand(cmd, sendStream)
		if err != nil {
			log.Error("Failing to add remote: %v", err)
		}
		for _, value := range v {
			cmd = exec.Command("devpod", "ssh", workSpaceId, "--command", "echo "+value.Key+"="+value.Value+" >> .dvc/config.local")
			err := getOutputCommand(cmd, sendStream)
			if err != nil {
				log.Error("Failing to add remote: %v", err)
			}
		}
	}

	cmd = exec.Command("devpod", "ssh", workSpaceId,
		"--command", "export DEVPOD_WORKSPACE_ID="+workSpaceId)
	err = getOutputCommand(cmd, sendStream)
	if err != nil {
		log.Error("Failing to add remote: %v", err)
	}

	cmd = exec.Command("devpod", "ssh", workSpaceId,
		"--command", "cat dvc-script.sh")
	var b bytes.Buffer
	cmd.Stderr = &b
	if err := cmd.Start(); err != nil {
		log.Fatal("%v", err)
	}

	// dvc-script.sh content exist
	if b.String() != "" {
		// devpod ssh <workspace-id> --command '[ -f dvc-script.sh ] && chmod +x dvc-script.sh && ./dvc-script.sh'
		cmd = exec.Command("devpod", "ssh", workSpaceId,
			"--command", "[ -f dvc-script.sh ] && chmod +x dvc-script.sh && bash dvc-script.sh")
		err = getOutputCommand(cmd, sendStream)
		if err != nil {
			log.Error("Failing to add remote: %v", err)
		}
	} else {

		// devpod ssh <workspace-id> --command 'dvc pull'
		cmd = exec.Command("devpod", "ssh", workSpaceId, "--command", "dvc pull")
		err = getOutputCommand(cmd, sendStream)
		if err != nil {
			log.Error("Failing to add remote: %v", err)
		}

		// devpod ssh <workspace-id> --command 'dvc exp run'
		cmd = exec.Command("devpod", "ssh", workSpaceId, "--command", "dvc exp run")
		err = getOutputCommand(cmd, sendStream)
		if err != nil {
			log.Error("Failing to add remote: %v", err)
		}

	}

	// devpod ssh <workspace-id> --command 'git add .'
	cmd = exec.Command("devpod", "ssh", workSpaceId, "--command", "git add .")
	err = getOutputCommand(cmd, sendStream)
	if err != nil {
		log.Error("Failing to add remote: %v", err)
	}

	// devpod ssh <workspace-id> --command 'git config user.name xxx'
	cmd = exec.Command("devpod", "ssh", workSpaceId, "--command", "git config user.name "+gitUser)
	if err := cmd.Run(); err != nil {
		log.Error("Failing to add remote: %v", err)
	}

	// devpod ssh <workspace-id> --command 'git config user.email xxx'
	cmd = exec.Command("devpod", "ssh", workSpaceId, "--command", "git config user.email "+gitEmail)
	if err = getOutputCommand(cmd, sendStream); err != nil {
		log.Error("Failing to add remote: %v", err)
	}

	// devpod ssh <workspace-id> --command 'git commit -m "exp run result"'
	cmd = exec.Command("devpod", "ssh", workSpaceId, "--command", "git commit -m 'exp run result'")
	if err = getOutputCommand(cmd, sendStream); err != nil {
		log.Error("Failing to add remote: %v", err)
	}

	// devpod ssh <workspace-id> --command 'git push origin'
	cmd = exec.Command("devpod", "ssh", workSpaceId, "--command", "git push origin")
	if err = getOutputCommand(cmd, sendStream); err != nil {
		log.Error("Failing to add remote: %v", err)
	}

	// devpod stop <workspace-id>
	cmd = exec.Command("devpod", "stop", workSpaceId)
	err = getOutputCommand(cmd, sendStream)
	if err != nil {
		log.Error("Failing to add remote: %v", err)
	}

	// devpod delete <workspace-id>
	cmd = exec.Command("devpod", "delete", workSpaceId)
	err = getOutputCommand(cmd, sendStream)
	if err != nil {
		log.Error("Failing to add remote: %v", err)
	}

	// devpod provider delete <provider-id>
	cmd = exec.Command("devpod", "provider", "delete", providerId)
	err = getOutputCommand(cmd, sendStream)
	if err != nil {
		log.Error("Failing to add provider: %v", err)
	}

	// running python script
	cmd = exec.Command("devpod", "ssh", workSpaceId, "--command", "python3 main.py")
	if err = getOutputCommand(cmd, sendStream); err != nil {
		cmd = exec.Command("devpod", "ssh", workSpaceId, "--command", "python main.py")
		if err := getOutputCommand(cmd, sendStream); err != nil {
			log.Error("Failed to run main.py: %v", err)
		}
	}

	return err
}

// getOutputCommand is a helper function to get the stdout and stderr of a command and send it to the client
func getOutputCommand(cmd *exec.Cmd, sendStream func(string)) error {
	// Create a pipe for reading the command's output
	outReader, err := cmd.StdoutPipe()
	if err != nil {
		log.Error("Error creating StdoutPipe for Cmd: %v", err)
		return err
	}

	// Create a pipe for reading the command's stderr
	errReader, err := cmd.StderrPipe()
	if err != nil {
		log.Error("Error creating StderrPipe for Cmd: %v", err)
		return err
	}

	// Start the command
	err = cmd.Start()
	if err != nil {
		log.Error("Error starting Cmd: %v", err)
		return err
	}

	cmdReader := io.MultiReader(outReader, errReader)

	// Create a scanner to read the output
	scanner := bufio.NewScanner(cmdReader)
	scanner.Split(bufio.ScanLines)

	// Continuously send updates to the client
	for scanner.Scan() {
		// Send the output to the client as an SSE message
		sendStream(scanner.Text())
		time.Sleep(10 * time.Millisecond) // Add a delay to streaming
	}

	// get any errors encountered during scanning
	if err := scanner.Err(); err != nil {
		log.Error("Error with scanner: %v", err)
		return err
	}

	// Wait for the command to finish
	err = cmd.Wait()
	if err != nil {
		log.Error("Error waiting for Cmd: %v", err)
		return err
	}

	return nil
}

// GitCredential is a struct to handle git credentials with access token for private repository
// every git credential will be stored in a file as tmp and removed after use
// it also set environment variables for username and token & remove after use
// it also set git config credential.helper and remove after use

// why we need this?
//   - To access private repository gitea
//   - DevPod not support if use git clone with username and password in url and custom branch
//     Example use url not support for devpod:
//     ```devpod up http://username:xxxxxxxx@gitea.host@branch --provider providerId --ide none --id workSpaceId```
//     this will not work for devpod
//   - DevPod will forward git credentials to a remote machine so that you can also pull private repositories.

type GitCredential struct {
	Username      string
	Token         string
	WokSpaceId    string
	envUser       string
	envToken      string
	scriptPath    string
	scriptContent string
}

// NewGitCredential creates a new GitCredential with the given workspace ID, username, and token.
//
// Parameters:
// - WokSpaceId string
// - username string
// - token string
// Return type: *GitCredential
func NewGitCredential(WokSpaceId, username, token string) *GitCredential {
	g := &GitCredential{
		Username:   username,
		Token:      token,
		WokSpaceId: WokSpaceId,
	}
	g.Init()
	return g
}

// Init initializes the GitCredential.
//
// No parameters.
// No return type.
func (g *GitCredential) Init() {
	// set environment variables
	g.envUser = "GIT_USER_" + g.WokSpaceId
	g.envToken = "GIT_TOKEN_" + g.WokSpaceId

	// create a file to store the script
	tmpDir := os.TempDir()
	g.scriptPath = tmpDir + "/git_creds_" + g.WokSpaceId + ".sh"

	// script content to store in file
	g.scriptContent = fmt.Sprintf(`#!/bin/bash
sleep 1
echo username=$%s
echo password=$%s
`, g.envUser, g.envToken)
}

// Set writes the git credentials to a file, sets up the git config, and sets environment variables.
//
// Returns an error.
func (g *GitCredential) Set() error {
	// create a file to store the script
	err := os.WriteFile(g.scriptPath, []byte(g.scriptContent), 0o600)
	if err != nil {
		log.Error("Failing creating git creds file: %v", err)
		return err
	}

	// execute: git config --global credential.helper "/bin/bash /tmp/git_creds_<workspaceid>.sh"
	cmd := exec.Command("git", "config", "--global", "credential.helper", "/bin/bash "+g.scriptPath)
	if err := cmd.Run(); err != nil {
		log.Error("Failing to set git creds: %v", err)
		return err
	}

	// set environment variables
	os.Setenv(g.envUser, g.Username)
	os.Setenv(g.envToken, g.Token)

	return nil
}

// Remove removes environment variables, git creds file from tmp, and git config credential.helper.
//
// No parameters.
// Returns an error.
func (g *GitCredential) Remove() error {
	// remove environment variables
	os.Unsetenv(g.envUser)
	os.Unsetenv(g.envToken)

	// remove git creds file from tmp
	err := os.Remove(g.scriptPath)
	if err != nil {
		log.Error("Failing to remove git creds file: %v", err)
	}
	// remove git config credential.helper
	// execute: git config --global --unset credential.helper
	err = exec.Command("git", "config", "--global", "--unset", "credential.helper").Run()
	if err != nil {
		log.Error("Failing to remove git creds file: %v", err)
	}
	return err
}
