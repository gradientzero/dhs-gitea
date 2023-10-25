package repo

import (
	"code.gitea.io/gitea/modules/base"
	"code.gitea.io/gitea/modules/charset"
	"code.gitea.io/gitea/modules/context"
	"code.gitea.io/gitea/modules/dvc"
	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/modules/setting"
	"code.gitea.io/gitea/modules/typesniffer"
	"code.gitea.io/gitea/modules/util"
	"code.gitea.io/gitea/modules/web"
	"code.gitea.io/gitea/services/forms"
	files_service "code.gitea.io/gitea/services/repository/files"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
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

	branches, err := findBranches(ctx)

	branch := ctx.Req.URL.Query().Get("branch")
	if branch != "" && slices.Contains(branches, branch) {
		ctx.Repo.BranchName = branch
	} else {
		branch = ctx.Repo.BranchName // reset default branch name
	}

	remotes, err := dvc.RemoteList(ctx)
	if err != nil {
		log.Error("err when remote list: %v", err)
		errorMsg := fmt.Sprintf("error occured when remote list: %v", err)
		ctx.Data["Message"] = errorMsg
	}
	ctx.Data["Branches"] = branches
	ctx.Data["Branch"] = branch
	ctx.Data["RemoteList"] = remotes

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

func writeToRepo(ctx *context.Context, filename string, text string, commitMessage string) error {
	if _, err := files_service.ChangeRepoFiles(ctx, ctx.Repo.Repository, ctx.Doer, &files_service.ChangeRepoFilesOptions{
		LastCommitID: ctx.Repo.CommitID,
		OldBranch:    ctx.Repo.BranchName,
		//NewBranch:    GetUniquePatchBranchName(ctx),
		Message: commitMessage,
		Files: []*files_service.ChangeRepoFile{
			{
				Operation:     "create",
				FromTreePath:  ctx.Repo.TreePath,
				TreePath:      filename,
				ContentReader: strings.NewReader(text),
			},
		},
		Signoff: true,
	}); err != nil {
		return err
	}
	return nil
}

// RenderNewDataset render creating a dataset page
func RenderNewDataset(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("repo.datasets.new")
	ctx.Data["CancelLink"] = ctx.Repo.Repository.Link() + "/datasets"
	ctx.Data["IsDatasetPage"] = true // to show highlight in tab

	ctx.HTML(http.StatusOK, tplDatasetsList)
}

// reading text from file inside repository
func readFromRepo(ctx *context.Context, filename string) (string, error) {
	entry, err := ctx.Repo.Commit.GetTreeEntryByPath(filename)
	if err != nil {
		return "", err
	}

	blob := entry.Blob()
	if blob.Size() >= setting.UI.MaxDisplayFileSize {
		return "", err
	}

	dataRc, err := blob.DataAsync()
	if err != nil {
		return "", err
	}

	defer dataRc.Close()

	buf := make([]byte, 1024)
	n, _ := util.ReadAtMost(dataRc, buf)
	buf = buf[:n]

	// Only some file types are editable online as text.
	if !typesniffer.DetectContentType(buf).IsRepresentableAsText() {
		return "", errors.New("this is not text")
	}

	d, _ := io.ReadAll(dataRc)
	if err := dataRc.Close(); err != nil {
		return "", err
	}

	buf = append(buf, d...)
	if content, err := charset.ToUTF8WithErr(buf); err != nil {
		return string(buf), err
	} else {
		return content, nil
	}
}

func SyncDataset(ctx *context.Context) {
	name := ctx.Params("name")

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
	name := ctx.Params("name")

	ctx.Data["Title"] = ctx.Tr("repo.dataset")
	ctx.Data["Name"] = name
	ctx.Data["Link"] = ctx.Repo.Repository.Link() + "/datasets/remote/" + name + "/delete"
	ctx.Data["CancelLink"] = ctx.Repo.Repository.Link() + "/datasets"
	ctx.Data["IsDatasetPage"] = true // to show highlight in tab

	ctx.HTML(http.StatusOK, tplDatasetsDelete)
}

func DeleteDatasetPost(ctx *context.Context) {
	name := ctx.Params("name")

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
