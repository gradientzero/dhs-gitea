package devpod

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	org_model "code.gitea.io/gitea/models/organization"
	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/services/context"
	"github.com/buildkite/terminal-to-html/v3"
	"github.com/rs/xid"
)

func Execute(
	ctx *context.Context,
	isRepoPrivate bool,
	privateKey, user, host string,
	port int32,
	gitUrl, gitBranch, gitUser, gitEmail, gitToken string,
	devPodCredentials map[string][]org_model.OrgDevpodCredential,
	_sendStream func(string),
) error {

	// create a folder to store runs
	runID := time.Now().Format("20060102-150405")
	repoPath := filepath.Join(RunsFolder, fmt.Sprintf("repo-%d", ctx.Repo.Repository.ID))
	runFile := filepath.Join(repoPath, fmt.Sprintf("run-%s.log", runID))

	var sendStream = func(result string) {
		if len(gitToken) > 0 {
			result = strings.ReplaceAll(result, gitToken, "***")
		}
		output := terminal.Render([]byte(result))
		ctx.Write([]byte(string(output) + "\n"))
		if err := LogRunMessageToFile(runFile, result); err != nil {
			log.Error("Failed to write to run file: %v", err)
		}
		ctx.Resp.Flush()
	}

	// Create the file and write initial content
	if err := CreateRunFile(runFile); err != nil {
		log.Error("Failed to create run file: %v", err)
	}

	sendStream("#===================================")
	sendStream("# Starting Run ...")
	sendStream("#===================================")

	providerId := xid.New().String()
	workSpaceId := xid.New().String()
	sendStream("ProviderId: " + providerId)
	sendStream("WorkSpaceId: " + workSpaceId)

	// git credential handling
	// if gitToken is available then create git credential file and set it
	if gitToken != "" {
		log.Info("Creating git creds for devpod: %v", workSpaceId)
		g := NewGitCredential(workSpaceId, gitUser, gitToken)
		if err := g.Set(); err != nil {
			log.Error("Failing creating git creds file: %v", err)
			sendStream(err.Error())
		}
		defer g.Remove()
	}
	sendStream("Using Git Credentials:")
	sendStream("- Git User: " + gitUser)
	sendStream("- Access Token: ***")

	// setting up private key for use
	privateKeyFile := "/tmp/" + providerId + ".key"
	// Important for line ending for correct
	privateKey = strings.Replace(privateKey, "\r\n", "\n", -1)

	// also make sure there is new line in end of file
	if err := os.WriteFile(privateKeyFile, []byte(privateKey+"\n"), 0o600); err != nil {
		log.Error("Failing creating private key file: %v", err)
		sendStream(err.Error())
	}
	sendStream("Copied Private SSH Key:" + privateKeyFile)

	defer func() {
		if err := os.Remove(privateKeyFile); err != nil {
			log.Error("Failing to remove private key file: %v", err)
			sendStream(err.Error())
		}
	}()
	sendStream("")

	sendStream("#===================================")
	sendStream("# Creating Provider ...")
	sendStream("#===================================")

	var err error
	var cmd *exec.Cmd
	createSSHProvider := !strings.Contains(gitUrl, "localhost")

	if createSSHProvider {
		// create SSH provider (real provider):
		// devpod provider add ssh --name providerId -o HOST=user@host -o PORT=port -o EXTRA_FLAGS=-i /tmp/providerId.key
		cmd = exec.Command(
			"devpod", "provider", "add", "ssh",
			"--name", providerId,
			"-o", "HOST="+user+"@"+host,
			"-o", "PORT="+strconv.Itoa(int(port)),
			"-o", "EXTRA_FLAGS=-i "+privateKeyFile,
		)
	} else {
		// create internal docker provider:
		// devpod provider add docker --name providerId
		cmd = exec.Command("devpod", "provider", "add", "docker", "--name", providerId)
	}
	cmd.Env = os.Environ()
	cmd.Dir = "/tmp"
	sendStream("Running Command: ")
	sendStream("- " + cmd.String())
	sendStream("Output: ")
	if err := readCmdOutput(cmd, sendStream); err != nil {
		log.Error("Failing to add provider: %v", err)
		sendStream(err.Error())
	}
	sendStream("")

	sendStream("#===================================")
	sendStream("# Creating Workspace ...")
	sendStream("#===================================")
	sendStream("Git Repo Private: " + fmt.Sprintf("%t", isRepoPrivate))
	if isRepoPrivate {
		// user:accesstoken@host:port schema for private repository
		gitUrl = strings.Replace(gitUrl, "https://", "https://"+gitUser+":"+gitToken+"@", -1)
		gitUrl = strings.Replace(gitUrl, "http://", "http://"+gitUser+":"+gitToken+"@", -1)
	}
	// add branch to gitUrl
	gitUrl = gitUrl + "@" + gitBranch

	sendStream("Git Repo Url: " + gitUrl)
	sendStream("Git Repo Branch: " + gitBranch)

	// devpod up test-workspace --source=git:ssh://git@sandbox.gradient0.com:2221/sandbox/dvc.git --provider sandbox-remote-ssh --ide none --id test-workspace
	cmd = exec.Command(
		"devpod", "up", workSpaceId,
		"--source=git:"+gitUrl,
		"--provider", providerId,
		"--ide", "none",
		"--id", workSpaceId,
		"--debug",
	)
	sendStream("Running Command: ")
	sendStream("- " + cmd.String())
	if err = readCmdOutput(cmd, sendStream); err != nil {
		log.Error("Failing to up devpod: %v", err)
		sendStream(err.Error())
	}

	if len(devPodCredentials) > 0 {
		sendStream("")
		sendStream("#===================================")
		sendStream("# DevPod Credentials ...")
		sendStream("#===================================")
		for remoteName, keyValPair := range devPodCredentials {
			sendStream("Remote: " + remoteName)
			for _, value := range keyValPair {
				sendStream(" - " + value.Key + ": ***")
			}
		}
		fileContent := ""
		for remoteName, keyValPair := range devPodCredentials {
			fileContent += fmt.Sprintf("['remote \"%s\"']\n", remoteName)
			for _, value := range keyValPair {
				fileContent += fmt.Sprintf("    %s = %s\n", value.Key, value.Value)
			}
		}
		if len(fileContent) > 0 {
			// base64 encode the content of .dvc/config.local
			// Reason: we cannot pass content with single and double quotes
			// in command; additionally devpod disallows stdin which would also be an option
			// echo "Test" | base64 -i - -o .dvc/config.local.b64
			// devpod ssh ${WORKSPACE_ID} \
			// 		--command "echo \"$(cat .dvc/config.local.b64)\" > .dvc/config.local.b64 && base64 -d .dvc/config.local.b64 > .dvc/config.local && rm .dvc/config.local.b64"
			// rm .dvc/config.local.b64
			fileContentB64 := base64.StdEncoding.EncodeToString([]byte(fileContent))
			_cmd := fmt.Sprintf("echo \"%s\" > .dvc/config.local.b64 && base64 -d .dvc/config.local.b64 > .dvc/config.local && rm .dvc/config.local.b64", fileContentB64)
			cmd = exec.Command("devpod", "ssh", workSpaceId, "--command", _cmd, "--silent")
			sendStream("Running Command: ")
			sendStream("- " + strings.Replace(cmd.String(), fileContentB64, "<base64:***>", -1))
			if err := readCmdOutput(cmd, sendStream); err != nil {
				log.Error(err.Error())
				sendStream(err.Error())
			}
		}
	}

	sendStream("")
	sendStream("#===================================")
	sendStream("# Find & Execute run.sh ...")
	sendStream("#===================================")
	// devpod ssh <workspace-id> --command 'run.sh'
	cmd = exec.Command("devpod", "ssh", workSpaceId, "--command", "chmod +x run.sh && ./run.sh", "--silent")
	sendStream("Running Command: ")
	sendStream("- " + cmd.String())
	if err = readCmdOutput(cmd, sendStream); err != nil {
		log.Error(err.Error())
		sendStream(err.Error())
	}

	sendStream("")
	sendStream("#===================================")
	sendStream("# Config Git ...")
	sendStream("#===================================")
	//devpod ssh ${WORKSPACE_ID} --command "echo '${SANDBOX_PROTOCOL}://${GIT_USER}:${GIT_ACCESS_TOKEN}@${SANDBOX_HOST_DST}' > ~/.git-credentials"

	extractBaseURL := func(url string) string {
		re := regexp.MustCompile(`^(https?://[^/]+)`)
		match := re.FindStringSubmatch(url)
		if len(match) > 1 {
			return match[1]
		}
		return ""
	}
	srcBaseUrl := extractBaseURL(gitUrl)
	dstBaseUrl := extractBaseURL(gitUrl)

	// host of remote machine is localhost - usually used for development with docker
	// => because DevPod runs within docker we have change from localhost to host.docker.internal
	// => this is only needed for the git credentials when pushing results to origin (=host.docker.internal)
	remoteMachineIsLocalhost := strings.Contains(srcBaseUrl, "localhost:")
	if remoteMachineIsLocalhost {
		dstBaseUrl = strings.Replace(srcBaseUrl, "localhost", "host.docker.internal", -1)
	}
	// devpod ssh <workspace-id> --command 'echo 'http://user:token@host:port' > ~/.git-credentials'
	// devpod ssh <workspace-id> --command 'git config credential.helper store'
	_cmd := fmt.Sprintf("echo \"%s\" > ~/.git-credentials && git config credential.helper store", dstBaseUrl)
	cmd = exec.Command("devpod", "ssh", workSpaceId, "--command", _cmd, "--silent")
	sendStream("Running Command: ")
	sendStream("- " + cmd.String())
	if err = readCmdOutput(cmd, sendStream); err != nil {
		log.Error(err.Error())
		sendStream(err.Error())
	}

	// if remote machine is localhost we need to change the git config url to host.docker.internal, too
	if remoteMachineIsLocalhost {
		// devpod ssh <workspace-id> --command "git config url."http://user:token@host.docker.internal:3001/".insteadOf "http://user:token@localhost:3001/""
		_cmd = fmt.Sprintf("git config url.\"%s\".insteadOf \"%s\"", dstBaseUrl, srcBaseUrl)
		cmd = exec.Command("devpod", "ssh", workSpaceId, "--command", _cmd, "--silent")
		sendStream("Running Command: ")
		sendStream("- " + cmd.String())
		if err = readCmdOutput(cmd, sendStream); err != nil {
			log.Error(err.Error())
			sendStream(err.Error())
		}
	}

	// devpod ssh <workspace-id> --command 'git config user.name xxx'
	// devpod ssh <workspace-id> --command 'git config user.email xxx'
	_cmd = fmt.Sprintf("git config user.name \"%s\" && git config user.email \"%s\"", gitUser, gitEmail)
	cmd = exec.Command("devpod", "ssh", workSpaceId, "--command", _cmd, "--silent")
	sendStream("Running Command: ")
	sendStream("- " + cmd.String())
	if err = readCmdOutput(cmd, sendStream); err != nil {
		log.Error(err.Error())
		sendStream(err.Error())
	}

	sendStream("")
	sendStream("#===================================")
	sendStream("# Pushing Results ...")
	sendStream("#===================================")
	// devpod ssh <workspace-id> --command 'git add .'
	// devpod ssh <workspace-id> --command 'git commit -m "exp run result"'
	// devpod ssh <workspace-id> --command 'git push origin'

	commitMsg := "exp run result"
	_cmd = fmt.Sprintf("git add . && git commit -m \"%s\" && git push origin", commitMsg)
	cmd = exec.Command("devpod", "ssh", workSpaceId, "--command", _cmd, "--silent")
	sendStream("Running Command: ")
	sendStream("- " + cmd.String())
	if err = readCmdOutput(cmd, sendStream); err != nil {
		log.Error(err.Error())
		sendStream(err.Error())
	}

	skipCleanup := false
	if skipCleanup {
		return nil
	}
	sendStream("")
	sendStream("#===================================")
	sendStream("# Cleanup ...")
	sendStream("#===================================")
	// devpod stop <workspace-id>
	cmd = exec.Command("devpod", "stop", workSpaceId)
	sendStream("Running Command: ")
	sendStream("- " + cmd.String())
	if err = readCmdOutput(cmd, sendStream); err != nil {
		log.Error(err.Error())
		sendStream(err.Error())
	}

	// devpod delete <workspace-id>
	cmd = exec.Command("devpod", "delete", workSpaceId)
	sendStream("Running Command: ")
	sendStream("- " + cmd.String())
	if err = readCmdOutput(cmd, sendStream); err != nil {
		log.Error(err.Error())
		sendStream(err.Error())
	}

	// devpod provider delete <provider-id>
	cmd = exec.Command("devpod", "provider", "delete", providerId)
	sendStream("Running Command: ")
	sendStream("- " + cmd.String())
	if err = readCmdOutput(cmd, sendStream); err != nil {
		log.Error(err.Error())
		sendStream(err.Error())
	}

	sendStream("Compute done.")
	return err
}

func readCmdOutput(cmd *exec.Cmd, sendStream func(string)) error {
	// Create pipes for stdout and stderr
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating StdoutPipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("error creating StderrPipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting command: %w", err)
	}

	// Use goroutines to read stdout and stderr concurrently
	outputChan := make(chan string)
	go streamOutput(stdoutPipe, outputChan)
	go streamOutput(stderrPipe, outputChan)

	// Read from the channel and send output to the stream function
	go func() {
		for output := range outputChan {
			sendStream(output)
		}
	}()

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		close(outputChan)
		return fmt.Errorf("command execution failed: %w", err)
	}

	close(outputChan)
	return nil
}

func streamOutput(pipe io.ReadCloser, outputChan chan<- string) {
	scanner := bufio.NewScanner(pipe)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		outputChan <- scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		outputChan <- fmt.Sprintf("Error reading output: %v", err)
	}
}

// getOutputCommand is a helper function to get the stdout and stderr of a command and send it to the client
func getOutputCommand(cmd *exec.Cmd, sendStream func(string)) error {
	// Create a pipe for reading the command's output
	outReader, err := cmd.StdoutPipe()
	if err != nil {
		log.Error("Error creating StdoutPipe for Cmd: %v", err)
		sendStream(fmt.Sprintf("Error creating StdoutPipe for Cmd: %v", err))
		return err
	}

	// Create a pipe for reading the command's stderr
	errReader, err := cmd.StderrPipe()
	if err != nil {
		log.Error("Error creating StderrPipe for Cmd: %v", err)
		sendStream(fmt.Sprintf("Error creating StderrPipe for Cmd: %v", err))
		return err
	}

	// Start the command
	err = cmd.Start()
	if err != nil {
		log.Error("Error starting Cmd: %v", err)
		sendStream(fmt.Sprintf("Error starting Cmd: %v", err))
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
		sendStream(fmt.Sprintf("Error scanner: %v", err))
		return err
	}

	// Wait for the command to finish
	err = cmd.Wait()
	if err != nil {
		log.Error("Error waiting for Cmd: %v", err)
		sendStream(fmt.Sprintf("Error waiting for Cmd: %v", err))
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
