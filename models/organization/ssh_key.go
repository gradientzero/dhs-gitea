package organization

import (
	"code.gitea.io/gitea/models/asymkey"
	"code.gitea.io/gitea/models/db"
	"code.gitea.io/gitea/modules/timeutil"
)

type OrgSshKey struct {
	ID          int64              `xorm:"pk autoincr"`
	OwnerID     int64              `xorm:"INDEX NOT NULL"`
	Name        string             `xorm:"NOT NULL"`
	Fingerprint string             `xorm:"INDEX NOT NULL"`
	PublicKey   string             `xorm:"MEDIUMTEXT NOT NULL"`
	PrivateKey  string             `xorm:"MEDIUMTEXT NOT NULL"`
	CreatedUnix timeutil.TimeStamp `xorm:"created"`
	UpdatedUnix timeutil.TimeStamp `xorm:"updated"`
	Verified    bool               `xorm:"NOT NULL DEFAULT false"`
}

// AddSshKey adds ssh key to database
func AddSshKey(ownerID int64, name string, publicKey string, privateKey string) error {

	ctx, committer, err := db.TxContext(db.DefaultContext)
	if err != nil {
		return err
	}

	defer committer.Close()

	fingerPrint, err := asymkey.CalcFingerprint(publicKey)
	if err != nil {
		return err
	}

	// Not have so insert new record here
	key := &OrgSshKey{
		OwnerID:     ownerID,
		Name:        name,
		PublicKey:   publicKey,
		PrivateKey:  privateKey,
		Fingerprint: fingerPrint,
	}

	// Save SSH key.
	if err = db.Insert(ctx, key); err != nil {
		return err
	}

	return committer.Commit()
}

func DeleteOrgSshKey(ID int64) error {
	ctx, committer, err := db.TxContext(db.DefaultContext)
	if err != nil {
		return err
	}

	defer committer.Close()

	_, err = db.GetEngine(ctx).
		ID(ID).
		Delete(&OrgSshKey{})

	if err != nil {
		return err
	}

	return committer.Commit()
}

func GetOrgSshKey(ownerID int64) ([]OrgSshKey, error) {

	var allKeys []OrgSshKey
	err := db.GetEngine(db.DefaultContext).
		Where("owner_id = ? ", ownerID).
		Find(&allKeys)

	if err != nil {
		return nil, err
	}
	return allKeys, nil
}

func GetKeyById(ID int64, ownerID int64) (*OrgSshKey, error) {
	key := new(OrgSshKey)

	_, err := db.GetEngine(db.DefaultContext).
		Where("id = ? AND owner_id = ?", ID, ownerID).
		Get(key)

	if err != nil {
		return nil, err
	}
	return key, nil
}
