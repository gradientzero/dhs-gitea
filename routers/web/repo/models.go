package repo

import (
	"net/http"

	"code.gitea.io/gitea/modules/base"
	"code.gitea.io/gitea/modules/cache"
	"code.gitea.io/gitea/modules/dvc"
	"code.gitea.io/gitea/services/context"
)

const (
	tplModelsList     base.TplName = "repo/models/list"
	modelCacheTimeout int64        = 60 * 5 // 5 minutes
)

// MustEnableModels check if projects are enabled in settings
func MustEnableModels(ctx *context.Context) {
	// TODO: later enable models in project settings
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

// Models show list of models in projects
// this tab no need to chose branch, because gto release is based on git tag
func Models(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("repo.models")
	ctx.Data["IsModelPage"] = true // to show highlight in tab

	result := [][]string{}
	var err error
	var cached bool
	cc := cache.GetCache()
	cached, _ = cc.GetJSON("dvc-models", &result)

	if !cached {
		result, err = dvc.ReleaseModel(ctx)
		if err != nil {
			ctx.Flash.Error(err.Error(), true)
		}

		_ = cc.PutJSON("dvc-models", result, modelCacheTimeout)
	}

	if len(result) > 1 {
		ctx.Data["Header"] = result[0]
		ctx.Data["Data"] = result[1:]
	}

	ctx.HTML(http.StatusOK, tplModelsList)
}
