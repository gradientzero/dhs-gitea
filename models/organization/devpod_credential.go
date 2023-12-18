package organization

import (
	"code.gitea.io/gitea/models/db"
	"code.gitea.io/gitea/modules/timeutil"
)

type OrgDevpodCredential struct {
	ID          int64              `xorm:"pk autoincr"`
	OwnerID     int64              `xorm:"INDEX NOT NULL"`
	Remote      string             `xorm:"NOT NULL"`
	Key         string             `xorm:"NOT NULL"`
	Value       string             `xorm:"NOT NULL"`
	CreatedUnix timeutil.TimeStamp `xorm:"created"`
	UpdatedUnix timeutil.TimeStamp `xorm:"updated"`
}

func AddDevpodCredential(ownerID int64, remote, key, value string) error {

	ctx, committer, err := db.TxContext(db.DefaultContext)
	if err != nil {
		return err
	}

	defer committer.Close()

	devpodCredential := &OrgDevpodCredential{
		OwnerID: ownerID,
		Remote:  remote,
		Key:     key,
		Value:   value,
	}

	// Save SSH key.
	if err = db.Insert(ctx, devpodCredential); err != nil {
		return err
	}

	return committer.Commit()
}

func UpdateDevpodCredential(ownerID int64, ID int64,
	remote, key, value string) error {

	ctx, committer, err := db.TxContext(db.DefaultContext)
	if err != nil {
		return err
	}

	defer committer.Close()

	credential := &OrgDevpodCredential{
		Remote: remote,
		Key:    key,
		Value:  value,
	}

	_, err = db.GetEngine(ctx).ID(ID).
		Cols("remote", "key", "value").
		Update(credential)

	return committer.Commit()
}

func GetDevpodCredentialById(ID int64, ownerID int64) (*OrgDevpodCredential, error) {
	credential := new(OrgDevpodCredential)

	_, err := db.GetEngine(db.DefaultContext).
		Where("id = ? AND owner_id = ?", ID, ownerID).
		Get(credential)

	if err != nil {
		return nil, err
	}
	return credential, nil
}

func DeleteOrgDevpodCredential(ID int64) error {
	ctx, committer, err := db.TxContext(db.DefaultContext)
	if err != nil {
		return err
	}

	defer committer.Close()

	_, err = db.GetEngine(ctx).
		ID(ID).
		Delete(&OrgDevpodCredential{})

	if err != nil {
		return err
	}

	return committer.Commit()
}

func GetOrgDevpodCredential(ownerID int64) ([]OrgDevpodCredential, error) {

	var allCredentials []OrgDevpodCredential
	err := db.GetEngine(db.DefaultContext).
		Where("owner_id = ? ", ownerID).
		Find(&allCredentials)

	if err != nil {
		return nil, err
	}
	return allCredentials, nil
}
