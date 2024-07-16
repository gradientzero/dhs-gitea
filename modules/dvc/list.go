package dvc

import (
	"os/exec"

	"code.gitea.io/gitea/modules/json"
	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/services/context"
	"github.com/go-git/go-git/v5"
)

type File struct {
	// https://stackoverflow.com/questions/10998222/json-parsing-of-int64-in-go-null-values
	// use *int64 instead of int64 to check null value
	Size *uint64
	Path string
}

// ParseJsonFileInfo Parse Json File Info to DvcFile slices
func ParseJsonFileInfo(jsonStr []byte) (result []File, err error) {
	var files []File

	err = json.Unmarshal(jsonStr, &files)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.Size != nil {
			result = append(result, file)
		}
	}
	return result, nil
}

func FileList(ctx *context.Context) (files []File, err error) {
	err = executeTempRepo(ctx, func(tempRepoPath string, repository *git.Repository) error {
		cmd := exec.Command("dvc", "list", "--dvc-only", "-R", "--size", "--json", ".")
		cmd.Dir = tempRepoPath
		jsonOut, err := cmd.Output()

		if err != nil {
			log.Error("error when dvc list: %v", err)
			return err
		}

		files, err = ParseJsonFileInfo(jsonOut)
		return err
	})

	return files, err
}
