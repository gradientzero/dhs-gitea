package v1_21

import (
	"code.gitea.io/gitea/modules/timeutil"
	"xorm.io/xorm"
)

func AddOrgSshKeyTable(x *xorm.Engine) error {
	type OrgSshKey struct {
		ID          int64              `xorm:"pk autoincr"`
		OwnerID     int64              `xorm:"INDEX NOT NULL"`
		PublicKey   string             `xorm:"MEDIUMTEXT NOT NULL"`
		PrivateKey  string             `xorm:"MEDIUMTEXT NOT NULL"`
		CreatedUnix timeutil.TimeStamp `xorm:"created"`
		UpdatedUnix timeutil.TimeStamp `xorm:"updated"`
		Verified    bool               `xorm:"NOT NULL DEFAULT false"`
	}

	return x.Sync(
		new(OrgSshKey),
	)
}
