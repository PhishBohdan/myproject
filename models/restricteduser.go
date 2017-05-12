package models

import (
	"strings"

	"github.com/jinzhu/gorm"
)

// RestrictedUser represents the restricteduser model for gophish.
type RestrictedUser struct {
	Id             int64                `json:"id"`
	Username       string               `json:"username" sql:"not null;unique"`
	ParentUsername string               `json:"parentusername" sql:"not null"`
	Campaigns      []RestrictedCampaign `json:"campaigns" sql:"not null"`
	Hash           string               `json:"-"`
}

type RestrictedCampaign struct {
	ID               int
	RestrictedUserID int   `gorm:"index"` // Foreign key (belongs to), tag `index` will create index for this column
	CampaignId       int64 `json:"campaignid" sql:"not null"`
}

func (r *RestrictedUser) GetParentUser() (User, error) {
	return GetUserByUsername(r.ParentUsername)
}

func GetEmailDomainName(email string) string {
	arr := strings.Split(email, "@")
	if len(arr) <= 1 {
		return ""
	}
	return arr[1]
}

func FuzzyMatch(campaignName, userDomainName string) bool {
	parts := strings.Split(userDomainName, ".")
	part1 := parts[0]
	part1 = strings.ToLower(part1)
	campaignName = strings.ToLower(campaignName)
	return strings.Contains(campaignName, part1)
}

func (r *RestrictedUser) AllowCampaign(c Campaign) bool {
	for _, rc := range r.Campaigns {
		if rc.CampaignId == c.Id {
			return true
		}
	}
	userDomainName := GetEmailDomainName(r.Username)
	if userDomainName == "" {
		return false
	}
	for _, r := range c.Results {
		if GetEmailDomainName(r.Email) == userDomainName {
			return true
		}
	}
	return FuzzyMatch(c.Name, userDomainName)
}

func (r *RestrictedUser) AllowCampaignResults(cr CampaignResults) bool {
	for _, rc := range r.Campaigns {
		if rc.CampaignId == cr.Id {
			return true
		}
	}
	userDomainName := GetEmailDomainName(r.Username)
	if userDomainName == "" {
		return false
	}
	for _, r := range cr.Results {
		if GetEmailDomainName(r.Email) == userDomainName {
			return true
		}
	}
	return FuzzyMatch(cr.Name, userDomainName)
}

func (r *RestrictedUser) FilterAllowedCampaigns(campaigns []Campaign) []Campaign {
	restrictedCampaigns := []Campaign{}
	for _, c := range campaigns {
		if r.AllowCampaign(c) {
			restrictedCampaigns = append(restrictedCampaigns, c)
		}
	}
	return restrictedCampaigns
}

// GetRestrictedUser returns the user that the given id corresponds to. If no user is found, an
// error is thrown.
func GetRestrictedUser(id int64) (RestrictedUser, error) {
	u := RestrictedUser{}
	err := db.Where("id=?", id).First(&u).Error
	if err != nil {
		return u, err
	}
	var campaigns []RestrictedCampaign
	db.Model(&u).Related(&campaigns)
	u.Campaigns = campaigns
	return u, nil
}

// GetRestrictedUserByUsername returns the restricted user that the given username corresponds to. If no restricted user is found, an
// error is thrown.
func GetRestrictedUserByUsername(username string) (RestrictedUser, error) {
	u := RestrictedUser{}
	err := db.Where("username = ?", username).First(&u).Error
	// No issue if we don't find a record
	if err == gorm.ErrRecordNotFound {
		return u, nil
	} else if err == nil {
		u, e := GetRestrictedUser(u.Id)
		if e != nil {
			return u, e
		} else {
			return u, ErrUsernameTaken
		}
	}
	return u, err
}

// GetRestrictedUsersByParent returns the restricted users with parent username.
func GetRestrictedUsersByParent(parent string) []RestrictedUser {
	all := []RestrictedUser{}
	users := []RestrictedUser{}
	db.Find(&all)
	for _, u := range all {
		if u.ParentUsername == parent {
			users = append(users, u)
		}
	}
	return users
}

// PutUser updates the given user
func PutRestrictedUser(u *RestrictedUser) error {
	err := db.Save(u).Error
	return err
}
