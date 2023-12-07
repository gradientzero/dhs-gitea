package org

import (
	"code.gitea.io/gitea/models/organization"
	"code.gitea.io/gitea/modules/base"
	"code.gitea.io/gitea/modules/context"
	"net/http"
)

const (
	tplSettingsDevpodCredentialList base.TplName = "org/settings/devpod-credential-list"
	tplSettingsDevpodCredentialForm base.TplName = "org/settings/devpod-credential-form"
)

func SettingsDevpodCredentialList(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("org.settings")
	ctx.Data["PageIsOrgSettings"] = true
	ctx.Data["PageIsSettingsDevpodCredential"] = true

	credentials, err := organization.GetOrgDevpodCredential(ctx.Org.Organization.ID)
	if err != nil {
		ctx.Flash.Error("error on loading tokens")
	}
	ctx.Data["Credentials"] = credentials

	ctx.HTML(http.StatusOK, tplSettingsDevpodCredentialList)
}

func SettingsDevpodCredentialCreate(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("org.settings")
	ctx.Data["PageIsOrgSettings"] = true
	ctx.Data["PageIsSettingsDevpodCredential"] = true

	if ctx.Req.Method == "POST" {
		name := ctx.FormString("name")
		key := ctx.FormString("key")
		value := ctx.FormString("value")
		err := organization.AddDevpodCredential(ctx.Org.Organization.ID, name, key, value)
		if err != nil {
			ctx.Flash.Error("error on saving credential")
			ctx.Redirect(ctx.Org.OrgLink + "/settings/devpod-credential/new")
		}

		ctx.Flash.Success("New credential added successfully")
		ctx.Redirect(ctx.Org.OrgLink + "/settings/devpod-credential")
	}

	ctx.HTML(http.StatusOK, tplSettingsDevpodCredentialForm)
}

func SettingsDevpodCredentialDelete(ctx *context.Context) {
	id := ctx.FormInt64("id")
	_ = organization.DeleteOrgDevpodCredential(id)
	ctx.Flash.Warning(ctx.Tr("org.settings.devpod_credential_deleted"))
	ctx.Redirect(ctx.Org.OrgLink + "/settings/devpod-credential")
}
