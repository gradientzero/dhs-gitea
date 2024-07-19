package repo

import (
	"fmt"
	"net/http"

	"code.gitea.io/gitea/modules/base"
	"code.gitea.io/gitea/modules/dvc"
	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/modules/web"
	"code.gitea.io/gitea/services/context"
	"code.gitea.io/gitea/services/forms"
)

const (
	tplDatasetsNew    base.TplName = "repo/datasets/new"
	tplDatasetsList   base.TplName = "repo/datasets/list"
	tplDatasetsDelete base.TplName = "repo/datasets/delete"
)

// MustEnableDatasets check if projects are enabled in settings
func MustEnableDatasets(ctx *context.Context) {
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

// Datasets show list of dataset in projects
func Datasets(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("repo.dataset")
	//	ctx.Data["CanWriteProjects"] = ctx.Repo.Permission.CanWrite(unit.TypeProjects) @todo check permission handling
	ctx.Data["Link"] = ctx.Repo.Repository.Link() + "/datasets/new"
	ctx.Data["RemoteLink"] = ctx.Repo.Repository.Link() + "/datasets/remote"
	ctx.Data["IsDatasetPage"] = true // to show highlight in tab

	branch := ctx.Req.URL.Query().Get("branch")
	tag := ctx.Req.URL.Query().Get("tag")

	// change with IsBranchExist method to check branch is valid or exists
	if branch != "" && ctx.Repo.GitRepo.IsBranchExist(branch) {
		ctx.Repo.BranchName = branch
	} else if tag != "" && ctx.Repo.GitRepo.IsTagExist(tag) {
		ctx.Repo.TagName = tag
		ctx.Repo.IsViewTag = true
	}

	remotes, err := dvc.RemoteList(ctx)
	if err != nil {
		log.Error("err when remote list: %v", err)
		errorMsg := fmt.Sprintf("error occured when remote list: %v", err)
		ctx.Flash.Error(errorMsg, true)
	}
	// Set branch active or selected branch
	ctx.Data["BranchName"] = ctx.Repo.BranchName
	ctx.Data["TagName"] = ctx.Repo.TagName
	ctx.Data["IsViewTag"] = ctx.Repo.IsViewTag
	ctx.Data["RemoteList"] = remotes

	files, err := dvc.FileList(ctx)
	if err != nil {
		log.Error("err when remote file list: %v", err)
	}
	ctx.Data["Files"] = files

	ctx.HTML(http.StatusOK, tplDatasetsList)
}

func NewDatasetGet(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("repo.datasets.new")
	ctx.Data["Link"] = ctx.Repo.Repository.Link() + "/datasets/new"
	ctx.Data["CancelLink"] = ctx.Repo.Repository.Link() + "/datasets"
	ctx.Data["IsDatasetPage"] = true // to show highlight in tab

	ctx.HTML(http.StatusOK, tplDatasetsNew)
}

// NewDatasetPost creates a new dataset
func NewDatasetPost(ctx *context.Context) {
	form := web.GetForm(ctx).(*forms.CreateDatasetForm)
	ctx.Data["Title"] = ctx.Tr("repo.datasets.new")

	if ctx.HasError() {
		RenderNewDataset(ctx)
		return
	}

	canCommit := renderCommitRights(ctx)
	if !canCommit {
		ctx.Flash.Error("you don't have permission to add")
		ctx.Redirect(ctx.Repo.RepoLink + "/datasets")
		return
	}

	err := dvc.ValidateRemoteName(form.Name)

	if err != nil {
		errorMsg := fmt.Sprintf("error occured when add remote: %v", err)
		ctx.Flash.Error(errorMsg)
		ctx.Redirect(ctx.Repo.RepoLink + "/datasets")
		return
	}

	err = dvc.RemoteAdd(ctx, dvc.Remote{
		Name: form.Name,
		Url:  form.Url,
	})

	if err != nil {
		errorMsg := fmt.Sprintf("error occured when add remote: %v", err)
		ctx.Flash.Error(errorMsg)
		ctx.Redirect(ctx.Repo.RepoLink + "/datasets")
		return
	}

	ctx.Flash.Success(ctx.Tr("repo.datasets.add_success", form.Name))
	ctx.Redirect(ctx.Repo.RepoLink + "/datasets")
}

// RenderNewDataset render creating a dataset page
func RenderNewDataset(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("repo.datasets.new")
	ctx.Data["CancelLink"] = ctx.Repo.Repository.Link() + "/datasets"
	ctx.Data["IsDatasetPage"] = true // to show highlight in tab

	ctx.HTML(http.StatusOK, tplDatasetsList)
}

func SyncDataset(ctx *context.Context) {
	name := ctx.PathParam("name")

	output, err := dvc.RemotePull(ctx, dvc.Remote{
		Name: name,
	})

	if err != nil {
		ctx.Flash.Error("error occurred when dvc remote pull")
	} else {
		ctx.Flash.Success(output)
	}
	// TODO: need to extract remote repo credential to access pull
	//ctx.Flash.Success(ctx.Tr("repo.datasets.add_success", name))
	ctx.Redirect(ctx.Repo.RepoLink + "/datasets")
}

func DeleteDatasetGet(ctx *context.Context) {
	name := ctx.PathParam("name")

	ctx.Data["Title"] = ctx.Tr("repo.dataset")
	ctx.Data["Name"] = name
	ctx.Data["Link"] = ctx.Repo.Repository.Link() + "/datasets/remote/" + name + "/delete"
	ctx.Data["CancelLink"] = ctx.Repo.Repository.Link() + "/datasets"
	ctx.Data["IsDatasetPage"] = true // to show highlight in tab

	ctx.HTML(http.StatusOK, tplDatasetsDelete)
}

func DeleteDatasetPost(ctx *context.Context) {
	name := ctx.PathParam("name")

	output, err := dvc.RemoteDelete(ctx, dvc.Remote{
		Name: name,
	})

	if err != nil {
		ctx.Flash.Error("error occurred when dvc remote delete")
	} else {
		ctx.Flash.Success(output)
	}
	// TODO: need to extract remote repo credential to access pull
	//ctx.Flash.Success(ctx.Tr("repo.datasets.add_success", name))
	ctx.Redirect(ctx.Repo.RepoLink + "/datasets")
}

func findBranches(ctx *context.Context) ([]string, error) {
	var branches []string
	brs, _, err := ctx.Repo.GitRepo.GetBranches(0, 10)
	if err != nil {
		return branches, err
	}

	for _, v := range brs {
		branches = append(branches, v.Name)
	}
	return branches, nil
}
