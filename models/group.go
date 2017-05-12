package models

import (
	"errors"
	"net/mail"
	"time"

	"github.com/jinzhu/gorm"
)

// Group contains the fields needed for a user -> group mapping
// Groups contain 1..* Targets
type Group struct {
	Id           int64     `json:"id"`
	UserId       int64     `json:"-"`
	Name         string    `json:"name"`
	ModifiedDate time.Time `json:"modified_date"`
	Targets      []Target  `json:"targets" sql:"-"`
}

// GroupTarget is used for a many-to-many relationship between 1..* Groups and 1..* Targets
type GroupTarget struct {
	GroupId  int64 `json:"-"`
	TargetId int64 `json:"-"`
}

// Target contains the fields needed for individual targets specified by the user
// Groups contain 1..* Targets, but 1 Target may belong to 1..* Groups
type Target struct {
	Id         int64  `json:"-"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Email      string `json:"email"`
	Position   string `json:"position"`
	Department string `json:"department"`
}

// ErrNoEmailSpecified is thrown when no email is specified for the Target
var ErrEmailNotSpecified = errors.New("No email address specified")

// ErrGroupNameNotSpecified is thrown when a group name is not specified
var ErrGroupNameNotSpecified = errors.New("Group name not specified")

// ErrNoTargetsSpecified is thrown when no targets are specified by the user
var ErrNoTargetsSpecified = errors.New("No targets specified")

// Validate performs validation on a group given by the user
func (g *Group) Validate() error {
	switch {
	case g.Name == "":
		return ErrGroupNameNotSpecified
	case len(g.Targets) == 0:
		return ErrNoTargetsSpecified
	}
	return nil
}

// GetGroups returns the groups owned by the given user.
func GetGroups(uid int64) ([]Group, error) {
	gs := []Group{}
	err := db.Where("user_id=?", uid).Find(&gs).Error
	if err != nil {
		Logger.Println(err)
		return gs, err
	}
	for i, _ := range gs {
		gs[i].Targets, err = GetTargets(gs[i].Id)
		if err != nil {
			Logger.Println(err)
		}
	}
	return gs, nil
}

// GetGroup returns the group, if it exists, specified by the given id and user_id.
func GetGroup(id int64, uid int64) (Group, error) {
	g := Group{}
	err := db.Where("user_id=? and id=?", uid, id).Find(&g).Error
	if err != nil {
		Logger.Println(err)
		return g, err
	}
	g.Targets, err = GetTargets(g.Id)
	if err != nil {
		Logger.Println(err)
	}
	return g, nil
}

// GetGroupByName returns the group, if it exists, specified by the given name and user_id.
func GetGroupByName(n string, uid int64) (Group, error) {
	g := Group{}
	err := db.Where("user_id=? and name=?", uid, n).Find(&g).Error
	if err != nil {
		Logger.Println(err)
		return g, err
	}
	g.Targets, err = GetTargets(g.Id)
	if err != nil {
		Logger.Println(err)
	}
	return g, err
}

// PostGroup creates a new group in the database.
func PostGroup(g *Group) error {
	if err := g.Validate(); err != nil {
		return err
	}
	// Insert the group into the DB
	err = db.Save(g).Error
	if err != nil {
		Logger.Println(err)
		return err
	}
	for _, t := range g.Targets {
		insertTargetIntoGroup(t, g.Id)
	}
	return nil
}

// PutGroup updates the given group if found in the database.
func PutGroup(g *Group) error {
	if err := g.Validate(); err != nil {
		return err
	}
	// Fetch group's existing targets from database.
	ts := []Target{}
	ts, err = GetTargets(g.Id)
	if err != nil {
		Logger.Printf("Error getting targets from group ID: %d", g.Id)
		return err
	}
	// Check existing targets, removing any that are no longer in the group.
	tExists := false
	for _, t := range ts {
		tExists = false
		// Is the target still in the group?
		for _, nt := range g.Targets {
			if t.Email == nt.Email {
				tExists = true
				break
			}
		}
		// If the target does not exist in the group any longer, we delete it
		if !tExists {
			err = db.Where("group_id=? and target_id=?", g.Id, t.Id).Delete(&GroupTarget{}).Error
			if err != nil {
				Logger.Printf("Error deleting email %s\n", t.Email)
			}
		}
	}
	// Add any targets that are not in the database yet.
	for _, nt := range g.Targets {
		// Check and see if the target already exists in the db
		tExists = false
		for _, t := range ts {
			if t.Email == nt.Email {
				tExists = true
				nt.Id = t.Id
				break
			}
		}
		// Add target if not in database, otherwise update target information.
		if !tExists {
			insertTargetIntoGroup(nt, g.Id)
		} else {
			UpdateTarget(nt)
		}
	}
	err = db.Save(g).Error
	if err != nil {
		Logger.Println(err)
		return err
	}
	return nil
}

// DeleteGroup deletes a given group by group ID and user ID
func DeleteGroup(g *Group) error {
	// Delete all the group_targets entries for this group
	err := db.Where("group_id=?", g.Id).Delete(&GroupTarget{}).Error
	if err != nil {
		Logger.Println(err)
		return err
	}
	// Delete the group itself
	err = db.Delete(g).Error
	if err != nil {
		Logger.Println(err)
		return err
	}
	return err
}

func insertTargetIntoGroup(t Target, gid int64) error {
	if _, err = mail.ParseAddress(t.Email); err != nil {
		Logger.Printf("Invalid email %s\n", t.Email)
		return err
	}
	trans := db.Begin()
	trans.Where(t).FirstOrCreate(&t)
	if err != nil {
		Logger.Printf("Error adding target: %s\n", t.Email)
		return err
	}
	err = trans.Where("group_id=? and target_id=?", gid, t.Id).Find(&GroupTarget{}).Error
	if err == gorm.ErrRecordNotFound {
		err = trans.Save(&GroupTarget{GroupId: gid, TargetId: t.Id}).Error
		if err != nil {
			Logger.Println(err)
			return err
		}
	}
	if err != nil {
		Logger.Printf("Error adding many-many mapping for %s\n", t.Email)
		return err
	}
	err = trans.Commit().Error
	if err != nil {
		Logger.Printf("Error committing db changes\n")
		return err
	}
	return nil
}

// UpdateTarget updates the given target information in the database.
func UpdateTarget(target Target) error {
	targetInfo := map[string]interface{}{
		"first_name": target.FirstName,
		"last_name":  target.LastName,
		"position":   target.Position,
		"department": target.Department,
	}
	err := db.Model(&target).Where("id = ?", target.Id).Updates(targetInfo).Error
	if err != nil {
		Logger.Printf("Error updating target information for %s\n", target.Email)
	}
	return err
}

// GetTargets performs a many-to-many select to get all the Targets for a Group
func GetTargets(gid int64) ([]Target, error) {
	ts := []Target{}
	err := db.Table("targets").Select("targets.id, targets.email, targets.first_name, targets.last_name, targets.position").Joins("left join group_targets gt ON targets.id = gt.target_id").Where("gt.group_id=?", gid).Scan(&ts).Error
	return ts, err
}
