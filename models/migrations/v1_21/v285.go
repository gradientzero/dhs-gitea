package v1_21

import "xorm.io/xorm"

import (
	"code.gitea.io/gitea/models/migrations/base"
)

func RenameDevpodCredentialNameToRemote(x *xorm.Engine) error {

	type OrgDevpodCredential struct {
		ID     int64 `xorm:"pk autoincr"`
		Name   string
		Remote string
	}

	sess := x.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}

	if err := sess.Sync(new(OrgDevpodCredential)); err != nil {
		return err
	}
	if _, err := sess.Exec("UPDATE org_devpod_credential SET remote = name;"); err != nil {
		return err
	}
	if err := sess.Commit(); err != nil {
		return err
	}

	if err := sess.Begin(); err != nil {
		return err
	}
	if err := base.DropTableColumns(sess, "org_devpod_credential", "name"); err != nil {
		return err
	}

	return sess.Commit()
}
