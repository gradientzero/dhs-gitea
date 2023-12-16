package org

import (
	"code.gitea.io/gitea/models/organization"
	"code.gitea.io/gitea/modules/base"
	"code.gitea.io/gitea/modules/context"
	"net/http"
	"strconv"
)

const (
	tplSettingsDevpodCredentialList base.TplName = "org/settings/devpod_credential_list"
	tplSettingsDevpodCredentialForm base.TplName = "org/settings/devpod_credential_form"
	tplSettingsDevpodCredentialEdit base.TplName = "org/settings/devpod_credential_edit"
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
		remote := ctx.FormString("remote")
		key := ctx.FormString("key")
		value := ctx.FormString("value")
		err := organization.AddDevpodCredential(ctx.Org.Organization.ID, remote, key, value)
		if err != nil {
			ctx.Flash.Error("error on saving credential")
			ctx.Redirect(ctx.Org.OrgLink + "/settings/devpod-credential/new")
		}

		ctx.Flash.Success("New credential added successfully")
		ctx.Redirect(ctx.Org.OrgLink + "/settings/devpod-credential")
	}

	ctx.HTML(http.StatusOK, tplSettingsDevpodCredentialForm)
}

func SettingsDevpodCredentialEdit(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("org.settings")
	ctx.Data["PageIsOrgSettings"] = true
	ctx.Data["PageIsSettingsDevpodCredential"] = true

	id := ctx.FormInt64("id")

	credential := &organization.OrgDevpodCredential{}
	// check id exist or not
	credential, err := organization.GetDevpodCredentialById(id, ctx.Org.Organization.ID)
	if err != nil {
		ctx.Flash.Error("error loading credential data")
		ctx.Redirect(ctx.Org.OrgLink + "/settings/devpod-credential/edit")
	}

	if ctx.Req.Method == "POST" {
		remote := ctx.FormString("remote")
		key := ctx.FormString("key")
		value := ctx.FormString("value")

		err := organization.UpdateDevpodCredential(ctx.Org.Organization.ID, id,
			remote, key, value)

		if err != nil {
			ctx.Flash.Error("error on saving credential")
			ctx.Redirect(ctx.Org.OrgLink + "/settings/devpod-credential/edit?id=" + strconv.FormatInt(id, 10))
		}

		ctx.Flash.Success("Credential updated successfully")
		ctx.Redirect(ctx.Org.OrgLink + "/settings/devpod-credential")
	}

	ctx.Data["ID"] = id
	ctx.Data["Credential"] = credential

	ctx.HTML(http.StatusOK, tplSettingsDevpodCredentialEdit)
}

func SettingsDevpodCredentialDelete(ctx *context.Context) {
	id := ctx.FormInt64("id")
	_ = organization.DeleteOrgDevpodCredential(id)
	ctx.Flash.Warning(ctx.Tr("org.settings.devpod_credential_deleted"))
	ctx.Redirect(ctx.Org.OrgLink + "/settings/devpod-credential")
}
