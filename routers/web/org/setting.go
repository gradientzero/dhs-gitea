// Copyright 2014 The Gogs Authors. All rights reserved.
// Copyright 2019 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package org

import (
	"code.gitea.io/gitea/models/asymkey"
	"code.gitea.io/gitea/models/organization"
	"net/http"
	"net/url"
	"strings"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/models/db"
	repo_model "code.gitea.io/gitea/models/repo"
	user_model "code.gitea.io/gitea/models/user"
	"code.gitea.io/gitea/models/webhook"
	"code.gitea.io/gitea/modules/base"
	"code.gitea.io/gitea/modules/context"
	"code.gitea.io/gitea/modules/log"
	repo_module "code.gitea.io/gitea/modules/repository"
	"code.gitea.io/gitea/modules/setting"
	"code.gitea.io/gitea/modules/web"
	shared_user "code.gitea.io/gitea/routers/web/shared/user"
	user_setting "code.gitea.io/gitea/routers/web/user/setting"
	"code.gitea.io/gitea/services/forms"
	org_service "code.gitea.io/gitea/services/org"
	repo_service "code.gitea.io/gitea/services/repository"
	user_service "code.gitea.io/gitea/services/user"
)

const (
	// tplSettingsOptions template path for render settings
	tplSettingsOptions base.TplName = "org/settings/options"

	tplSettingsSshKeyList base.TplName = "org/settings/ssh-key-list"
	tplSettingsSshKeyForm base.TplName = "org/settings/ssh-key-form"

	tplSettingsMachineList base.TplName = "org/settings/machine-list"
	tplSettingsMachineForm base.TplName = "org/settings/machine-form"
	tplSettingsMachineEdit base.TplName = "org/settings/machine-edit"

	// tplSettingsDelete template path for render delete repository
	tplSettingsDelete base.TplName = "org/settings/delete"
	// tplSettingsHooks template path for render hook settings
	tplSettingsHooks base.TplName = "org/settings/hooks"
	// tplSettingsLabels template path for render labels settings
	tplSettingsLabels base.TplName = "org/settings/labels"
)

// Settings render the main settings page
func Settings(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("org.settings")
	ctx.Data["PageIsOrgSettings"] = true
	ctx.Data["PageIsSettingsOptions"] = true
	ctx.Data["CurrentVisibility"] = ctx.Org.Organization.Visibility
	ctx.Data["RepoAdminChangeTeamAccess"] = ctx.Org.Organization.RepoAdminChangeTeamAccess
	ctx.Data["ContextUser"] = ctx.ContextUser

	err := shared_user.LoadHeaderCount(ctx)
	if err != nil {
		ctx.ServerError("LoadHeaderCount", err)
		return
	}

	ctx.HTML(http.StatusOK, tplSettingsOptions)
}

// SettingsPost response for settings change submitted
func SettingsPost(ctx *context.Context) {
	form := web.GetForm(ctx).(*forms.UpdateOrgSettingForm)
	ctx.Data["Title"] = ctx.Tr("org.settings")
	ctx.Data["PageIsOrgSettings"] = true
	ctx.Data["PageIsSettingsOptions"] = true
	ctx.Data["CurrentVisibility"] = ctx.Org.Organization.Visibility

	if ctx.HasError() {
		ctx.HTML(http.StatusOK, tplSettingsOptions)
		return
	}

	org := ctx.Org.Organization
	nameChanged := org.Name != form.Name

	// Check if organization name has been changed.
	if nameChanged {
		err := org_service.RenameOrganization(ctx, org, form.Name)
		switch {
		case user_model.IsErrUserAlreadyExist(err):
			ctx.Data["OrgName"] = true
			ctx.RenderWithErr(ctx.Tr("form.username_been_taken"), tplSettingsOptions, &form)
			return
		case db.IsErrNameReserved(err):
			ctx.Data["OrgName"] = true
			ctx.RenderWithErr(ctx.Tr("repo.form.name_reserved", err.(db.ErrNameReserved).Name), tplSettingsOptions, &form)
			return
		case db.IsErrNamePatternNotAllowed(err):
			ctx.Data["OrgName"] = true
			ctx.RenderWithErr(ctx.Tr("repo.form.name_pattern_not_allowed", err.(db.ErrNamePatternNotAllowed).Pattern), tplSettingsOptions, &form)
			return
		case err != nil:
			ctx.ServerError("org_service.RenameOrganization", err)
			return
		}

		// reset ctx.org.OrgLink with new name
		ctx.Org.OrgLink = setting.AppSubURL + "/org/" + url.PathEscape(form.Name)
		log.Trace("Organization name changed: %s -> %s", org.Name, form.Name)
	}

	// In case it's just a case change.
	org.Name = form.Name
	org.LowerName = strings.ToLower(form.Name)

	if ctx.Doer.IsAdmin {
		org.MaxRepoCreation = form.MaxRepoCreation
	}

	org.FullName = form.FullName
	org.Email = form.Email
	org.Description = form.Description
	org.Website = form.Website
	org.Location = form.Location
	org.RepoAdminChangeTeamAccess = form.RepoAdminChangeTeamAccess

	visibilityChanged := form.Visibility != org.Visibility
	org.Visibility = form.Visibility

	if err := user_model.UpdateUser(ctx, org.AsUser(), false); err != nil {
		ctx.ServerError("UpdateUser", err)
		return
	}

	// update forks visibility
	if visibilityChanged {
		repos, _, err := repo_model.GetUserRepositories(&repo_model.SearchRepoOptions{
			Actor: org.AsUser(), Private: true, ListOptions: db.ListOptions{Page: 1, PageSize: org.NumRepos},
		})
		if err != nil {
			ctx.ServerError("GetRepositories", err)
			return
		}
		for _, repo := range repos {
			repo.OwnerName = org.Name
			if err := repo_service.UpdateRepository(ctx, repo, true); err != nil {
				ctx.ServerError("UpdateRepository", err)
				return
			}
		}
	}

	log.Trace("Organization setting updated: %s", org.Name)
	ctx.Flash.Success(ctx.Tr("org.settings.update_setting_success"))
	ctx.Redirect(ctx.Org.OrgLink + "/settings")
}

// SettingsAvatar response for change avatar on settings page
func SettingsAvatar(ctx *context.Context) {
	form := web.GetForm(ctx).(*forms.AvatarForm)
	form.Source = forms.AvatarLocal
	if err := user_setting.UpdateAvatarSetting(ctx, form, ctx.Org.Organization.AsUser()); err != nil {
		ctx.Flash.Error(err.Error())
	} else {
		ctx.Flash.Success(ctx.Tr("org.settings.update_avatar_success"))
	}

	ctx.Redirect(ctx.Org.OrgLink + "/settings")
}

// SettingsDeleteAvatar response for delete avatar on settings page
func SettingsDeleteAvatar(ctx *context.Context) {
	if err := user_service.DeleteAvatar(ctx.Org.Organization.AsUser()); err != nil {
		ctx.Flash.Error(err.Error())
	}

	ctx.JSONRedirect(ctx.Org.OrgLink + "/settings")
}

// region SshKey

// SettingsSshKeyList list ssh key for organization
func SettingsSshKeyList(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("org.settings")
	ctx.Data["PageIsOrgSettings"] = true
	ctx.Data["PageIsSettingsSshKey"] = true

	keys, err := organization.GetOrgSshKey(ctx.Org.Organization.ID)

	if err != nil {
		log.Error("error when loading ssh Keys %v", err)
	}

	ctx.Data["SshKeys"] = keys

	ctx.HTML(http.StatusOK, tplSettingsSshKeyList)
}

// SettingsSshKeyCreate handle ssh key for organization
func SettingsSshKeyCreate(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("org.settings")
	ctx.Data["PageIsOrgSettings"] = true
	ctx.Data["PageIsSettingsSshKey"] = true

	var err error

	if ctx.Req.Method == "POST" {
		name := ctx.FormString("key_name")
		publicKey := ctx.FormString("public_key")
		privateKey := ctx.FormString("private_key")

		// TODO: validate ssh key here
		// save to db
		err = organization.AddSshKey(ctx.Org.Organization.ID, name, publicKey, privateKey)
		if err != nil {
			if asymkey.IsErrKeyUnableVerify(err) {
				ctx.Flash.Error("Invalid ssh key")
			} else {
				ctx.Flash.Error("can't save ssh key")
			}
			ctx.Redirect(ctx.Org.OrgLink + "/settings/ssh-key/new")
		}

		ctx.Flash.Success(ctx.Tr("org.settings.ssh_key_added"))
		// Redirect After post request
		ctx.Redirect(ctx.Org.OrgLink + "/settings/ssh-key")
	}
	// Handling for Get
	err = shared_user.LoadHeaderCount(ctx)
	if err != nil {
		ctx.ServerError("LoadHeaderCount", err)
		return
	}

	ctx.HTML(http.StatusOK, tplSettingsSshKeyForm)
}

// SettingsSshKeyDelete handle ssh key for organization
func SettingsSshKeyDelete(ctx *context.Context) {
	id := ctx.FormInt64("id")
	// save to db
	// TODO: check error here
	_ = organization.DeleteOrgSshKey(id)

	ctx.Flash.Warning(ctx.Tr("org.settings.ssh_key_deleted"))
	// Redirect After post request
	ctx.Redirect(ctx.Org.OrgLink + "/settings/ssh-key")
}

// endregion

// region machine

func SettingsMachineList(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("org.settings")
	ctx.Data["PageIsOrgSettings"] = true
	ctx.Data["PageIsSettingsMachine"] = true

	machines, err := organization.GetOrgMachine(ctx.Org.Organization.ID)

	if err != nil {
		ctx.Flash.Error("error on loading machines")
	}

	ctx.Data["Machines"] = machines

	ctx.HTML(http.StatusOK, tplSettingsMachineList)
}

func SettingsMachineCreate(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("org.settings")
	ctx.Data["PageIsOrgSettings"] = true
	ctx.Data["PageIsSettingsMachine"] = true

	if ctx.Req.Method == "POST" {
		form := web.GetForm(ctx).(*forms.SettingMachineForm)
		err := organization.AddMachine(ctx.Org.Organization.ID, form.Name, form.User, form.Host, form.Port, form.SshKey)
		if err != nil {
			ctx.Flash.Error("error on saving machine")
			ctx.Redirect(ctx.Org.OrgLink + "/settings/machine/new")
		}

		ctx.Flash.Success("Machine added successfully")
		ctx.Redirect(ctx.Org.OrgLink + "/settings/machine")
	}
	keys, err := organization.GetOrgSshKey(ctx.Org.Organization.ID)
	if err != nil {
		ctx.Flash.Error("Error on loading ssh Keys")
	}
	ctx.Data["Keys"] = keys
	ctx.HTML(http.StatusOK, tplSettingsMachineForm)
}

func SettingsMachineEdit(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("org.settings")
	ctx.Data["PageIsOrgSettings"] = true
	ctx.Data["PageIsSettingsMachine"] = true

	id := ctx.FormInt64("id")

	machine := &organization.OrgMachine{}
	// check id exist or not
	machine, err := organization.GetMachineById(id, ctx.Org.Organization.ID)
	if err != nil {
		ctx.Flash.Error("error loading machine data")
		ctx.Redirect(ctx.Org.OrgLink + "/settings/machine/edit")
	}

	if ctx.Req.Method == "POST" {
		form := web.GetForm(ctx).(*forms.SettingMachineForm)
		err := organization.UpdateMachine(ctx.Org.Organization.ID, id,
			form.Name, form.User, form.Host, form.Port, form.SshKey)

		if err != nil {
			ctx.Flash.Error("error on saving machine")
			ctx.Redirect(ctx.Org.OrgLink + "/settings/machine/edit?id=" + string(id))
		}

		ctx.Flash.Success("Machine updated successfully")
		ctx.Redirect(ctx.Org.OrgLink + "/settings/machine")
	}

	keys, err := organization.GetOrgSshKey(ctx.Org.Organization.ID)
	if err != nil {
		ctx.Flash.Error("Error on loading ssh Keys")
	}

	ctx.Data["ID"] = id
	ctx.Data["Machine"] = machine
	ctx.Data["Keys"] = keys

	ctx.HTML(http.StatusOK, tplSettingsMachineEdit)
}

func SettingsMachineDelete(ctx *context.Context) {
	id := ctx.FormInt64("id")
	// save to db
	// TODO: check error here
	_ = organization.DeleteMachine(id)

	ctx.Flash.Warning(ctx.Tr("org.settings.machine_deleted"))
	// Redirect After post request
	ctx.Redirect(ctx.Org.OrgLink + "/settings/machine")
}

// endregion

// SettingsDelete response for deleting an organization
func SettingsDelete(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("org.settings")
	ctx.Data["PageIsOrgSettings"] = true
	ctx.Data["PageIsSettingsDelete"] = true

	if ctx.Req.Method == "POST" {
		if ctx.Org.Organization.Name != ctx.FormString("org_name") {
			ctx.Data["Err_OrgName"] = true
			ctx.RenderWithErr(ctx.Tr("form.enterred_invalid_org_name"), tplSettingsDelete, nil)
			return
		}

		if err := org_service.DeleteOrganization(ctx.Org.Organization); err != nil {
			if models.IsErrUserOwnRepos(err) {
				ctx.Flash.Error(ctx.Tr("form.org_still_own_repo"))
				ctx.Redirect(ctx.Org.OrgLink + "/settings/delete")
			} else if models.IsErrUserOwnPackages(err) {
				ctx.Flash.Error(ctx.Tr("form.org_still_own_packages"))
				ctx.Redirect(ctx.Org.OrgLink + "/settings/delete")
			} else {
				ctx.ServerError("DeleteOrganization", err)
			}
		} else {
			log.Trace("Organization deleted: %s", ctx.Org.Organization.Name)
			ctx.Redirect(setting.AppSubURL + "/")
		}
		return
	}

	err := shared_user.LoadHeaderCount(ctx)
	if err != nil {
		ctx.ServerError("LoadHeaderCount", err)
		return
	}

	ctx.HTML(http.StatusOK, tplSettingsDelete)
}

// Webhooks render webhook list page
func Webhooks(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("org.settings")
	ctx.Data["PageIsOrgSettings"] = true
	ctx.Data["PageIsSettingsHooks"] = true
	ctx.Data["BaseLink"] = ctx.Org.OrgLink + "/settings/hooks"
	ctx.Data["BaseLinkNew"] = ctx.Org.OrgLink + "/settings/hooks"
	ctx.Data["Description"] = ctx.Tr("org.settings.hooks_desc")

	ws, err := webhook.ListWebhooksByOpts(ctx, &webhook.ListWebhookOptions{OwnerID: ctx.Org.Organization.ID})
	if err != nil {
		ctx.ServerError("ListWebhooksByOpts", err)
		return
	}

	err = shared_user.LoadHeaderCount(ctx)
	if err != nil {
		ctx.ServerError("LoadHeaderCount", err)
		return
	}

	ctx.Data["Webhooks"] = ws
	ctx.HTML(http.StatusOK, tplSettingsHooks)
}

// DeleteWebhook response for delete webhook
func DeleteWebhook(ctx *context.Context) {
	if err := webhook.DeleteWebhookByOwnerID(ctx.Org.Organization.ID, ctx.FormInt64("id")); err != nil {
		ctx.Flash.Error("DeleteWebhookByOwnerID: " + err.Error())
	} else {
		ctx.Flash.Success(ctx.Tr("repo.settings.webhook_deletion_success"))
	}

	ctx.JSONRedirect(ctx.Org.OrgLink + "/settings/hooks")
}

// Labels render organization labels page
func Labels(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("repo.labels")
	ctx.Data["PageIsOrgSettings"] = true
	ctx.Data["PageIsOrgSettingsLabels"] = true
	ctx.Data["LabelTemplateFiles"] = repo_module.LabelTemplateFiles

	err := shared_user.LoadHeaderCount(ctx)
	if err != nil {
		ctx.ServerError("LoadHeaderCount", err)
		return
	}

	ctx.HTML(http.StatusOK, tplSettingsLabels)
}
