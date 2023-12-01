package devpod

import (
	"code.gitea.io/gitea/modules/log"
	"fmt"
	"github.com/rs/xid"
	"os"
	"os/exec"
	"strconv"
)

// TODO: 1. need to forward S3 credentials
// TODO: 2. need to find git credential handling for pushing back
// not sure concurrent safety for devpod for now

func Execute(privateKey string) {

	user := "vagrant"
	host := "localhost"
	var port uint = 2222
	gitUrl := "git@gitlab.com:grz1/aqua-research.git" // NOTE: make sure to trim, will get surprise if there is space

	providerId := xid.New().String()
	workSpaceId := xid.New().String()

	// setting up private key for use
	privateKeyFile := "/tmp/" + providerId + ".key"
	fmt.Println(privateKeyFile + " will created")

	// TODO: make sure append new line end of key
	err := os.WriteFile(privateKeyFile, []byte(privateKey), 0600)
	if err != nil {
		log.Error("Failing creating private key file: %v", err)
	}

	defer func() {
		err := os.Remove(privateKeyFile)
		if err != nil {
			log.Error("Failing to remove private key file: %v", err)
		}
	}()

	var result string

	// devpod provider add ssh --name <provider-id> -o HOST=vagrant@localhost -o PORT=2222 -o EXTRA_FLAGS="-i /tmp/private_key"
	cmd := exec.Command("devpod", "provider", "add", "ssh", "--name", providerId,
		"-o", "HOST="+user+"@"+host, "-o", "PORT="+strconv.Itoa(int(port)),
		"-o", "EXTRA_FLAGS=-i "+privateKeyFile) // Don't use \" in EXTRA_FLAGS for some reason it not working
	cmd.Env = os.Environ()
	cmd.Dir = "/tmp"
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error("error when devpod ssh provider add: %v", err)
	}
	result += string(output)

	// devpod up --provider <provider-id> <git-url> --ide none --debug --id <workspace-id>
	cmd = exec.Command("devpod", "up", gitUrl, "--provider", providerId, "--ide", "none", "--id", workSpaceId)
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Error("error when devpod up: %v", err)
	}
	result += string(output)

	//devpod ssh <workspace-id> --command 'dvc pull'
	cmd = exec.Command("devpod", "ssh", workSpaceId, "--command", "dvc pull")
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Error("error when dvc pull: %v", err)
	}
	result += string(output)

	//devpod ssh <workspace-id> --command 'dvc exp run'
	cmd = exec.Command("devpod", "ssh", workSpaceId, "--command", "dvc exp run")
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Error("error when dvc exp run: %v", err)
	}
	result += string(output)

	//devpod ssh <workspace-id> --command 'git add .'
	cmd = exec.Command("devpod", "ssh", workSpaceId, "--command", "git add .")
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Error("error when git add: %v", err)
	}
	result += string(output)

	//devpod ssh <workspace-id> --command 'git commit -m "exp run result"'
	cmd = exec.Command("devpod", "ssh", workSpaceId, "--command", "git commit -m 'exp run result'")
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Error("error when git commit: %v", err)
	}
	result += string(output)

	//devpod ssh <workspace-id> --command 'git push origin'
	cmd = exec.Command("devpod", "ssh", workSpaceId, "--command", "git push origin")
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Error("error when git push: %v", err)
	}
	result += string(output)

	//devpod stop <workspace-id>
	cmd = exec.Command("devpod", "stop", workSpaceId)
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Error("error when devpod stop: %v", err)
	}
	result += string(output)

	//devpod delete <workspace-id>
	cmd = exec.Command("devpod", "delete", workSpaceId)
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Error("error when devpod delete: %v", err)
	}
	result += string(output)

	// devpod provider delete <provider-id>
	cmd = exec.Command("devpod", "provider", "delete", providerId)
	//cmd.Dir = tempRepoPath
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Error("error when devpod ssh provider remove: %v", err)
	}
	result += string(output)

	fmt.Println(result)
}
