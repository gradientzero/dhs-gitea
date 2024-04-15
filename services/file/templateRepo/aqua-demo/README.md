# Aqua Predict Research

Aqua Predict Research Repo.

AI/ML-based groundwater analysis and prediction solutions.

## Repo Structure

* [data](data) - research and development data sets. Data is managed by DVC
* [code](code) - code repo
* [papers](papers) - scientific papers and other information

## Data Versioning

Data is managed by [DVC](https://dvc.org/doc). Later [DetaBord](https://detabord.com) will offer more advanced data and AI management.
DVC is built on top of git. This means everything is git managed. Use the normal git workflow to use this repository. DVC adds additional features to manage (large) data files.

### DVC Setup

DVC manages data metadate and uses remote data repositories to store the actual data sets. The preferred data storage provider is Exoscale. But this S3 service is not ready yet, in the meantime Azure Blob Storage with a local German zone (west germany) is used.

Environment (python)
```bash
conda create -n dvc python=3.11
conda activate dvc
pip install -r requirements.txt
```

DVC Version (3.4.0)
```bash
# ensure you have installed DVC version 3.4.0 or higher
dvc --version
```

More information for DVC installation:
https://dvc.org/doc/install


#### ExoScale (primary S3 storage)

Follow the installation instructions: https://community.exoscale.com/documentation/storage/quick-start/
```bash
brew install s3cmd
```

Create a config file `~/.s3cfg` with the following content:
```bash
[default]
host_base = sos-de-fra-1.exo.io
host_bucket = %(bucket)s.sos-de-fra-1.exo.io
access_key = $EXO_SOS_KEY
secret_key = $EXO_SOS_SECRET
use_https = True
```
Both `$EXO_SOS_KEY` and `$EXO_SOS_SECRET` you have to request from us once. host_bucket should stay as above.

Ensure you have access to ExoScale:
```bash
s3cmd ls
# 2023-07-03 13:40  s3://aqua01
```

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
Again, both `$EXO_SOS_KEY` and `$EXO_SOS_SECRET` equals to values we already have stored in `~/.s3cfg`

Use `dvc push` and `dvc pull` for data handling.


#### Azure (alternative S3 storage)

Azure accounts are managed by Active Directory. Invites shall be sent via email. Contact jb@gradient0.com for help with the accounts.

Users with access to the aqua01 storage account have the "Storage Blog Data Contributor" role assignment. To access the blog storage setup the connection via the Azure CLI.

Install Azure CLI
[https://learn.microsoft.com/en-us/cli/azure/install-azure-cli?source=recommendations](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli?source=recommendations)

Then, login to your Azure Account
`az login`

And add and config the data remote:

`dvc remote add -d aquaremote azure://aqua01`

`dvc remote modify aquaremote account_name 'aqua01'`

This will use the local Azure CLI config for storage access.

Use `dvc push` and `dvc pull` for data handling. Refer to the DVC docs (see above) for detailed information.