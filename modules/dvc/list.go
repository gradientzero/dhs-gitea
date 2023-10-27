package dvc

import (
	"bufio"
	"code.gitea.io/gitea/modules/context"
	"code.gitea.io/gitea/modules/log"
	"github.com/go-git/go-git/v5"
	"os/exec"
	"strings"
)

func FileList(ctx *context.Context) (files []string, err error) {

	err = executeTempRepo(ctx, func(tempRepoPath string, repository *git.Repository) error {
		cmd := exec.Command("dvc", "list", "--dvc-only", "-R", ".")
		cmd.Dir = tempRepoPath
		out, err := cmd.CombinedOutput()

		if err != nil {
			log.Error("error when dvc list: %v", err)
			return err
		}

		sc := bufio.NewScanner(strings.NewReader(string(out)))
		for sc.Scan() {
			files = append(files, sc.Text())
		}
		return nil
	})

	return files, err
}
