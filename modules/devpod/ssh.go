package devpod

import (
	"code.gitea.io/gitea/modules/log"
	"github.com/melbahja/goph"
	"golang.org/x/crypto/ssh"
)

func ExecuteInRemoteMachine(privateKey string, user string,
	host string, port uint,
	callback func(client goph.Client) error) {

	auth, err := goph.RawKey(privateKey, "")
	if err != nil {
		log.Fatal("%v", err)
	}

	client, err := goph.NewConn(&goph.Config{
		User:     user,
		Addr:     host,
		Port:     port,
		Auth:     auth,
		Timeout:  goph.DefaultTimeout,
		Callback: ssh.InsecureIgnoreHostKey(),
	})

	if err != nil {
		log.Fatal("%v", err)
	}

	// Defer closing the network connection.
	defer client.Close()

	// Execute your command.
	err = callback(*client)

	if err != nil {
		log.Fatal("%v", err)
	}
}

func BuildCommands(repoUrl, branch, tempDir string) []string {
	return []string{
		"git clone " + repoUrl + " -b " + branch + " " + tempDir,
		"cd " + tempDir,
		"devpod up . --ide none",
		"devpod ssh . --command \"dvc pull\"",
		"devpod ssh . --command \"dvc exp run\"",
	}
}
