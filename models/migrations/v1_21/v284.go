package v1_21

import (
	"code.gitea.io/gitea/modules/timeutil"
	"xorm.io/xorm"
)

func AddOrgDevpodCredentialTable(x *xorm.Engine) error {

	type OrgDevpodCredential struct {
		ID          int64              `xorm:"pk autoincr"`
		OwnerID     int64              `xorm:"INDEX NOT NULL"`
		Name        string             `xorm:"NOT NULL"`
		Key         string             `xorm:"NOT NULL"`
		Value       string             `xorm:"NOT NULL"`
		CreatedUnix timeutil.TimeStamp `xorm:"created"`
		UpdatedUnix timeutil.TimeStamp `xorm:"updated"`
	}

	return x.Sync(
		new(OrgDevpodCredential),
	)
}
