# Sandbox Template Repository

## Data Versioning

Data is managed by [DVC](https://dvc.org/doc).
DVC is built on top of git. This means everything is git managed. Use the normal git workflow to use this repository. DVC adds additional features to manage (large) data files.

# Python Environment

Environment (python)
```bash
conda create -n dvc python=3.11
conda activate dvc
pip install -r requirements.txt
```

### DVC Setup

DVC manages data metadata and uses remote data repositories to store the actual data sets. The preferred data storage provider is Exoscale.

DVC Version (3.4.0)
```bash
# ensure you have installed DVC version 3.4.0 or higher
dvc --version
```

More information for DVC installation:
https://dvc.org/doc/install


### Setup S3 Connection (ExoScale - primary S3 storage)

Add data remote and use custom ExoScale endpoint:
```bash
dvc remote add -d aquaremote s3://aqua01 --force
dvc remote modify aquaremote endpointurl https://sos-de-fra-1.exo.io
# this will modify the file ".dvc/config"
```

DVC requires ExoScale credentials, we will provide them locally only:
```bash
dvc remote modify aquaremote --local access_key_id $EXO_SOS_KEY
dvc remote modify aquaremote --local secret_access_key $EXO_SOS_SECRET
# this will create a new file "config.local" that contains credentials for using ExoScale
```
Both `$EXO_SOS_KEY` and `$EXO_SOS_SECRET` must be taken from <1password>.


Ensure your dvc config is correct:
```bash
dvc config -l
# remote.aquaremote.url=s3://aqua01
# remote.aquaremote.endpointurl=https://sos-de-fra-1.exo.io
# core.remote=aquaremote
# remote.aquaremote.access_key_id=<...>
# remote.aquaremote.secret_access_key=<...>
```

Check remote connection with dvc:

```bash
dvc status --cloud
```

Use `dvc push` and `dvc pull` for data handling.

# Pushing data with DVC
If you want to work with data, please follow the instructions: https://dvc.org/doc/start/data-management/data-versioning
```bash
# just for reference how data/super-secret.txt was added to DVC and uploaded to bucket:
dvc add data/super-secret.txt
git add data/.gitignore data/super-secret.txt.dvc
dvc push
# 1 file pushed
```

# Pulling data with DVC
```bash
dvc pull data
dvc pull data-any-other
# or
dvc pull
```

# Appendix: Create experiment stage (test)
```bash
# create
dvc stage add -n simple_run \
  -p simple \
  -d code/simple.py \
  python main.py
```
