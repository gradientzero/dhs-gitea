package v1_21

import "xorm.io/xorm"

func AddSshKeyToOrgMachineTable(x *xorm.Engine) error {
	type OrgMachine struct {
		SshKey int64
	}

	return x.Sync(new(OrgMachine))
}
