package dvc

import (
	"bufio"
	"code.gitea.io/gitea/modules/context"
	"code.gitea.io/gitea/modules/log"
	repo_module "code.gitea.io/gitea/modules/repository"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type Remote struct {
	Name       string
	Url        string
	AuthorName string
	DateAdded  time.Time
	Link       string
}

type RemoteGitBlame struct {
	AuthorName string
	Date       time.Time
}

func ValidateRemoteName(name string) error {
	if strings.Contains(name, " ") {
		return errors.New("remote name can't have space")
	}
	return nil
}

func RemoteAdd(ctx *context.Context, remote Remote) (err error) {

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

func RemoteList(ctx *context.Context) (remotes []Remote, err error) {

	err = executeTempRepo(ctx, func(tempRepoPath string, repository *git.Repository) error {
		cmd := exec.Command("dvc", "remote", "list")
		cmd.Dir = tempRepoPath
		output, err := cmd.CombinedOutput()

		remotes = ParseRemote(string(output))

		if err != nil {
			log.Error("error when dvc pull: %v", err)
			return err
		}

		blames, err := RemoteCreatedDate(repository)
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

func changeFromProtocolToLink(protocol string) string {
	split := strings.Split(protocol, "://")
	if len(split) >= 2 {
		if split[0] == "gdrive" {
			return strings.Join([]string{"https://drive.google.com/drive/folders", split[1]}, "/")
		}
	}
	return protocol
}

func ParseRemote(output string) (remotes []Remote) {
	re := regexp.MustCompile("\\s")

	sc := bufio.NewScanner(strings.NewReader(output))
	for sc.Scan() {
		split := re.Split(strings.TrimSpace(sc.Text()), -1)
		if len(split) == 2 {
			remotes = append(remotes, Remote{
				Name: split[0],
				Url:  split[1],
				Link: changeFromProtocolToLink(split[1]),
			})
		}
	}
	return remotes
}

// RemoteCreatedDate extract remote created date from git repo
func RemoteCreatedDate(repo *git.Repository) (m map[string]RemoteGitBlame, err error) {

	m = make(map[string]RemoteGitBlame)

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
			m[matches[1]] = RemoteGitBlame{
				AuthorName: v.AuthorName,
				Date:       v.Date,
			}
		}
	}

	return m, err
}

func RemotePull(ctx *context.Context, remote Remote) (output string, err error) {

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

func RemoteDelete(ctx *context.Context, remote Remote) (output string, err error) {

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

	log.Info("git clone %s %s %s", repoPath, branchName, tempRepoPath)
	repo, err := git.PlainClone(tempRepoPath, false, &git.CloneOptions{
		URL: repoPath,
		// Ref: https://github.com/src-d/go-git/issues/553
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branchName)),
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
