package v1_21

import "xorm.io/xorm"

func AddNameToOrgSshKeys(x *xorm.Engine) error {
	type OrgSshKey struct {
		Name        string `xorm:"NOT NULL DEFAULT ''"`
		Fingerprint string `xorm:"INDEX NOT NULL DEFAULT ''"`
	}

	return x.Sync(new(OrgSshKey))
}
