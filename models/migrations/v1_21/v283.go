package v1_21

import (
	"code.gitea.io/gitea/modules/timeutil"
	"xorm.io/xorm"
)

func AddOrgGiteaTokenTable(x *xorm.Engine) error {

	type OrgGiteaToken struct {
		ID          int64              `xorm:"pk autoincr"`
		OwnerID     int64              `xorm:"INDEX NOT NULL"`
		Name        string             `xorm:"NOT NULL"`
		Token       string             `xorm:"NOT NULL"`
		CreatedUnix timeutil.TimeStamp `xorm:"created"`
		UpdatedUnix timeutil.TimeStamp `xorm:"updated"`
		Verified    bool               `xorm:"NOT NULL DEFAULT false"`
	}

	return x.Sync(
		new(OrgGiteaToken),
	)
}
