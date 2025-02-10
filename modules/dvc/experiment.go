package dvc

import (
	"html/template"
	"os/exec"

	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/modules/markup"
	"code.gitea.io/gitea/modules/markup/markdown"
	"code.gitea.io/gitea/services/context"
	"github.com/go-git/go-git/v5"
)

func ExperimentHtml(ctx *context.Context) (html template.HTML, err error) {

	err = executeTempRepo(ctx, func(tempRepoPath string, repository *git.Repository) error {
		cmd := exec.Command("dvc", "exp", "show", "--md", "--all-commits")
		cmd.Dir = tempRepoPath
		out, err := cmd.Output() // dont' combine output

		if err != nil {
			log.Error("error when dvc exp show: %v", err)
			return err
		}

		html, err = markdown.RenderString(&markup.RenderContext{}, string(out))

		if err != nil {
			log.Error("error when markup to html: %v", err)
			return err
		}
		return nil
	})

	return html, err
}
