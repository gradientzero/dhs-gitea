package organization

import (
	"code.gitea.io/gitea/models/db"
	"code.gitea.io/gitea/modules/timeutil"
)

type OrgGiteaToken struct {
	ID          int64              `xorm:"pk autoincr"`
	OwnerID     int64              `xorm:"INDEX NOT NULL"`
	Name        string             `xorm:"NOT NULL"`
	Token       string             `xorm:"NOT NULL"`
	CreatedUnix timeutil.TimeStamp `xorm:"created"`
	UpdatedUnix timeutil.TimeStamp `xorm:"updated"`
	Verified    bool               `xorm:"NOT NULL DEFAULT false"`
}

func AddGiteaToken(ownerID int64, name, token string) error {

	ctx, committer, err := db.TxContext(db.DefaultContext)
	if err != nil {
		return err
	}

	defer committer.Close()

	giteaToken := &OrgGiteaToken{
		OwnerID: ownerID,
		Name:    name,
		Token:   token,
	}

	// Save SSH key.
	if err = db.Insert(ctx, giteaToken); err != nil {
		return err
	}

	return committer.Commit()
}

func DeleteOrgGiteaToken(ID int64) error {
	ctx, committer, err := db.TxContext(db.DefaultContext)
	if err != nil {
		return err
	}

	defer committer.Close()

	_, err = db.GetEngine(ctx).
		ID(ID).
		Delete(&OrgGiteaToken{})

	if err != nil {
		return err
	}

	return committer.Commit()
}

func GetOrgGiteaToken(ownerID int64) ([]OrgGiteaToken, error) {

	var allKeys []OrgGiteaToken
	err := db.GetEngine(db.DefaultContext).
		Where("owner_id = ? ", ownerID).
		Find(&allKeys)

	if err != nil {
		return nil, err
	}
	return allKeys, nil
}
