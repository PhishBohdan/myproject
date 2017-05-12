package controllers

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gophish/gophish/auth"
	ctx "github.com/gophish/gophish/context"
	"github.com/gophish/gophish/models"
	"github.com/gophish/gophish/util"
	"github.com/gophish/gophish/worker"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/jordan-wright/email"
)

// Worker is the worker that processes phishing events and updates campaigns.
var Worker *worker.Worker

func init() {
	Worker = worker.New()
	go Worker.Start()
}

// API (/api) provides access to api documentation
func API(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		templates := template.New("template")
		_, err := templates.ParseFiles("templates/docs.html")
		if err != nil {
			Logger.Println(err)
		}
		template.Must(templates, err).ExecuteTemplate(w, "base", nil)
	}
}

// API (/api/reset) resets a user's API key
func API_Reset(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "POST":
		u := ctx.Get(r, "user").(models.User)
		u.ApiKey = auth.GenerateSecureKey()
		err := models.PutUser(&u)
		if err != nil {
			http.Error(w, "Error setting API Key", http.StatusInternalServerError)
		} else {
			JSONResponse(w, models.Response{Success: true, Message: "API Key successfully reset!", Data: u.ApiKey}, http.StatusOK)
		}
	}
}

// API_Campaigns returns a list of campaigns if requested via GET.
// If requested via POST, API_Campaigns creates a new campaign and returns a reference to it.
func API_Campaigns(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		var restrictedUser models.RestrictedUser
		if ctx.Get(r, "restricteduser") == nil {
		} else {
			restrictedUser = ctx.Get(r, "restricteduser").(models.RestrictedUser)
		}

		allCampaigns, err := models.GetCampaigns(ctx.Get(r, "user_id").(int64), true)
		if err != nil {
			Logger.Println(err)
		}
		cs := []models.Campaign{}
		if ctx.Get(r, "restricteduser") != nil {
			cs = restrictedUser.FilterAllowedCampaigns(allCampaigns)
		} else {
			cs = allCampaigns
		}
		vars := mux.Vars(r)
		max, err := strconv.ParseInt(vars["max"], 0, 64)
		if err != nil {
			max = 0
		}
		// max == 0 denotes return all results
		if max != 0 && len(cs) > int(max) {
			cs = cs[0:max]
		}
		JSONResponse(w, cs, http.StatusOK)
	//POST: Create a new campaign and return it as JSON
	case r.Method == "POST":
		c := models.Campaign{}
		// Put the request into a campaign
		err := json.NewDecoder(r.Body).Decode(&c)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid JSON structure"}, http.StatusBadRequest)
			return
		}
		err = models.PostCampaign(&c, ctx.Get(r, "user_id").(int64))
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		JSONResponse(w, c, http.StatusCreated)
	}
}

func matchFilter(campaignName string, filterTerms []string) bool {
	for _, filter := range filterTerms {
		if !strings.Contains(campaignName, filter) {
			return false
		}
	}
	return true
}

// API_CampaignsRange returns a list of campaigns in the given range of Id's.
func API_CampaignsRange(w http.ResponseWriter, r *http.Request) {
	var restrictedUser models.RestrictedUser
	filter := ""
	matchExact := "off"
	filter = r.URL.Query().Get("filter")
	if r.URL.Query().Get("matchExact") != "" {
		matchExact = r.URL.Query().Get("matchExact")
	}
	filterTerms := []string{}
	if matchExact == "on" {
		filterTerms = []string{filter}
	} else {
		filterTerms = strings.Split(filter, " ")
	}

	if ctx.Get(r, "restricteduser") == nil {
	} else {
		restrictedUser = ctx.Get(r, "restricteduser").(models.RestrictedUser)
	}
	allCampaigns, err := models.GetCampaigns(ctx.Get(r, "user_id").(int64), true)
	if err != nil {
		Logger.Println(err)
	}
	cs := []models.Campaign{}
	if ctx.Get(r, "restricteduser") != nil {
		cs = restrictedUser.FilterAllowedCampaigns(allCampaigns)
	} else {
		cs = allCampaigns
	}
	csFiltered := []models.Campaign{}
	for _, c := range cs {
		if matchFilter(c.Name, filterTerms) || len(filterTerms) == 0 {
			csFiltered = append(csFiltered, c)
		}
	}
	cs = csFiltered
	vars := mux.Vars(r)
	max, err := strconv.ParseInt(vars["max"], 0, 64)
	if err != nil {
		max = 0
	}
	min, err := strconv.ParseInt(vars["min"], 0, 64)
	if err != nil {
		min = 0
	}
	// max == 0 denotes return all results
	if max != 0 {
		if int(max) > len(cs) {
			max = int64(len(cs))
		}
		if min > max {
			min = max
		}
		cs = cs[min:max]
	}
	JSONResponse(w, cs, http.StatusOK)
}

// API_SummaryStats returns a list campaign summary stats that match the given filter.
func API_SummaryStats(w http.ResponseWriter, r *http.Request) {
	var restrictedUser models.RestrictedUser
	filter := ""
	matchExact := "off"
	filter = r.URL.Query().Get("filter")
	if r.URL.Query().Get("matchExact") != "" {
		matchExact = r.URL.Query().Get("matchExact")
	}
	filterTerms := []string{}
	if matchExact == "on" {
		filterTerms = []string{filter}
	} else {
		filterTerms = strings.Split(filter, " ")
	}
	fmt.Println("FILTER:", filter)
	fmt.Println("MATCH_EXACT:", matchExact)
	if ctx.Get(r, "restricteduser") == nil {
	} else {
		restrictedUser = ctx.Get(r, "restricteduser").(models.RestrictedUser)
	}
	allCampaigns, err := models.GetCampaignsRaw(ctx.Get(r, "user_id").(int64), false)
	if err != nil {
		Logger.Println(err)
	}
	cs := []models.Campaign{}
	if ctx.Get(r, "restricteduser") != nil {
		cs = restrictedUser.FilterAllowedCampaigns(allCampaigns)
	} else {
		cs = allCampaigns
	}
	csFiltered := []models.Campaign{}
	for _, c := range cs {
		if matchFilter(c.Name, filterTerms) || len(filterTerms) == 0 {
			csFiltered = append(csFiltered, c)
		}
	}
	cs = csFiltered
	var summaryStats []models.CampaignSummaryStats
	summaryStats = make([]models.CampaignSummaryStats, 0, 0)
	campaignIds := []int{}
	for _, c := range cs {
		campaignIds = append(campaignIds, int(c.Id))
		summaryStats = append(summaryStats, models.CampaignSummaryStats{ID: c.Id, Name: c.Name, Status: c.Status, CreatedDate: c.CreatedDate.String()})
	}
	rawResults, err := models.GetRawResults(campaignIds)
	if err != nil {
		Logger.Println(err)
	}
	for _, result := range rawResults {
		for i := range summaryStats {
			if summaryStats[i].ID == result.CampaignId {
				switch result.Status {
				case "Error Sending Email":
					summaryStats[i].ErrorSending++
				case "Email Sent":
					summaryStats[i].Sent++
				case "Email Opened":
					summaryStats[i].Sent++
					summaryStats[i].Opened++
				case "Clicked Link":
					summaryStats[i].Sent++
					summaryStats[i].Opened++
					summaryStats[i].Clicked++
				case "Success":
					summaryStats[i].Sent++
					summaryStats[i].Opened++
					summaryStats[i].Clicked++
					summaryStats[i].CredentialsEntered++
					summaryStats[i].UniqueCredentialsEntered++
				}
			}
		}
	}
	JSONResponse(w, summaryStats, http.StatusOK)
}

// API_PhishingResults returns a list campaign(s) targets that match the given filter.
func API_PhishingResults(w http.ResponseWriter, r *http.Request) {
	var restrictedUser models.RestrictedUser
	filter := ""
	matchExact := "off"
	statuses := []string{"Email Sent"}
	filter = r.URL.Query().Get("filter")
	if r.URL.Query().Get("matchExact") != "" {
		matchExact = r.URL.Query().Get("matchExact")
	}
	if s, ok := r.Form["status[]"]; ok {
		statuses = s
	}
	filterTerms := []string{}
	if matchExact == "on" {
		filterTerms = []string{filter}
	} else {
		filterTerms = strings.Split(filter, " ")
	}
	if ctx.Get(r, "restricteduser") == nil {
	} else {
		restrictedUser = ctx.Get(r, "restricteduser").(models.RestrictedUser)
	}
	allCampaigns, err := models.GetCampaigns(ctx.Get(r, "user_id").(int64), false)
	if err != nil {
		Logger.Println(err)
	}
	cs := []models.Campaign{}
	if ctx.Get(r, "restricteduser") != nil {
		cs = restrictedUser.FilterAllowedCampaigns(allCampaigns)
	} else {
		cs = allCampaigns
	}
	csFiltered := []models.Campaign{}
	for _, c := range cs {
		if matchFilter(c.Name, filterTerms) || len(filterTerms) == 0 {
			csFiltered = append(csFiltered, c)
		}
	}
	cs = csFiltered
	var phishingResults []models.TrimmedResult
	phishingResults = make([]models.TrimmedResult, 0, 0)
	for _, c := range cs {
		// Only get events, nothing else
		if err := c.GetDetails(true, false, false, false); err != nil {
			// TODO: handle error
			Logger.Println(err)
		} else {
			for _, r := range c.GetPhishingResults(statuses) {
				phishingResults = append(phishingResults, r)
			}
		}
	}
	if len(phishingResults) >= 2001 {
		phishingResults = phishingResults[0:2001]
	}
	JSONResponse(w, phishingResults, http.StatusOK)
}

// API_TargetEvents returns a list target results that match the given email address.
func API_TargetEvents(w http.ResponseWriter, r *http.Request) {
	var restrictedUser models.RestrictedUser
	email := r.URL.Query().Get("email")
	if ctx.Get(r, "restricteduser") == nil {
	} else {
		restrictedUser = ctx.Get(r, "restricteduser").(models.RestrictedUser)
	}
	allCampaigns, err := models.GetCampaigns(ctx.Get(r, "user_id").(int64), false)
	if err != nil {
		Logger.Println(err)
	}
	cs := []models.Campaign{}
	if ctx.Get(r, "restricteduser") != nil {
		cs = restrictedUser.FilterAllowedCampaigns(allCampaigns)
	} else {
		cs = allCampaigns
	}
	events := models.GetEmailEvents(email)
	filteredEvents := []models.Event{}
	for _, event := range events {
		for _, c := range cs {
			if c.Id == event.CampaignId {
				filteredEvents = append(filteredEvents, event)
				break
			}
		}
	}
	JSONResponse(w, filteredEvents, http.StatusOK)
}

// API_CampaignNames returns a list of campaign names.
func API_CampaignNames(w http.ResponseWriter, r *http.Request) {
	var restrictedUser models.RestrictedUser
	vars := mux.Vars(r)
	max, err := strconv.ParseInt(vars["max"], 0, 64)
	if err != nil {
		Logger.Println(err)
	}
	if ctx.Get(r, "restricteduser") == nil {
	} else {
		restrictedUser = ctx.Get(r, "restricteduser").(models.RestrictedUser)
	}
	allCampaignNames, err := models.GetCampaignNames(ctx.Get(r, "user_id").(int64))
	allCampaigns, err := models.GetCampaigns(ctx.Get(r, "user_id").(int64), true)
	if err != nil {
		Logger.Println(err)
	}
	cs := []models.Campaign{}
	if ctx.Get(r, "restricteduser") != nil {
		for _, c := range allCampaignNames {
			add := false
			for _, camp := range allCampaigns {
				if c.Id == camp.Id {
					if restrictedUser.AllowCampaign(camp) {
						add = true
						break
					}
				}
			}
			if add {
				cs = append(cs, c)
			}
		}
	} else {
		cs = allCampaignNames
	}
	// max == 0 denotes return all results
	if max != 0 && len(cs) > int(max) {
		cs = cs[0:max]
	}
	JSONResponse(w, cs, http.StatusOK)
}

// API_Campaigns_Id returns details about the requested campaign. If the campaign is not
// valid, API_Campaigns_Id returns null.
func API_Campaigns_Id(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	c, err := models.GetCampaign(id, ctx.Get(r, "user_id").(int64))

	var restrictedUser models.RestrictedUser
	if ctx.Get(r, "restricteduser") == nil {
	} else {
		restrictedUser = ctx.Get(r, "restricteduser").(models.RestrictedUser)
		add := restrictedUser.AllowCampaign(c)
		if !add {
			JSONResponse(w, models.Response{Success: false, Message: "Campaign permission denied"}, http.StatusNotFound)
		}

	}

	if err != nil {
		Logger.Println(err)
		JSONResponse(w, models.Response{Success: false, Message: "Campaign not found"}, http.StatusNotFound)
		return
	}
	switch {
	case r.Method == "GET":
		JSONResponse(w, c, http.StatusOK)
	case r.Method == "DELETE":
		err = models.DeleteCampaign(id)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error deleting campaign"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Campaign deleted successfully!"}, http.StatusOK)
	}
}

// API_Campaigns_Id_Results returns just the results for a given campaign to
// significantly reduce the information returned.
func API_Campaigns_Id_Results(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	cr, err := models.GetCampaignResults(id, ctx.Get(r, "user_id").(int64))
	if err != nil {
		Logger.Println(err)
		JSONResponse(w, models.Response{Success: false, Message: "Campaign not found"}, http.StatusNotFound)
		return
	}

	var restrictedUser models.RestrictedUser
	if ctx.Get(r, "restricteduser") == nil {
	} else {
		restrictedUser = ctx.Get(r, "restricteduser").(models.RestrictedUser)
		add := restrictedUser.AllowCampaignResults(cr)
		if !add {
			JSONResponse(w, models.Response{Success: false, Message: "Campaign permission denied"}, http.StatusNotFound)
		}

	}

	if r.Method == "GET" {
		JSONResponse(w, cr, http.StatusOK)
		return
	}
}

// API_Campaigns_Id_Complete effectively "ends" a campaign.
// Future phishing emails clicked will return a simple "404" page.
func API_Campaigns_Id_Complete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	switch {
	case r.Method == "GET":
		err := models.CompleteCampaign(id, ctx.Get(r, "user_id").(int64))
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error completing campaign"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Campaign completed successfully!"}, http.StatusOK)
	}
}

// API_Groups returns a list of groups if requested via GET.
// If requested via POST, API_Groups creates a new group and returns a reference to it.
func API_Groups(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		gs, err := models.GetGroups(ctx.Get(r, "user_id").(int64))
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "No groups found"}, http.StatusNotFound)
			return
		}
		JSONResponse(w, gs, http.StatusOK)
	//POST: Create a new group and return it as JSON
	case r.Method == "POST":
		g := models.Group{}
		// Put the request into a group
		err := json.NewDecoder(r.Body).Decode(&g)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid JSON structure"}, http.StatusBadRequest)
			return
		}
		_, err = models.GetGroupByName(g.Name, ctx.Get(r, "user_id").(int64))
		if err != gorm.ErrRecordNotFound {
			JSONResponse(w, models.Response{Success: false, Message: "Group name already in use"}, http.StatusConflict)
			return
		}
		g.ModifiedDate = time.Now()
		g.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PostGroup(&g)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		JSONResponse(w, g, http.StatusCreated)
	}
}

// API_Groups_Id returns details about the requested group.
// If the group is not valid, API_Groups_Id returns null.
func API_Groups_Id(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	g, err := models.GetGroup(id, ctx.Get(r, "user_id").(int64))
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Group not found"}, http.StatusNotFound)
		return
	}
	switch {
	case r.Method == "GET":
		JSONResponse(w, g, http.StatusOK)
	case r.Method == "DELETE":
		err = models.DeleteGroup(&g)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error deleting group"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Group deleted successfully!"}, http.StatusOK)
	case r.Method == "PUT":
		// Change this to get from URL and uid (don't bother with id in r.Body)
		g = models.Group{}
		err = json.NewDecoder(r.Body).Decode(&g)
		if g.Id != id {
			JSONResponse(w, models.Response{Success: false, Message: "Error: /:id and group_id mismatch"}, http.StatusInternalServerError)
			return
		}
		g.ModifiedDate = time.Now()
		g.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PutGroup(&g)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		JSONResponse(w, g, http.StatusOK)
	}
}

// API_Templates handles the functionality for the /api/templates endpoint
func API_Templates(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		ts, err := models.GetTemplates(ctx.Get(r, "user_id").(int64))
		if err != nil {
			Logger.Println(err)
		}
		JSONResponse(w, ts, http.StatusOK)
	//POST: Create a new template and return it as JSON
	case r.Method == "POST":
		t := models.Template{}
		// Put the request into a template
		err := json.NewDecoder(r.Body).Decode(&t)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid JSON structure"}, http.StatusBadRequest)
			return
		}
		_, err = models.GetTemplateByName(t.Name, ctx.Get(r, "user_id").(int64))
		if err != gorm.ErrRecordNotFound {
			JSONResponse(w, models.Response{Success: false, Message: "Template name already in use"}, http.StatusConflict)
			return
		}
		t.ModifiedDate = time.Now()
		t.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PostTemplate(&t)
		if err == models.ErrTemplateNameNotSpecified {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		if err == models.ErrTemplateMissingParameter {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error inserting template into database"}, http.StatusInternalServerError)
			Logger.Println(err)
			return
		}
		JSONResponse(w, t, http.StatusCreated)
	}
}

// API_Templates_Id handles the functions for the /api/templates/:id endpoint
func API_Templates_Id(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	t, err := models.GetTemplate(id, ctx.Get(r, "user_id").(int64))
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Template not found"}, http.StatusNotFound)
		return
	}
	switch {
	case r.Method == "GET":
		JSONResponse(w, t, http.StatusOK)
	case r.Method == "DELETE":
		err = models.DeleteTemplate(id, ctx.Get(r, "user_id").(int64))
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error deleting template"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Template deleted successfully!"}, http.StatusOK)
	case r.Method == "PUT":
		t = models.Template{}
		err = json.NewDecoder(r.Body).Decode(&t)
		if err != nil {
			Logger.Println(err)
		}
		if t.Id != id {
			JSONResponse(w, models.Response{Success: false, Message: "Error: /:id and template_id mismatch"}, http.StatusBadRequest)
			return
		}
		t.ModifiedDate = time.Now()
		t.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PutTemplate(&t)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		JSONResponse(w, t, http.StatusOK)
	}
}

// API_Pages handles requests for the /api/pages/ endpoint
func API_Pages(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		ps, err := models.GetPages(ctx.Get(r, "user_id").(int64))
		if err != nil {
			Logger.Println(err)
		}
		JSONResponse(w, ps, http.StatusOK)
	//POST: Create a new page and return it as JSON
	case r.Method == "POST":
		p := models.Page{}
		// Put the request into a page
		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid request"}, http.StatusBadRequest)
			return
		}
		// Check to make sure the name is unique
		_, err = models.GetPageByName(p.Name, ctx.Get(r, "user_id").(int64))
		if err != gorm.ErrRecordNotFound {
			JSONResponse(w, models.Response{Success: false, Message: "Page name already in use"}, http.StatusConflict)
			Logger.Println(err)
			return
		}
		p.ModifiedDate = time.Now()
		p.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PostPage(&p)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, p, http.StatusCreated)
	}
}

// API_Pages_Id contains functions to handle the GET'ing, DELETE'ing, and PUT'ing
// of a Page object
func API_Pages_Id(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	p, err := models.GetPage(id, ctx.Get(r, "user_id").(int64))
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Page not found"}, http.StatusNotFound)
		return
	}
	switch {
	case r.Method == "GET":
		JSONResponse(w, p, http.StatusOK)
	case r.Method == "DELETE":
		err = models.DeletePage(id, ctx.Get(r, "user_id").(int64))
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error deleting page"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Page Deleted Successfully"}, http.StatusOK)
	case r.Method == "PUT":
		p = models.Page{}
		err = json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			Logger.Println(err)
		}
		if p.Id != id {
			JSONResponse(w, models.Response{Success: false, Message: "/:id and /:page_id mismatch"}, http.StatusBadRequest)
			return
		}
		p.ModifiedDate = time.Now()
		p.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PutPage(&p)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error updating page: " + err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, p, http.StatusOK)
	}
}

// API_SMTP handles requests for the /api/smtp/ endpoint
func API_SMTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		ss, err := models.GetSMTPs(ctx.Get(r, "user_id").(int64))
		if err != nil {
			Logger.Println(err)
		}
		JSONResponse(w, ss, http.StatusOK)
	//POST: Create a new SMTP and return it as JSON
	case r.Method == "POST":
		s := models.SMTP{}
		// Put the request into a page
		err := json.NewDecoder(r.Body).Decode(&s)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid request"}, http.StatusBadRequest)
			return
		}
		// Check to make sure the name is unique
		_, err = models.GetSMTPByName(s.Name, ctx.Get(r, "user_id").(int64))
		if err != gorm.ErrRecordNotFound {
			JSONResponse(w, models.Response{Success: false, Message: "SMTP name already in use"}, http.StatusConflict)
			Logger.Println(err)
			return
		}
		s.ModifiedDate = time.Now()
		s.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PostSMTP(&s)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, s, http.StatusCreated)
	}
}

// API_SMTP_Id contains functions to handle the GET'ing, DELETE'ing, and PUT'ing
// of a SMTP object
func API_SMTP_Id(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	s, err := models.GetSMTP(id, ctx.Get(r, "user_id").(int64))
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "SMTP not found"}, http.StatusNotFound)
		return
	}
	switch {
	case r.Method == "GET":
		JSONResponse(w, s, http.StatusOK)
	case r.Method == "DELETE":
		err = models.DeleteSMTP(id, ctx.Get(r, "user_id").(int64))
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error deleting SMTP"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "SMTP Deleted Successfully"}, http.StatusOK)
	case r.Method == "PUT":
		s = models.SMTP{}
		err = json.NewDecoder(r.Body).Decode(&s)
		if err != nil {
			Logger.Println(err)
		}
		if s.Id != id {
			JSONResponse(w, models.Response{Success: false, Message: "/:id and /:smtp_id mismatch"}, http.StatusBadRequest)
			return
		}
		err = s.Validate()
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		s.ModifiedDate = time.Now()
		s.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PutSMTP(&s)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error updating page"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, s, http.StatusOK)
	}
}

// API_Import_Group imports a CSV of group members
func API_Import_Group(w http.ResponseWriter, r *http.Request) {
	ts, err := util.ParseCSV(r)
	if err != nil {
		fmt.Println("IMPORT GROUP ERR:", err)
		JSONResponse(w, models.Response{Success: false, Message: "Error parsing CSV"}, http.StatusInternalServerError)
		return
	}
	JSONResponse(w, ts, http.StatusOK)
	return
}

// API_Import_Email allows for the importing of email.
// Returns a Message object
func API_Import_Email(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		JSONResponse(w, models.Response{Success: false, Message: "Method not allowed"}, http.StatusBadRequest)
		return
	}
	ir := struct {
		Content      string `json:"content"`
		ConvertLinks bool   `json:"convert_links"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&ir)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Error decoding JSON Request"}, http.StatusBadRequest)
		return
	}
	e, err := email.NewEmailFromReader(strings.NewReader(ir.Content))
	if err != nil {
		Logger.Println(err)
	}
	// If the user wants to convert links to point to
	// the landing page, let's make it happen by changing up
	// e.HTML
	if ir.ConvertLinks {
		d, err := goquery.NewDocumentFromReader(bytes.NewReader(e.HTML))
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		d.Find("a").Each(func(i int, a *goquery.Selection) {
			a.SetAttr("href", "{{.URL}}")
		})
		h, err := d.Html()
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		e.HTML = []byte(h)
	}
	er := emailResponse{
		Subject: e.Subject,
		Text:    string(e.Text),
		HTML:    string(e.HTML),
	}
	JSONResponse(w, er, http.StatusOK)
	return
}

// API_Import_Site allows for the importing of HTML from a website
// Without "include_resources" set, it will merely place a "base" tag
// so that all resources can be loaded relative to the given URL.
func API_Import_Site(w http.ResponseWriter, r *http.Request) {
	cr := cloneRequest{}
	if r.Method != "POST" {
		JSONResponse(w, models.Response{Success: false, Message: "Method not allowed"}, http.StatusBadRequest)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&cr)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Error decoding JSON Request"}, http.StatusBadRequest)
		return
	}
	if err = cr.validate(); err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
		return
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(cr.URL)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
		return
	}
	// Insert the base href tag to better handle relative resources
	d, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
		return
	}
	// Assuming we don't want to include resources, we'll need a base href
	if d.Find("head base").Length() == 0 {
		d.Find("head").PrependHtml(fmt.Sprintf("<base href=\"%s\">", cr.URL))
	}
	forms := d.Find("form")
	forms.Each(func(i int, f *goquery.Selection) {
		// We'll want to store where we got the form from
		// (the current URL)
		url := f.AttrOr("action", cr.URL)
		if !strings.HasPrefix(url, "http") {
			url = fmt.Sprintf("%s%s", cr.URL, url)
		}
		f.PrependHtml(fmt.Sprintf("<input type=\"hidden\" name=\"__original_url\" value=\"%s\"/>", url))
	})
	h, err := d.Html()
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
		return
	}
	cs := cloneResponse{HTML: h}
	JSONResponse(w, cs, http.StatusOK)
	return
}

// API_Send_Test_Email sends a test email using the template name
// and Target given.
func API_Send_Test_Email(w http.ResponseWriter, r *http.Request) {
	s := &models.SendTestEmailRequest{}
	if r.Method != "POST" {
		JSONResponse(w, models.Response{Success: false, Message: "Method not allowed"}, http.StatusBadRequest)
		return
	}
	err := json.NewDecoder(r.Body).Decode(s)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Error decoding JSON Request"}, http.StatusBadRequest)
		return
	}
	// Validate the given request
	if err = s.Validate(); err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
		return
	}

	// If a Template is not specified use a default
	if s.Template.Name == "" {
		//default message body
		text := "It works!\n\nThis is an email letting you know that your gophish\nconfiguration was successful.\n" +
			"Here are the details:\n\nWho you sent from: {{.From}}\n\nWho you sent to: \n" +
			"{{if .FirstName}} First Name: {{.FirstName}}\n{{end}}" +
			"{{if .LastName}} Last Name: {{.LastName}}\n{{end}}" +
			"{{if .Position}} Position: {{.Position}}\n{{end}}" +
			"{{if .Department}} Department: {{.Department}}\n{{end}}" +
			"{{if .TrackingURL}} Tracking URL: {{.TrackingURL}}\n{{end}}" +
			"\nNow go send some phish!"
		t := models.Template{
			Subject: "Default Email from Gophish",
			Text:    text,
		}
		s.Template = t
		// Try to lookup the Template by name
	} else {
		// Get the Template requested by name
		s.Template, err = models.GetTemplateByName(s.Template.Name, ctx.Get(r, "user_id").(int64))
		if err == gorm.ErrRecordNotFound {
			Logger.Printf("Error - Template %s does not exist", s.Template.Name)
			JSONResponse(w, models.Response{Success: false, Message: models.ErrTemplateNotFound.Error()}, http.StatusBadRequest)
		} else if err != nil {
			Logger.Println(err)
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
	}

	// If a complete sending profile is provided use it
	if err := s.SMTP.Validate(); err != nil {
		// Otherwise get the SMTP requested by name
		s.SMTP, err = models.GetSMTPByName(s.SMTP.Name, ctx.Get(r, "user_id").(int64))
		if err == gorm.ErrRecordNotFound {
			Logger.Printf("Error - Sending profile %s does not exist", s.SMTP.Name)
			JSONResponse(w, models.Response{Success: false, Message: models.ErrSMTPNotFound.Error()}, http.StatusBadRequest)
		} else if err != nil {
			Logger.Println(err)
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
	}

	// Send the test email
	err = worker.SendTestEmail(s)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
		return
	}
	JSONResponse(w, models.Response{Success: true, Message: "Email Sent"}, http.StatusOK)
	return
}

// JSONResponse attempts to set the status code, c, and marshal the given interface, d, into a response that
// is written to the given ResponseWriter.
func JSONResponse(w http.ResponseWriter, d interface{}, c int) {
	dj, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		Logger.Println(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(c)
	fmt.Fprintf(w, "%s", dj)
}

type cloneRequest struct {
	URL              string `json:"url"`
	IncludeResources bool   `json:"include_resources"`
}

func (cr *cloneRequest) validate() error {
	if cr.URL == "" {
		return errors.New("No URL Specified")
	}
	return nil
}

type cloneResponse struct {
	HTML string `json:"html"`
}

type emailResponse struct {
	Text    string `json:"text"`
	HTML    string `json:"html"`
	Subject string `json:"subject"`
}
