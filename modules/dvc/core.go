package dvc

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"code.gitea.io/gitea/modules/log"
	repo_module "code.gitea.io/gitea/modules/repository"
	api "code.gitea.io/gitea/modules/structs"
	"code.gitea.io/gitea/services/context"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func ValidateRemoteName(name string) error {
	if strings.Contains(name, " ") {
		return errors.New("remote name can't have space")
	}
	return nil
}

func RemoteAdd(ctx *context.Context, remote api.Remote) (err error) {

	return executeTempRepo(ctx, func(tempRepoPath string, repo *git.Repository) error {

		err = DvcInit(tempRepoPath)
		if err != nil {
			return err
		}

		cmd := exec.Command("dvc", "remote", "add", remote.Name, remote.Url)
		cmd.Dir = tempRepoPath
		_, err := cmd.CombinedOutput()
		if err != nil {
			log.Error("error when dvc remote add: %v", err)
			return err
		}

		w, err := repo.Worktree()
		_, err = w.Add(".dvc/config")

		// These two file need to add, in case new dvc init is run
		_, err = w.Add("./dvc/.gitignore")
		_, err = w.Add(".dvcignore")

		message := fmt.Sprintf("dvc remote %s is added", remote.Name)
		_, err = w.Commit(message, &git.CommitOptions{
			Author: &object.Signature{
				Name:  ctx.Doer.Name,
				Email: ctx.Doer.Email,
				When:  time.Now(),
			},
		})

		// err = r.Push(&git.PushOptions{})
		// Not using r.Push here
		// currently can't modify environment variable from that
		// If so, using exec.Command for git push
		cmd = exec.Command("git", "push", "origin")
		cmd.Dir = tempRepoPath
		cmd.Env = append(os.Environ(), "GITEA_INTERNAL_PUSH=true")
		_, err = cmd.CombinedOutput()

		if err != nil {
			log.Error("error occurred: %v\n", err)
			return err
		}
		return nil
	})
}

func RemoteList(ctx *context.Context) (remotes []api.Remote, err error) {

	err = executeTempRepo(ctx, func(tempRepoPath string, repository *git.Repository) error {
		cmd := exec.Command("dvc", "remote", "list")
		cmd.Dir = tempRepoPath
		output, err := cmd.CombinedOutput()

		remotes = ParseRemote(ctx, string(output))

		if err != nil {
			log.Error("error when dvc pull: %v", err)
			return err
		}

		blames, _ := RemoteCreatedDate(repository)
		for k, v := range remotes {
			remoteGitBlame, ok := blames[v.Name]
			if ok {
				v.AuthorName = remoteGitBlame.AuthorName
				v.DateAdded = remoteGitBlame.Date
				remotes[k] = v
			}
		}
		return nil
	})

	return remotes, err
}

func changeFromProtocolToLink(protocol string, endpointurl string) string {
	protocolPath := strings.Split(protocol, "://")

	if len(protocolPath) == 2 {
		if protocolPath[0] == "s3" {
			path := strings.Split(protocolPath[1], "/")
			if len(path) == 2 {
				return fmt.Sprintf("https://s3.amazonaws.com/%s/%s", path[0], path[1])
			} else if endpointurl != "" {
				endpointurl = strings.TrimSuffix(endpointurl, "\n")
				return fmt.Sprintf("%s/%s", endpointurl, path[0])
			}
		}
		if protocolPath[0] == "gs" {
			path := strings.Split(protocolPath[1], "/")
			if len(path) == 2 {
				return fmt.Sprintf("https://storage.googleapis.com/%s/%s", path[0], path[1])
			}
		}
		if protocolPath[0] == "gdrive" {
			return fmt.Sprintf("https://drive.google.com/drive/folders/%s", protocolPath[1])
		}
		if protocolPath[0] == "oss" {
			path := strings.Split(protocolPath[1], "/")
			if len(path) == 2 {
				return fmt.Sprintf("https://oss.aliyuncs.com/%s/%s", path[0], path[1])
			}
		}
		if protocolPath[0] == "webhdfs" {
			path := strings.Split(protocolPath[1], "/")
			host := strings.Split(path[0], "@")
			if len(path) == 2 && len(host) == 2 {
				return fmt.Sprintf("http://%s/webhdfs/v1/%s", host[1], path[1])
			}
		}
	}
	return protocol
}

func ParseRemote(ctx *context.Context, output string) (remotes []api.Remote) {
	re := regexp.MustCompile("\\s")

	sc := bufio.NewScanner(strings.NewReader(output))
	for sc.Scan() {
		split := re.Split(strings.TrimSpace(sc.Text()), -1)

		// get endpointurl value
		endpointurl := ""
		executeTempRepo(ctx, func(tempRepoPath string, repository *git.Repository) error {
			str := fmt.Sprintf("remote.%s.endpointurl", split[0])
			cmd := exec.Command("dvc", "config", str)
			cmd.Dir = tempRepoPath
			out, err := cmd.CombinedOutput()

			if err != nil {
				log.Error("error when dvc pull: %v", err)
				return err
			}

			endpointurl = string(out)
			return nil
		})

		if len(split) == 2 {
			remotes = append(remotes, api.Remote{
				Name: split[0],
				Url:  split[1],
				Link: changeFromProtocolToLink(split[1], endpointurl),
			})
		}
	}
	return remotes
}

// RemoteCreatedDate extract remote created date from git repo
func RemoteCreatedDate(repo *git.Repository) (m map[string]api.RemoteGitBlame, err error) {

	m = make(map[string]api.RemoteGitBlame)

	// Regex To match `['remote "test"']` string and extract `test` value back
	re := regexp.MustCompile(`\['remote "([\w\-]+)"']`)

	head, err := repo.Head()
	if err != nil {
		return m, err
	}

	c, err := repo.CommitObject(head.Hash())
	if err != nil {
		return m, err
	}

	br, err := git.Blame(c, ".dvc/config")
	if err != nil {
		return m, err
	}

	for _, v := range br.Lines {
		matches := re.FindStringSubmatch(v.Text)
		if matches != nil {
			m[matches[1]] = api.RemoteGitBlame{
				AuthorName: v.AuthorName,
				Date:       v.Date,
			}
		}
	}

	return m, err
}

func RemotePull(ctx *context.Context, remote api.Remote) (output string, err error) {

	err = executeTempRepo(ctx, func(tempRepoPath string, repository *git.Repository) error {
		cmd := exec.Command("dvc", "pull", "--remote", remote.Name)
		cmd.Dir = tempRepoPath
		out, err := cmd.CombinedOutput()

		if err != nil {
			log.Error("error when dvc pull: %v", err)
			return err
		}

		output = string(out)
		return nil
	})

	return output, err
}

func RemoteDelete(ctx *context.Context, remote api.Remote) (output string, err error) {

	err = executeTempRepo(ctx, func(tempRepoPath string, repo *git.Repository) error {
		cmd := exec.Command("dvc", "remote", "remove", remote.Name)
		cmd.Dir = tempRepoPath
		out, err := cmd.CombinedOutput()

		if err != nil {
			log.Error("error when dvc remove: %v", err)
			return err
		}

		output = string(out)

		w, err := repo.Worktree()
		_, err = w.Add(".dvc/config")

		message := fmt.Sprintf("dvc remote %s removed", remote.Name)
		_, err = w.Commit(message, &git.CommitOptions{
			Author: &object.Signature{
				Name:  ctx.Doer.Name,
				Email: ctx.Doer.Email,
				When:  time.Now(),
			},
		})

		// err = r.Push(&git.PushOptions{})
		// Not using r.Push here
		// currently can't modify environment variable from that
		// If so, using exec.Command for git push
		cmd = exec.Command("git", "push", "origin")
		cmd.Dir = tempRepoPath
		cmd.Env = append(os.Environ(), "GITEA_INTERNAL_PUSH=true")
		_, err = cmd.CombinedOutput()

		if err != nil {
			log.Error("error occurred: %v\n", err)
			return err
		}

		return nil
	})

	return output, err
}

// execute code withing given repo,
// execute function accept tempRepoPath, and gitRepository
func executeTempRepo(ctx *context.Context, execute func(string, *git.Repository) error) error {

	repoPath := ctx.Repo.GitRepo.Path
	branchName := ctx.Repo.BranchName
	tagName := ctx.Repo.TagName

	refName := ""
	prefix := ""
	if tagName != "" {
		refName = tagName
		prefix = "refs/tags"
	} else {
		refName = branchName
		prefix = "refs/heads"
	}

	tempRepoPath, err := repo_module.CreateTemporaryPath("dataset")
	if err != nil {
		log.Error("error: %v", err)
		return err
	}

	defer func() {
		if err := repo_module.RemoveTemporaryPath(tempRepoPath); err != nil {
			log.Error("Error whilst removing removing temporary repo for %-v: %v", tempRepoPath, err)
		}
	}()

	log.Info("git clone %s %s %s", repoPath, refName, tempRepoPath)
	repo, err := git.PlainClone(tempRepoPath, false, &git.CloneOptions{
		URL: repoPath,
		// Ref: https://github.com/src-d/go-git/issues/553
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("%s/%s", prefix, refName)),
		SingleBranch:  true,
	})

	if err != nil {
		log.Error("error when git clone %v", err)
		return err
	}

	return execute(tempRepoPath, repo)
}

func DvcInit(tempRepo string) error {
	cmd := exec.Command("dvc", "status")
	cmd.Dir = tempRepo
	out, err := cmd.CombinedOutput()

	output := string(out)
	log.Info("dvc status output: %v", output)
	if err == nil {
		return nil // dvc init already
	}
	// if dvc status not success, check .dvc/config exist or not
	// try to run dvc init
	if _, err := os.Stat(".dvc/config"); errors.Is(err, os.ErrNotExist) {
		cmd = exec.Command("dvc", "init")
		cmd.Dir = tempRepo
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Error("dvc init fail")
			return err
		}

		log.Info("dvc init is running: %s", output)
	}

	return nil
}
