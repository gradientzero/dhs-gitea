package dvc

import (
	"bufio"
	"os/exec"
	"regexp"
	"strings"

	"code.gitea.io/gitea/services/context"
	"github.com/go-git/go-git/v5"
)

func ParseGtoRelease(dir string) ([][]string, error) {
	cmd := exec.Command("gto", "show", "--plain")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile("\\s+")
	output := string(out)
	var result [][]string

	sc := bufio.NewScanner(strings.NewReader(output))
	for sc.Scan() {
		split := re.Split(sc.Text(), -1)
		result = append(result, split)
	}

	return result, nil
}

func ReleaseModel(ctx *context.Context) (result [][]string, err error) {
	err = executeTempRepo(ctx, func(tempRepoPath string, repository *git.Repository) error {
		result, err = ParseGtoRelease(tempRepoPath)
		return err
	})
	return result, err
}
