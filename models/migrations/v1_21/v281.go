package v1_21

import (
	"code.gitea.io/gitea/modules/timeutil"
	"xorm.io/xorm"
)

func AddOrgMachineTable(x *xorm.Engine) error {

	type OrgMachine struct {
		ID          int64              `xorm:"pk autoincr"`
		OwnerID     int64              `xorm:"INDEX NOT NULL"`
		Name        string             `xorm:"NOT NULL"`
		User        string             `xorm:"VARCHAR(100)"`
		Host        string             `xorm:"VARCHAR(200)"`
		Port        int32              `xorm:"BIGINT"`
		CreatedUnix timeutil.TimeStamp `xorm:"created"`
		UpdatedUnix timeutil.TimeStamp `xorm:"updated"`
	}

	return x.Sync(
		new(OrgMachine),
	)
}
