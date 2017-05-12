package models

import (
	"errors"
	"sort"
	"time"

	"github.com/jinzhu/gorm"
)

// Campaign is a struct representing a created campaign
type Campaign struct {
	Id            int64     `json:"id"`
	UserId        int64     `json:"-"`
	Name          string    `json:"name" sql:"not null"`
	CreatedDate   time.Time `json:"created_date"`
	LaunchDate    time.Time `json:"launch_date"`
	CompletedDate time.Time `json:"completed_date"`
	TemplateId    int64     `json:"-"`
	Template      Template  `json:"template"`
	PageId        int64     `json:"-"`
	Page          Page      `json:"page"`
	Status        string    `json:"status"`
	Results       []Result  `json:"results,omitempty"`
	Groups        []Group   `json:"groups,omitempty"`
	Events        []Event   `json:"timeline,omitemtpy"`
	SMTPId        int64     `json:"-"`
	SMTP          SMTP      `json:"smtp"`
	URL           string    `json:"url"`
}

// CampaignResults is a struct representing the results from a campaign
type CampaignResults struct {
	Id      int64    `json:"id"`
	Name    string   `json:"name"`
	Status  string   `json:"status"`
	Results []Result `json:"results, omitempty"`
	Events  []Event  `json:"timeline,omitempty"`
}

// CampaignSummaryStats contains the summary stats for a campaign.
type CampaignSummaryStats struct {
	ID                       int64  `json:"id"`
	Name                     string `json:"name"`
	CreatedDate              string `json:"created_date"`
	Status                   string `json:"status"`
	Sent                     int    `json:"sent"`
	ErrorSending             int    `json:"error_sending"`
	Opened                   int    `json:"opened"`
	Clicked                  int    `json:"clicked"`
	CredentialsEntered       int    `json:"credentialsentered"`
	UniqueCredentialsEntered int    `json:"uniquecredentialsentered"`
}

// ErrCampaignNameNotSpecified indicates there was no template given by the user
var ErrCampaignNameNotSpecified = errors.New("Campaign name not specified")

// ErrGroupNotSpecified indicates there was no template given by the user
var ErrGroupNotSpecified = errors.New("No groups specified")

// ErrTemplateNotSpecified indicates there was no template given by the user
var ErrTemplateNotSpecified = errors.New("No email template specified")

// ErrPageNotSpecified indicates a landing page was not provided for the campaign
var ErrPageNotSpecified = errors.New("No landing page specified")

// ErrSMTPNotSpecified indicates a sending profile was not provided for the campaign
var ErrSMTPNotSpecified = errors.New("No sending profile specified")

// ErrTemplateNotFound indicates the template specified does not exist in the database
var ErrTemplateNotFound = errors.New("Template not found")

// ErrGroupnNotFound indicates a group specified by the user does not exist in the database
var ErrGroupNotFound = errors.New("Group not found")

// ErrPageNotFound indicates a page specified by the user does not exist in the database
var ErrPageNotFound = errors.New("Page not found")

// ErrSMTPNotFound indicates a sending profile specified by the user does not exist in the database
var ErrSMTPNotFound = errors.New("Sending profile not found")

// Validate checks to make sure there are no invalid fields in a submitted campaign
func (c *Campaign) Validate() error {
	switch {
	case c.Name == "":
		return ErrCampaignNameNotSpecified
	case len(c.Groups) == 0:
		return ErrGroupNotSpecified
	case c.Template.Name == "":
		return ErrTemplateNotSpecified
	case c.Page.Name == "":
		return ErrPageNotSpecified
	case c.SMTP.Name == "":
		return ErrSMTPNotSpecified
	}
	return nil
}

// SendTestEmailRequest is the structure of a request
// to send a test email to test an SMTP connection
type SendTestEmailRequest struct {
	Template    Template `json:"template"`
	Page        Page     `json:"page"`
	SMTP        SMTP     `json:"smtp"`
	URL         string   `json:"url"`
	Tracker     string   `json:"tracker"`
	TrackingURL string   `json:"tracking_url"`
	From        string   `json:"from"`
	Target
}

// Validate ensures the SendTestEmailRequest structure
// is valid.
func (s *SendTestEmailRequest) Validate() error {
	switch {
	case s.Email == "":
		return ErrEmailNotSpecified
	}
	return nil
}

// UpdateStatus changes the campaign status appropriately
func (c *Campaign) UpdateStatus(s string) error {
	// This could be made simpler, but I think there's a bug in gorm
	return db.Table("campaigns").Where("id=?", c.Id).Update("status", s).Error
}

// AddEvent creates a new campaign event in the database
func (c *Campaign) AddEvent(e Event) error {
	e.CampaignId = c.Id
	e.Time = time.Now()
	return db.Debug().Save(&e).Error
}

// GetDetails retrieves the related attributes of the campaign
// from the database. If the Events and the Results are not available,
// an error is returned. Otherwise, the attribute name is set to [Deleted],
// indicating the user deleted the attribute (template, smtp, etc.)
func (c *Campaign) GetDetails(includeEvents, includeTemplates, includePage, includeSendingProfile bool) error {
	err = db.Model(c).Related(&c.Results).Error
	if err != nil {
		Logger.Printf("%s: results not found for campaign\n", err)
		return err
	}
	if includeEvents {
		err = db.Model(c).Related(&c.Events).Error
		if err != nil {
			Logger.Printf("%s: events not found for campaign\n", err)
			return err
		}
	}
	if includeTemplates {
		err = db.Table("templates").Where("id=?", c.TemplateId).Find(&c.Template).Error
		if err != nil {
			if err != gorm.ErrRecordNotFound {
				return err
			}
			c.Template = Template{Name: "[Deleted]"}
			Logger.Printf("%s: template not found for campaign\n", err)
		}
		err = db.Where("template_id=?", c.Template.Id).Find(&c.Template.Attachments).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			Logger.Println(err)
			return err
		}
	}
	if includePage {
		err = db.Table("pages").Where("id=?", c.PageId).Find(&c.Page).Error
		if err != nil {
			if err != gorm.ErrRecordNotFound {
				return err
			}
			c.Page = Page{Name: "[Deleted]"}
			Logger.Printf("%s: page not found for campaign\n", err)
		}
	}
	if includeSendingProfile {
		err = db.Table("smtp").Where("id=?", c.SMTPId).Find(&c.SMTP).Error
		if err != nil {
			// Check if the SMTP was deleted
			if err != gorm.ErrRecordNotFound {
				return err
			}
			c.SMTP = SMTP{Name: "[Deleted]"}
			Logger.Printf("%s: sending profile not found for campaign\n", err)
		}
	}
	return nil
}

// GetRawResults
func GetRawResults(campaignIds []int) ([]Result, error) {
	results := []Result{}
	sort.Ints(campaignIds)
	rows, err := db.Raw("select Campaign_Id, Status from results").Rows() // (*sql.Rows, error)
	if err != nil {
		Logger.Println(err)
		return nil, err
	}
	defer rows.Close()
	var campaignId int64
	status := ""
	for rows.Next() {
		rows.Scan(&campaignId, &status)
		idx := sort.SearchInts(campaignIds, int(campaignId))
		if idx < len(campaignIds) && campaignIds[idx] == int(campaignId) {
			// campaignId is present at campaignIds[i]
			results = append(results, Result{CampaignId: campaignId, Status: status})
		}
	}
	return results, nil
}

// GetTargetResults calculates the results for the campaign matching the given status
func (c *Campaign) GetPhishingResults(statuses []string) []TrimmedResult {
	var sent, opened, clicked, submittedData, uniqueCreds int
	var creds map[string]bool
	creds = make(map[string]bool)
	var results = []TrimmedResult{}
	for _, r := range c.Results {
		switch r.Status {
		case "Email Sent":
			sent++
			break
		case "Email Opened":
			sent++
			opened++
			break
		case "Clicked Link":
			sent++
			opened++
			clicked++
			break
		case "Success":
			sent++
			opened++
			clicked++
			r.Status = "Clicked Link"
			break
		}
		for _, e := range c.Events {
			if e.Email == r.Email && e.Message == "Submitted Data" {
				r.Status = "Submitted Data"
				submittedData++
				if _, ok := creds[e.Email]; ok {

				} else {
					creds[e.Email] = true
					uniqueCreds++
				}
			}
		}
		for _, status := range statuses {
			if r.Status == status {
				trimmedResult := TrimmedResult{
					CampaignID:   r.CampaignId,
					CampaignName: c.Name,
					Email:        r.Email,
					FirstName:    r.FirstName,
					LastName:     r.LastName,
					Position:     r.Position,
					Department:   r.Department,
					Status:       r.Status}
				results = append(results, trimmedResult)
				break
			}
		}
	}
	return results
}

// Event contains the fields for an event
// that occurs during the campaign
type Event struct {
	Id         int64     `json:"-"`
	CampaignId int64     `json:"campaignid"`
	Email      string    `json:"email"`
	Time       time.Time `json:"time"`
	Message    string    `json:"message"`
	Details    string    `json:"details"`
}

// GetCampaignsRaw returns the campaigns owned by the given user.
func GetCampaignsRaw(uid int64, includeDetails bool) ([]Campaign, error) {
	cs := []Campaign{}
	err := db.Model(&User{Id: uid}).Related(&cs).Error
	if err != nil {
		Logger.Println(err)
	}
	return cs, err
}

// GetCampaigns returns the campaigns owned by the given user.
func GetCampaigns(uid int64, includeDetails bool) ([]Campaign, error) {
	cs := []Campaign{}
	err := db.Model(&User{Id: uid}).Related(&cs).Error
	if err != nil {
		Logger.Println(err)
	}
	for i, _ := range cs {
		err = cs[i].GetDetails(includeDetails, includeDetails, includeDetails, includeDetails)
		if err != nil {
			Logger.Println(err)
		}
	}
	return cs, err
}

// GetCampaignNames returns the names of the campaigns owned by the given user.
func GetCampaignNames(uid int64) ([]Campaign, error) {
	cs := []Campaign{}
	db.Select("id, name, created_date, status").Find(&cs, Campaign{UserId: uid})
	// if err != nil {
	// 	Logger.Println(err)
	// }
	// return cs, err
	return cs, nil
}

// GetCampaign returns the campaign, if it exists, specified by the given id and user_id.
func GetCampaign(id int64, uid int64) (Campaign, error) {
	c := Campaign{}
	err := db.Where("id = ?", id).Where("user_id = ?", uid).Find(&c).Error
	if err != nil {
		Logger.Printf("%s: campaign not found\n", err)
		return c, err
	}
	err = c.GetDetails(true, true, true, true)
	return c, err
}

// GetEmailEvents returns events matching the given email address
func GetEmailEvents(email string) []Event {
	targetEvents := []Event{}
	rows, err := db.Raw("select campaign_id, time, message from events where email = ?", email).Rows()
	if err != nil {
		Logger.Println("ERROR GetEmailEvents:", err)
		return targetEvents
	}
	defer rows.Close()
	for rows.Next() {
		var campaignid int64
		var time time.Time
		var message string
		rows.Scan(&campaignid, &time, &message)
		targetEvents = append(targetEvents, Event{CampaignId: campaignid, Time: time, Message: message})
	}
	return targetEvents
}

func GetCampaignResults(id int64, uid int64) (CampaignResults, error) {
	cr := CampaignResults{}
	err := db.Table("campaigns").Where("id=? and user_id=?", id, uid).Find(&cr).Error
	if err != nil {
		Logger.Printf("%s: campaign not found\n", err)
		return cr, err
	}
	err = db.Table("results").Where("campaign_id=? and user_id=?", cr.Id, uid).Find(&cr.Results).Error
	if err != nil {
		Logger.Printf("%s: results not found for campaign\n", err)
		return cr, err
	}
	err = db.Table("events").Where("campaign_id=?", cr.Id).Find(&cr.Events).Error
	if err != nil {
		Logger.Printf("%s: events not found for campaign\n", err)
		return cr, err
	}
	return cr, err
}

// GetQueuedCampaigns returns the campaigns that are queued up for this given minute
func GetQueuedCampaigns(t time.Time) ([]Campaign, error) {
	cs := []Campaign{}
	err := db.Where("launch_date <= ?", t).
		Where("status = ?", CAMPAIGN_QUEUED).Find(&cs).Error
	if err != nil {
		Logger.Println(err)
	}
	Logger.Printf("Found %d Campaigns to run\n", len(cs))
	for i, _ := range cs {
		err = cs[i].GetDetails(true, true, true, true)
		if err != nil {
			Logger.Println(err)
		}
	}
	return cs, err
}

// PostCampaign inserts a campaign and all associated records into the database.
func PostCampaign(c *Campaign, uid int64) error {
	if err := c.Validate(); err != nil {
		return err
	}
	// Fill in the details
	c.UserId = uid
	c.CreatedDate = time.Now()
	c.CompletedDate = time.Time{}
	c.Status = CAMPAIGN_CREATED
	if c.LaunchDate.IsZero() {
		c.LaunchDate = time.Now()
	}
	// Check to make sure all the groups already exist
	for i, g := range c.Groups {
		c.Groups[i], err = GetGroupByName(g.Name, uid)
		if err == gorm.ErrRecordNotFound {
			Logger.Printf("Error - Group %s does not exist", g.Name)
			return ErrGroupNotFound
		} else if err != nil {
			Logger.Println(err)
			return err
		}
	}
	// Check to make sure the template exists
	t, err := GetTemplateByName(c.Template.Name, uid)
	if err == gorm.ErrRecordNotFound {
		Logger.Printf("Error - Template %s does not exist", t.Name)
		return ErrTemplateNotFound
	} else if err != nil {
		Logger.Println(err)
		return err
	}
	c.Template = t
	c.TemplateId = t.Id
	// Check to make sure the page exists
	p, err := GetPageByName(c.Page.Name, uid)
	if err == gorm.ErrRecordNotFound {
		Logger.Printf("Error - Page %s does not exist", p.Name)
		return ErrPageNotFound
	} else if err != nil {
		Logger.Println(err)
		return err
	}
	c.Page = p
	c.PageId = p.Id
	// Check to make sure the sending profile exists
	s, err := GetSMTPByName(c.SMTP.Name, uid)
	if err == gorm.ErrRecordNotFound {
		Logger.Printf("Error - Sending profile %s does not exist", s.Name)
		return ErrPageNotFound
	} else if err != nil {
		Logger.Println(err)
		return err
	}
	c.SMTP = s
	c.SMTPId = s.Id
	// Insert into the DB
	err = db.Save(c).Error
	if err != nil {
		Logger.Println(err)
		return err
	}
	err = c.AddEvent(Event{Message: "Campaign Created"})
	if err != nil {
		Logger.Println(err)
	}
	// Insert all the results
	for _, g := range c.Groups {
		// Insert a result for each target in the group
		for _, t := range g.Targets {
			r := &Result{Email: t.Email, Position: t.Position, Department: t.Department, Status: STATUS_SENDING, CampaignId: c.Id, UserId: c.UserId, FirstName: t.FirstName, LastName: t.LastName}
			r.GenerateId()
			err = db.Save(r).Error
			if err != nil {
				Logger.Printf("Error adding result record for target %s\n", t.Email)
				Logger.Println(err)
			}
			c.Results = append(c.Results, *r)
		}
	}
	c.Status = CAMPAIGN_QUEUED
	err = db.Save(c).Error
	return err
}

//DeleteCampaign deletes the specified campaign
func DeleteCampaign(id int64) error {
	Logger.Printf("Deleting campaign %d\n", id)
	// Delete all the campaign results
	err := db.Where("campaign_id=?", id).Delete(&Result{}).Error
	if err != nil {
		Logger.Println(err)
		return err
	}
	err = db.Where("campaign_id=?", id).Delete(&Event{}).Error
	if err != nil {
		Logger.Println(err)
		return err
	}
	// Delete the campaign
	err = db.Delete(&Campaign{Id: id}).Error
	if err != nil {
		Logger.Println(err)
	}
	return err
}

// CompleteCampaign effectively "ends" a campaign.
// Any future emails clicked will return a simple "404" page.
func CompleteCampaign(id int64, uid int64) error {
	Logger.Printf("Marking campaign %d as complete\n", id)
	c, err := GetCampaign(id, uid)
	if err != nil {
		return err
	}
	// Don't overwrite original completed time
	if c.Status == CAMPAIGN_COMPLETE {
		return nil
	}
	// Mark the campaign as complete
	c.CompletedDate = time.Now()
	c.Status = CAMPAIGN_COMPLETE
	err = db.Where("id=? and user_id=?", id, uid).Save(&c).Error
	if err != nil {
		Logger.Println(err)
	}
	return err
}
