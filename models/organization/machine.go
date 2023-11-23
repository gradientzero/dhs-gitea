package organization

import (
	"code.gitea.io/gitea/models/db"
	"code.gitea.io/gitea/modules/timeutil"
)

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

// AddMachine adds machine to database
func AddMachine(ownerID int64, name, user, host string, port int32) error {

	ctx, committer, err := db.TxContext(db.DefaultContext)
	if err != nil {
		return err
	}

	defer committer.Close()

	// Not have so insert new record here
	key := &OrgMachine{
		OwnerID: ownerID,
		User:    user,
		Name:    name,
		Host:    host,
		Port:    port,
	}

	// Save SSH key.
	if err = db.Insert(ctx, key); err != nil {
		return err
	}

	return committer.Commit()
}

func DeleteMachine(ID int64) error {
	ctx, committer, err := db.TxContext(db.DefaultContext)
	if err != nil {
		return err
	}

	defer committer.Close()

	_, err = db.GetEngine(ctx).
		ID(ID).
		Delete(&OrgMachine{})

	if err != nil {
		return err
	}

	return committer.Commit()
}

func GetOrgMachine(ownerID int64) ([]OrgMachine, error) {

	var allMachines []OrgMachine
	err := db.GetEngine(db.DefaultContext).
		Where("owner_id = ? ", ownerID).
		Find(&allMachines)

	if err != nil {
		return nil, err
	}
	return allMachines, nil
}

func GetMachineById(ID int64, ownerID int64) (*OrgMachine, error) {
	machine := new(OrgMachine)

	_, err := db.GetEngine(db.DefaultContext).
		Where("id = ? AND owner_id = ?", ID, ownerID).
		Get(machine)

	if err != nil {
		return nil, err
	}
	return machine, nil
}
