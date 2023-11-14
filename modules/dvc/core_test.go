package dvc

import (
	"code.gitea.io/gitea/modules/log"
	repo_module "code.gitea.io/gitea/modules/repository"
	"fmt"
	"github.com/dustin/go-humanize"
	"os"
	"regexp"
	"testing"
)

func TestLocalCopy(t *testing.T) {
	basePath, err := repo_module.CreateTemporaryPath("dataset")
	if err != nil {
		log.Error("error: %v", err)
	}

	defer func() {
		if err := repo_module.RemoveTemporaryPath(basePath); err != nil {
			log.Error("Error whilst removing removing temporary repo for %-v: %v", basePath, err)
		}
	}()

	err = os.WriteFile(basePath+"/dat1", []byte("hello world"), 0644)
	fmt.Println(os.TempDir())
	dat, err := os.ReadFile(basePath + "/dat1")
	fmt.Print(string(dat))
	fmt.Println(err)
	fmt.Println(basePath)

}

/*func TestRemoteAdd(t *testing.T) {
	// clone repository
	url := "/home/lin/Desktop/btrs/dhs-gitea/data/gitea-repositories/adminadmin/hello.git"
	err := RemoteAdd(url, "test", "/tmp/data")
	if err != nil {
		return
	}
	fmt.Println("done")
}*/

/*func TestRemotePull(t *testing.T) {
	url := "/home/lin/Desktop/btrs/dhs-gitea/data/gitea-repositories/adminadmin/hello.git"
	output, err := RemotePull(url)
	assert.NoError(t, err)
	fmt.Println(output)
}*/

/*func TestGitBlame(t *testing.T) {
	url := "/home/lin/Desktop/btrs/dhs-gitea/data/gitea-repositories/adminadmin/hello.git"

	executeTempRepo(url, func(tempRepo string, repo *git.Repository) (err error) {

		m, err := RemoteCreatedDate(repo)
		fmt.Println(m)
		return err
	})
}*/

func TestRegexExtract(t *testing.T) {
	data := `['remote "test"']`
	re := regexp.MustCompile(`\['remote "(\w+)"']`)
	matches := re.FindStringSubmatch(data)
	if matches != nil {
		fmt.Println(matches[0])
		fmt.Println(matches[1])
	}
}

func TestJsonParse(t *testing.T) {
	data := `
[
  {
    "isout": true,
    "isdir": false,
    "isexec": false,
    "size": 14445097,
    "path": "data/data.xml"
  },
  {
    "isout": true,
    "isdir": false,
    "isexec": false,
    "size": null,
    "path": "data/features/test.pkl"
  },
  {
    "isout": true,
    "isdir": false,
    "isexec": false,
    "size": null,
    "path": "data/features/train.pkl"
  },
  {
    "isout": true,
    "isdir": false,
    "isexec": false,
    "size": null,
    "path": "data/prepared/test.tsv"
  },
  {
    "isout": true,
    "isdir": false,
    "isexec": false,
    "size": null,
    "path": "data/prepared/train.tsv"
  },
  {
    "isout": true,
    "isdir": false,
    "isexec": false,
    "size": null,
    "path": "eval/metrics.json"
  },
  {
    "isout": true,
    "isdir": false,
    "isexec": false,
    "size": null,
    "path": "eval/plots/images/importance.png"
  },
  {
    "isout": true,
    "isdir": false,
    "isexec": false,
    "size": null,
    "path": "eval/plots/sklearn/cm/test.json"
  },
  {
    "isout": true,
    "isdir": false,
    "isexec": false,
    "size": null,
    "path": "eval/plots/sklearn/cm/train.json"
  },
  {
    "isout": true,
    "isdir": false,
    "isexec": false,
    "size": null,
    "path": "eval/plots/sklearn/prc/test.json"
  },
  {
    "isout": true,
    "isdir": false,
    "isexec": false,
    "size": null,
    "path": "eval/plots/sklearn/prc/train.json"
  },
  {
    "isout": true,
    "isdir": false,
    "isexec": false,
    "size": null,
    "path": "eval/plots/sklearn/roc/test.json"
  },
  {
    "isout": true,
    "isdir": false,
    "isexec": false,
    "size": null,
    "path": "eval/plots/sklearn/roc/train.json"
  },
  {
    "isout": true,
    "isdir": false,
    "isexec": false,
    "size": 1957931,
    "path": "model.pkl"
  }
]
`
	result, _ := ParseJsonFileInfo([]byte(data))
	fmt.Printf("%v\n", result)
}

func TestSplit(t *testing.T) {

	output := `test	/tmp/data
	test-2	/tmp/data-2
	`
	remotes := ParseRemote(output)
	fmt.Println(remotes)

	fmt.Printf("That file is %s.", humanize.Bytes(1957931)) // That file is 83 MB.
}

func TestParseGtoRelease(t *testing.T) {
	output, _ := ParseGtoRelease("/tmp/example-gto")
	fmt.Printf("%v\n", output)
}
/*func TestDvcInit(t *testing.T) {
	url := "/home/lin/Desktop/btrs/dhs-gitea/data/gitea-repositories/adminadmin/second-repo.git"

	executeTempRepo(url, func(tempRepo string, repo *git.Repository) (err error) {

		cmd := exec.Command("dvc", "status")
		cmd.Dir = tempRepo
		out, err := cmd.CombinedOutput()

		output := string(out)
		fmt.Println(output)

		if err != nil {
			if _, err := os.Stat(".dvc/config"); errors.Is(err, os.ErrNotExist) {
				cmd = exec.Command("dvc", "init")
				cmd.Dir = tempRepo
				output, err := cmd.CombinedOutput()
				if err != nil {
					log.Error("dvc init fail")
				}

				log.Info("dvc init is running: %s", output)
			}
		}
		return nil
	})
}*/
