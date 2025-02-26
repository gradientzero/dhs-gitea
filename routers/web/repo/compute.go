package repo

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

	org_model "code.gitea.io/gitea/models/organization"
	"code.gitea.io/gitea/models/repo"
	"code.gitea.io/gitea/modules/base"
	"code.gitea.io/gitea/modules/devpod"
	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/services/context"
	"github.com/buildkite/terminal-to-html/v3"
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

	// change with IsBranchExist method to check branch is valid or exists
	branch := ctx.Req.URL.Query().Get("branch")
	tag := ctx.Req.URL.Query().Get("tag")

	if branch != "" && ctx.Repo.GitRepo.IsBranchExist(branch) {
		ctx.Repo.BranchName = branch
	} else if tag != "" && ctx.Repo.GitRepo.IsTagExist(tag) {
		ctx.Repo.TagName = tag
		ctx.Repo.IsViewTag = true
	}

	ctx.Data["BranchName"] = ctx.Repo.BranchName
	ctx.Data["TagName"] = ctx.Repo.TagName
	ctx.Data["IsViewTag"] = ctx.Repo.IsViewTag
	ctx.Data["CanCompute"] = true
	ctx.Data["Machines"] = machines

	// collect all runs from logs
	repoPath := filepath.Join(devpod.RunsFolder, fmt.Sprintf("repo-%d", ctx.Repo.Repository.ID))
	if logFiles, err := devpod.ReadLogFiles(repoPath); err != nil {
		log.Error("Failed to read log files: %v", err)
	} else {
		ctx.Data["Runs"] = logFiles
	}

	ctx.HTML(http.StatusOK, tplCompute)
}

func ComputeExecute(ctx *context.Context) {
	// send stream function
	var sendStream = func(result string) {
		output := terminal.Render([]byte(result))
		ctx.Write([]byte(string(output) + "\n"))
		ctx.Resp.Flush()
	}

	machineId := ctx.Req.URL.Query().Get("machineId")
	parseId, err := strconv.ParseInt(machineId, 10, 64)
	if err != nil {
		log.Error("Invalid machine Id")
	}

	// branch must be valid
	gitBranch := ctx.Req.URL.Query().Get("branch")
	tag := ctx.Req.URL.Query().Get("tag")

	if gitBranch != "" && !ctx.Repo.GitRepo.IsBranchExist(gitBranch) {
		log.Error("Invalid branch")
		return
	} else if tag != "" && !ctx.Repo.GitRepo.IsTagExist(tag) {
		log.Error(("Invalid tag"))
		return
	}

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

	// HTTP(s) is used to clone git repository, not SSH (public key)
	// If repository is private we MUST add a Gitea Token from
	// a user with access to the repository
	gitUrl := cloneLink.HTTPS // gitUrl := cloneLink.SSH

	// Machine
	tokens, err := org_model.GetOrgGiteaToken(ctx.Repo.Owner.ID)
	if err != nil {
		log.Error("%v", err)
	}

	// If repo is private user access token (gitea token) is required
	isRepoPrivate := ctx.Repo.Repository.IsPrivate
	if isRepoPrivate && len(tokens) == 0 {
		err := fmt.Errorf("Private repo requires a Gitea Token")
		log.Error(err.Error())
		sendStream(err.Error())
		return
	}

	var gitToken string
	if len(tokens) > 0 {
		gitToken = tokens[0].Token
	}

	credentials, err := org_model.GetOrgDevpodCredential(ctx.Repo.Owner.ID)
	if err != nil {
		log.Error("%v", err)
	}

	var devPodCredentials = make(map[string][]org_model.OrgDevpodCredential)
	for _, credential := range credentials {
		if v, ok := devPodCredentials[credential.Remote]; ok {
			devPodCredentials[credential.Remote] = append(v, credential)
		} else {
			var lst = []org_model.OrgDevpodCredential{credential}
			devPodCredentials[credential.Remote] = lst
		}
	}

	// add header for event stream
	h := ctx.Resp.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no")

	// send header
	ctx.Resp.Flush()

	gitUser := ctx.Doer.Name
	gitEmail := ctx.Doer.Email

	// execute devpod
	err = devpod.Execute(ctx,
		isRepoPrivate, privateKey,
		user, host, port, gitUrl,
		gitBranch, gitUser, gitEmail,
		gitToken, devPodCredentials, sendStream,
	)
	if err != nil {
		log.Error("%v", err)
	}
}

func DeleteLog(ctx *context.Context) {
	filename := ctx.Req.URL.Query().Get("filename")
	repoPath := filepath.Join(devpod.RunsFolder, fmt.Sprintf("repo-%d", ctx.Repo.Repository.ID))
	runFile := filepath.Join(repoPath, filename)
	if err := devpod.DeleteRunFile(runFile); err != nil {
		log.Error("%v", err)
	}
}
