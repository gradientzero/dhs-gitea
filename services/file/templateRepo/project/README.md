# Template Repository

Sandbox template Repo with devContainer, DVC, and ExoScale S3 integration ready to use.

## Repo Structure
- .devcontainer - DevContainer setup
- .dvc - DVC configuration
- .vscode - VSCode settings (optional)
- code - Code repo
- data - Data sets. Data is managed by DVC
- dvclive - DVC live experiment results

## Data Versioning
Data is managed by [DVC](https://dvc.org/doc). DVC is a version control system for data sets. It is used to track changes in data sets and to share data sets between team members. DVC is built on top of git. This means everything is git managed. Use the normal git workflow to use this repository. DVC adds additional features to manage (large) data files. With DVC you can easily track your experiments and their progress by only instrumenting your code, and collaborate on ML experiments like software engineers do for code.

## Python Environment
First, a clean python conda environment should be set up:

```bash
conda create -n my-env python=3.13
conda activate my-env
pip install -r requirements.txt
```

## DVC Setup
DVC manages data metadata and connects to remote storage providers (e.g., Exoscale). Ensure you have DVC version 3.4.0+ installed:

```bash
dvc --version
```
More information for DVC installation: https://dvc.org/doc/install

## Setup S3 Connection (ExoScale - primary S3 storage)
Add data remote and use custom ExoScale endpoint:
```bash
dvc remote add -d sandbox s3://sandbox-bucket --force
dvc remote modify sandbox endpointurl https://sos-at-vie-1.exo.io
# This modifies .dvc/config to include remote storage details.
```

DVC requires ExoScale credentials, we will provide them locally only:
```bash
dvc remote modify sandbox --local access_key_id $EXO_SOS_KEY
dvc remote modify sandbox --local secret_access_key $EXO_SOS_SECRET
# This creates .dvc/config.local storing credentials securely (not versioned).
```
Both `$EXO_SOS_KEY` and `$EXO_SOS_SECRET` must be taken from <1password>.

Ensure your dvc config is correct:
```bash
dvc config -l
# remote.sandbox.url=s3://sandbox-bucket
# remote.sandbox.endpointurl=https://sos-at-vie-1.exo.io
# core.remote=sandbox
# remote.sandbox.access_key_id=<...>
# remote.sandbox.secret_access_key=<...>
```

Test remote connection with dvc:
```bash
dvc status --cloud
```

Use `dvc push` and `dvc pull` for data handling.

## Pushing data with DVC
If you want to work with data, please follow the instructions: https://dvc.org/doc/start/data-management/data-versioning
```bash
# just for reference how data/super-secret.txt was added to DVC and uploaded to bucket:
dvc add data/super-secret.txt
git add data/.gitignore data/super-secret.txt.dvc
dvc push
# 1 file pushed
```

## Pulling data with DVC
```bash
dvc pull
# or dvc pull data/super-secret.txt
```

## Run Experiments with DVC
```bash
# requires dvc.yaml to be present
dvc exp run
```

## List all Experiments
```bash
dvc exp show
```

## Push Changes to Repository to Share Results
```bash
git add .
git commit -m "Update experiment results"
git push
```

## Appendix: Create experiment stage (test)
This will generate dvc.yaml file with the stage definition.

```bash
dvc stage add -n simple_run \
  -p simple \
  -d code/simple.py \
  python main.py
```
