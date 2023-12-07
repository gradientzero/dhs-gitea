package repo

import (
	org_model "code.gitea.io/gitea/models/organization"
	"code.gitea.io/gitea/models/repo"
	"code.gitea.io/gitea/modules/base"
	"code.gitea.io/gitea/modules/context"
	"code.gitea.io/gitea/modules/devpod"
	"code.gitea.io/gitea/modules/log"
	"github.com/buildkite/terminal-to-html/v3"
	"net/http"
	"strconv"
	"strings"
)

const (
	tplCompute base.TplName = "repo/compute"
)

// MustEnableComputes check if projects are enabled in settings
func MustEnableComputes(ctx *context.Context) {
	// TODO: later enable computes in project settings
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

func Computes(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("repo.models")
	ctx.Data["IsComputePage"] = true // to show highlight in tab

	// check repo owner is organization
	if ctx.Repo.Owner.IsOrganization() == false {
		ctx.Flash.Error("This repo is not owned by organization", true)
		ctx.Data["CanCompute"] = false
		ctx.HTML(http.StatusOK, tplCompute)
		return
	}

	// retrieve repo owner organization
	repoOrg := org_model.OrgFromUser(ctx.Repo.Owner)

	// check user is organization owner
	if ok, err := org_model.IsOrganizationOwner(ctx, repoOrg.ID, ctx.Doer.ID); err != nil {
		ctx.Flash.Error("error on reading organization", true)
		ctx.Data["CanCompute"] = false
		ctx.HTML(http.StatusOK, tplCompute)
		return
	} else if ok == false {
		ctx.Flash.Error("Not organization owner", true)
		ctx.Data["CanCompute"] = false
		ctx.HTML(http.StatusOK, tplCompute)
		return
	}

	machines, err := org_model.GetOrgMachine(repoOrg.ID)
	if err != nil {
		ctx.Flash.Error("error on loading machine list", true)
		ctx.Data["CanCompute"] = false
		ctx.HTML(http.StatusOK, tplCompute)
		return
	}

	ctx.Data["CanCompute"] = true
	ctx.Data["Machines"] = machines

	ctx.HTML(http.StatusOK, tplCompute)
}

func ComputeExecute(ctx *context.Context) {

	// TODO: validate repo and org owner

	machineId := ctx.Req.URL.Query().Get("machineId")
	parseId, err := strconv.ParseInt(machineId, 10, 64)
	if err != nil {
		log.Error("Invalid machine Id")
	}

	// TODO: validate machine id

	// Repo must be organization owned
	orgMachine, err := org_model.GetMachineById(parseId, ctx.Repo.Owner.ID)
	if err != nil {
		log.Error("error on getting machine")
	}

	orgSshKey, err := org_model.GetOrgSshKeyByID(orgMachine.SshKey, ctx.Repo.Owner.ID)
	if err != nil {
		log.Error("error on getting ssh key")
	}

	user := orgMachine.User
	host := orgMachine.Host
	port := orgMachine.Port

	privateKey := orgSshKey.PrivateKey
	cloneLink := ctx.Data["RepoCloneLink"].(*repo.CloneLink)
	gitUrl := cloneLink.HTTPS

	gitUser := ctx.Doer.Name
	tokens, err := org_model.GetOrgGiteaToken(ctx.Repo.Owner.ID)
	if err != nil {
		log.Error("%v", err)
	}
	
	var gitToken string
	if len(tokens) > 0 {
		gitToken = tokens[0].Token
	}

	gitUrl = strings.Replace(gitUrl, "://", "://"+gitUser+":"+gitToken+"@", 1)
	//gitUrl = "http://adminadmin:d718e5fe99591411878cc8b031d5f70c9481871f@gitea.local:3000/org-1/aqua-research.git"

	credentials, err := org_model.GetOrgDevpodCredential(ctx.Repo.Owner.ID)
	if err != nil {
		log.Error("%v", err)
	}

	var config map[string][]org_model.OrgDevpodCredential

	for _, credential := range credentials {
		if v, ok := config["name"]; ok {
			config["name"] = append(v, credential)
		} else {
			config["name"] = []org_model.OrgDevpodCredential{credential}
		}
	}

	result, err := devpod.Execute(privateKey, user, host, port, gitUrl, config)
	if err != nil {
		log.Error("%v", err)
	}

	output := terminal.Render([]byte(result))
	ctx.JSON(http.StatusOK, map[string]any{
		"machineId": machineId,
		"result":    string(output),
	})
}
