package org

import (
	"net/http"

	"code.gitea.io/gitea/models/organization"
	"code.gitea.io/gitea/modules/base"
	"code.gitea.io/gitea/services/context"
)

const (
	tplSettingsGiteaTokenList base.TplName = "org/settings/gitea-token-list"
	tplSettingsGiteaTokenForm base.TplName = "org/settings/gitea-token-form"
)

func SettingsGiteaTokenList(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("org.settings")
	ctx.Data["PageIsOrgSettings"] = true
	ctx.Data["PageIsSettingsGiteaToken"] = true

	tokens, err := organization.GetOrgGiteaToken(ctx.Org.Organization.ID)
	if err != nil {
		ctx.Flash.Error("error on loading tokens")
	}
	ctx.Data["Tokens"] = tokens

	ctx.HTML(http.StatusOK, tplSettingsGiteaTokenList)
}

func SettingsGiteaTokenCreate(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("org.settings")
	ctx.Data["PageIsOrgSettings"] = true
	ctx.Data["PageIsSettingsGiteaToken"] = true

	if ctx.Req.Method == "POST" {
		name := ctx.FormString("name")
		token := ctx.FormString("token")
		err := organization.AddGiteaToken(ctx.Org.Organization.ID, name, token)
		if err != nil {
			ctx.Flash.Error("error on saving token")
			ctx.Redirect(ctx.Org.OrgLink + "/settings/gitea-token/new")
		}

		ctx.Flash.Success("New token added successfully")
		ctx.Redirect(ctx.Org.OrgLink + "/settings/gitea-token")
	}

	ctx.HTML(http.StatusOK, tplSettingsGiteaTokenForm)
}

func SettingsGiteaTokenDelete(ctx *context.Context) {
	id := ctx.FormInt64("id")
	_ = organization.DeleteOrgGiteaToken(id)
	ctx.Flash.Warning(ctx.Tr("org.settings.gitea_token_deleted"))
	ctx.Redirect(ctx.Org.OrgLink + "/settings/gitea-token")
}
