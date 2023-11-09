package repo

import (
	"code.gitea.io/gitea/modules/base"
	"code.gitea.io/gitea/modules/context"
	"code.gitea.io/gitea/modules/dvc"
	"code.gitea.io/gitea/modules/log"
	"net/http"
	"slices"
)

const (
	tplExperimentsList base.TplName = "repo/experiments/list"
)

// MustEnableExperiments check if projects are enabled in settings
func MustEnableExperiments(ctx *context.Context) {
	// TODO: later enable datasets in project settings
	/*if unit.TypeDatasets.UnitGlobalDisabled() {
		ctx.NotFound("EnableDatasets", nil)
		return
	}

	if ctx.Repo.Repository != nil {
		if !ctx.Repo.CanRead(unit.TypeDatasets) {
			ctx.NotFound("MustEnableDatasets", nil)
			return
		}
	}*/
}

// Experiments show list of dataset in projects
func Experiments(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("repo.experiments")
	ctx.Data["IsExperimentPage"] = true // to show highlight in tab

	branches, err := findBranches(ctx)
	if err != nil {
		log.Error("err when finding branches: %v", err)
	}

	branch := ctx.Req.URL.Query().Get("branch")
	if branch != "" && slices.Contains(branches, branch) {
		ctx.Repo.BranchName = branch
	} else {
		branch = ctx.Repo.BranchName // reset default branch name
	}

	ctx.Data["Branches"] = branches
	ctx.Data["Branch"] = branch

	ctx.HTML(http.StatusOK, tplExperimentsList)
}

func ExperimentTable(ctx *context.Context) {

	branches, err := findBranches(ctx)
	if err != nil {
		log.Error("err when finding branches: %v", err)
	}

	branch := ctx.Req.URL.Query().Get("branch")
	if branch != "" && slices.Contains(branches, branch) {
		ctx.Repo.BranchName = branch
	} else {
		branch = ctx.Repo.BranchName // reset default branch name
	}

	html, err := dvc.ExperimentHtml(ctx)
	if err != nil {
		log.Error("err when dvc experiment to html: %v", err)
	}
	//ctx.Data["Experiments"] = html

	ctx.JSON(http.StatusOK, map[string]any{
		"htmlTable": html,
	})
}
