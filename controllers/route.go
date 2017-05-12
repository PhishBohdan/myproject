package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/gophish/gophish/auth"
	"github.com/gophish/gophish/config"
	ctx "github.com/gophish/gophish/context"
	mid "github.com/gophish/gophish/middleware"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

// Logger is used to send logging messages to stdout.
var Logger = log.New(os.Stdout, " ", log.Ldate|log.Ltime|log.Lshortfile)

// CreateAdminRouter creates the routes for handling requests to the web interface.
// This function returns an http.Handler to be used in http.ListenAndServe().
func CreateAdminRouter() http.Handler {
	router := mux.NewRouter()
	// Base Front-end routes
	router.HandleFunc("/", Use(Base, mid.RequireLogin))
	router.HandleFunc("/login", Login)
	router.HandleFunc("/logout", Use(Logout, mid.RequireLogin))

	router.HandleFunc("/logo", Use(Logo, mid.RequireLogin))
	router.HandleFunc("/campaigns", Use(Campaigns, mid.RequireLogin))
	router.HandleFunc("/campaigns/{id:[0-9]+}", Use(CampaignID, mid.RequireLogin))

	router.HandleFunc("/filteredcampaigns", Use(FilteredCampaigns, mid.RequireLogin))
	router.HandleFunc("/filterresults", Use(FilterPhishingResults, mid.RequireLogin))
	router.HandleFunc("/target", Use(Target, mid.RequireLogin))
	router.HandleFunc("/pdfreport", Use(PdfReport, mid.RequireLogin))
	router.HandleFunc("/exportcsv", Use(ExportCsv, mid.RequireLogin))

	router.HandleFunc("/templates", Use(Templates, mid.RequireLogin))
	router.HandleFunc("/users", Use(Users, mid.RequireLogin))
	router.HandleFunc("/landing_pages", Use(LandingPages, mid.RequireLogin))
	router.HandleFunc("/sending_profiles", Use(SendingProfiles, mid.RequireLogin))
	router.HandleFunc("/register", Use(Register, mid.RequireLogin))
	router.HandleFunc("/addrestricteduser", Use(AddRestrictedUser, mid.RequireLogin))
	router.HandleFunc("/editrestricteduser", Use(EditRestrictedUser, mid.RequireLogin))
	router.HandleFunc("/restrictedusers", Use(RestrictedUsers, mid.RequireLogin))
	router.HandleFunc("/settings", Use(Settings, mid.RequireLogin))
	router.HandleFunc("/about", Use(About, mid.RequireLogin))
	router.HandleFunc("/fixresultsstatus", Use(FixResultsStatus, mid.RequireLogin))

	// Create the API routes
	api := router.PathPrefix("/api").Subrouter()
	api = api.StrictSlash(true)
	api.HandleFunc("/", Use(API, mid.RequireLogin))
	api.HandleFunc("/reset", Use(API_Reset, mid.RequireLogin))
	api.HandleFunc("/campaigns/", Use(API_Campaigns, mid.RequireAPIKey))
	api.HandleFunc("/campaignnames/{max:[0-9]+}", Use(API_CampaignNames, mid.RequireAPIKey))
	api.HandleFunc("/campaignslimit/{max:[0-9]+}", Use(API_CampaignNames, mid.RequireAPIKey))
	api.HandleFunc("/campaignsrange/{min:[0-9]+}-{max:[0-9]+}", Use(API_CampaignsRange, mid.RequireAPIKey))
	api.HandleFunc("/campaignssummarystats/", Use(API_SummaryStats, mid.RequireAPIKey))
	api.HandleFunc("/phishingresults/", Use(API_PhishingResults, mid.RequireAPIKey))
	api.HandleFunc("/targetevents/", Use(API_TargetEvents, mid.RequireAPIKey))
	api.HandleFunc("/campaigns/{id:[0-9]+}", Use(API_Campaigns_Id, mid.RequireAPIKey))
	api.HandleFunc("/campaigns/{id:[0-9]+}/results", Use(API_Campaigns_Id_Results, mid.RequireAPIKey))
	api.HandleFunc("/campaigns/{id:[0-9]+}/complete", Use(API_Campaigns_Id_Complete, mid.RequireAPIKey))
	api.HandleFunc("/groups/", Use(API_Groups, mid.RequireAPIKey))
	api.HandleFunc("/groups/{id:[0-9]+}", Use(API_Groups_Id, mid.RequireAPIKey))
	api.HandleFunc("/templates/", Use(API_Templates, mid.RequireAPIKey))
	api.HandleFunc("/templates/{id:[0-9]+}", Use(API_Templates_Id, mid.RequireAPIKey))
	api.HandleFunc("/pages/", Use(API_Pages, mid.RequireAPIKey))
	api.HandleFunc("/pages/{id:[0-9]+}", Use(API_Pages_Id, mid.RequireAPIKey))
	api.HandleFunc("/smtp/", Use(API_SMTP, mid.RequireAPIKey))
	api.HandleFunc("/smtp/{id:[0-9]+}", Use(API_SMTP_Id, mid.RequireAPIKey))
	api.HandleFunc("/util/send_test_email", Use(API_Send_Test_Email, mid.RequireAPIKey))
	api.HandleFunc("/import/group", API_Import_Group)
	api.HandleFunc("/import/email", API_Import_Email)
	api.HandleFunc("/import/site", API_Import_Site)

	// Setup static file serving
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	// Setup CSRF Protection
	csrfHandler := csrf.Protect([]byte(auth.GenerateSecureKey()),
		csrf.FieldName("csrf_token"),
		csrf.Secure(config.Conf.AdminConf.UseTLS))
	csrfRouter := csrfHandler(router)
	return Use(csrfRouter.ServeHTTP, mid.CSRFExceptions, mid.GetContext)
}

// CreatePhishingRouter creates the router that handles phishing connections.
func CreatePhishingRouter() http.Handler {
	router := mux.NewRouter()
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/endpoint/"))))
	router.HandleFunc("/track", PhishTracker)
	router.HandleFunc("/{path:.*}/track", PhishTracker)
	router.HandleFunc("/{path:.*}", PhishHandler)
	return router
}

// PhishTracker tracks emails as they are opened, updating the status for the given Result
func PhishTracker(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	// v := r.Header.Get("Content-Type")
	// if v != "" {
	// 	d, _, err := mime.ParseMediaType(v)
	// 	if err == nil && d == "multipart/form-data" {
	// 		r.ParseMultipartForm(32 << 20)
	// 	}
	// }

	id := r.Form.Get("rid")
	if id == "" {
		Logger.Println("Missing Result ID")
		http.NotFound(w, r)
		return
	}
	rs, err := models.GetResult(id)
	if err != nil {
		Logger.Println("No Results found")
		http.NotFound(w, r)
		return
	}
	c, err := models.GetCampaign(rs.CampaignId, rs.UserId)
	if err != nil {
		Logger.Println(err)
	}
	// Don't process events for completed campaigns
	if c.Status == models.CAMPAIGN_COMPLETE {
		http.NotFound(w, r)
		return
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		Logger.Println(err)
		return
	}
	// Respect X-Forwarded headers
	if fips := r.Header.Get("X-Forwarded-For"); fips != "" {
		ip = strings.Split(fips, ", ")[0]
	}
	// Handle post processing such as GeoIP
	err = rs.UpdateGeo(ip)
	if err != nil {
		Logger.Println(err)
	}
	d := struct {
		Payload url.Values        `json:"payload"`
		Browser map[string]string `json:"browser"`
	}{
		Payload: r.Form,
		Browser: make(map[string]string),
	}
	d.Browser["address"] = ip
	d.Browser["user-agent"] = r.Header.Get("User-Agent")
	rj, err := json.Marshal(d)
	if err != nil {
		Logger.Println(err)
		http.NotFound(w, r)
		return
	}
	c.AddEvent(models.Event{Email: rs.Email, Message: models.EVENT_OPENED, Details: string(rj)})
	// Don't update the status if the user already clicked the link
	// or submitted data to the campaign
	if rs.Status == models.STATUS_SUCCESS {
		http.ServeFile(w, r, "static/images/pixel.png")
		return
	}
	err = rs.UpdateStatus(models.EVENT_OPENED)
	if err != nil {
		Logger.Println(err)
	}
	http.ServeFile(w, r, "static/images/pixel.png")
}

// PhishHandler handles incoming client connections and registers the associated actions performed
// (such as clicked link, etc.)
func PhishHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		Logger.Println(err)
		http.NotFound(w, r)
		return
	}
	// v := r.Header.Get("Content-Type")
	// if v != "" {
	// 	d, _, err := mime.ParseMediaType(v)
	// 	if err == nil && d == "multipart/form-data" {
	// 		r.ParseMultipartForm(32 << 20)
	// 	}
	// }

	id := r.Form.Get("rid")
	if id == "" {
		http.NotFound(w, r)
		return
	}
	rs, err := models.GetResult(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	c, err := models.GetCampaign(rs.CampaignId, rs.UserId)
	if err != nil {
		Logger.Println(err)
	}
	// Don't process events for completed campaigns
	if c.Status == models.CAMPAIGN_COMPLETE {
		http.NotFound(w, r)
		return
	}
	if r.Method == "POST" {
		// Only set to success if user submitted data
		rs.UpdateStatus(models.STATUS_SUCCESS)
	}
	p, err := models.GetPage(c.PageId, c.UserId)
	if err != nil {
		Logger.Println(err)
	}
	d := struct {
		Payload url.Values        `json:"payload"`
		Browser map[string]string `json:"browser"`
	}{
		Payload: r.Form,
		Browser: make(map[string]string),
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		Logger.Println(err)
		return
	}
	// Respect X-Forwarded headers
	if fips := r.Header.Get("X-Forwarded-For"); fips != "" {
		ip = strings.Split(fips, ", ")[0]
	}
	// Handle post processing such as GeoIP
	err = rs.UpdateGeo(ip)
	if err != nil {
		Logger.Println(err)
	}
	d.Browser["address"] = ip
	d.Browser["user-agent"] = r.Header.Get("User-Agent")
	rj, err := json.Marshal(d)
	if err != nil {
		Logger.Println(err)
		http.NotFound(w, r)
		return
	}
	switch {
	case r.Method == "GET":
		err = c.AddEvent(models.Event{Email: rs.Email, Message: models.EVENT_CLICKED, Details: string(rj)})
		if err != nil {
			Logger.Println(err)
		}
	case r.Method == "POST":
		// If data was POST'ed, let's record it
		// Store the data in an event
		c.AddEvent(models.Event{Email: rs.Email, Message: models.EVENT_DATA_SUBMIT, Details: string(rj)})
		if err != nil {
			Logger.Println(err)
		}
		// Redirect to the desired page
		if p.RedirectURL != "" {
			http.Redirect(w, r, p.RedirectURL, 302)
			return
		}
	}
	var htmlBuff bytes.Buffer
	tmpl, err := template.New("html_template").Parse(p.HTML)
	if err != nil {
		Logger.Println(err)
		http.NotFound(w, r)
	}
	f, err := mail.ParseAddress(c.SMTP.FromAddress)
	if err != nil {
		Logger.Println(err)
	}
	fn := f.Name
	if fn == "" {
		fn = f.Address
	}
	rsf := struct {
		models.Result
		URL  string
		From string
	}{
		rs,
		c.URL + "?rid=" + rs.RId,
		fn,
	}
	err = tmpl.Execute(&htmlBuff, rsf)
	if err != nil {
		Logger.Println(err)
		http.NotFound(w, r)
	}
	w.Write(htmlBuff.Bytes())
}

// Use allows us to stack middleware to process the request
// Example taken from https://github.com/gorilla/mux/pull/36#issuecomment-25849172
func Use(handler http.HandlerFunc, mid ...func(http.Handler) http.HandlerFunc) http.HandlerFunc {
	for _, m := range mid {
		handler = m(handler)
	}
	return handler
}

// Register creates a new user
func Register(w http.ResponseWriter, r *http.Request) {
	// If it is a post request, attempt to register the account
	// Now that we are all registered, we can log the user in
	params := struct {
		Title   string
		Flashes []interface{}
		User    models.User
		Token   string
	}{Title: "Register", Token: csrf.Token(r)}
	session := ctx.Get(r, "session").(*sessions.Session)
	switch {
	case r.Method == "GET":
		params.Flashes = session.Flashes()
		session.Save(r, w)
		templates := template.New("template")
		_, err := templates.ParseFiles("templates/register.html", "templates/flashes.html")
		if err != nil {
			Logger.Println(err)
		}
		template.Must(templates, err).ExecuteTemplate(w, "base", params)
	case r.Method == "POST":
		//Attempt to register
		succ, err := auth.Register(r)
		//If we've registered, redirect to the login page
		if succ {
			session.AddFlash(models.Flash{
				Type:    "success",
				Message: "Registration successful!.",
			})
			session.Save(r, w)
			http.Redirect(w, r, "/login", 302)
			return
		}
		// Check the error
		m := err.Error()
		Logger.Println(err)
		session.AddFlash(models.Flash{
			Type:    "danger",
			Message: m,
		})
		session.Save(r, w)
		http.Redirect(w, r, "/register", 302)
		return
	}
}

// AddRestrictedUser creates a new restricteduser
func AddRestrictedUser(w http.ResponseWriter, r *http.Request) {
	// If it is a post request, attempt to register the account
	// Now that we are all registered, we can log the user in
	params := struct {
		Title          string
		Flashes        []interface{}
		RestrictedUser models.RestrictedUser
		Token          string
	}{Title: "Add Restricted User", Token: csrf.Token(r)}
	session := ctx.Get(r, "session").(*sessions.Session)
	user := ctx.Get(r, "user").(models.User)
	switch {
	case r.Method == "GET":
		params.Flashes = session.Flashes()
		session.Save(r, w)
		templates := template.New("template")
		_, err := templates.ParseFiles("templates/addrestricteduser.html", "templates/flashes.html")
		if err != nil {
			Logger.Println(err)
		}
		template.Must(templates, err).ExecuteTemplate(w, "base", params)
	case r.Method == "POST":
		//Attempt to register
		succ, err := auth.AddRestrictedUser(user, r)
		//If we've registered, redirect to the login page
		if succ {
			session.AddFlash(models.Flash{
				Type:    "success",
				Message: "Registration successful!.",
			})
			session.Save(r, w)
			http.Redirect(w, r, "/restrictedusers", 302)
			return
		}
		// Check the error
		m := err.Error()
		Logger.Println(err)
		session.AddFlash(models.Flash{
			Type:    "danger",
			Message: m,
		})
		session.Save(r, w)
		http.Redirect(w, r, "/addrestricteduser", 302)
		return
	}
}

// EditRestrictedUser edits an existing restricteduser
func EditRestrictedUser(w http.ResponseWriter, r *http.Request) {
	// If it is a post request, attempt to register the account
	// Now that we are all registered, we can log the user in
	restrictedUsernameId, err := strconv.Atoi(r.URL.Query()["usernameid"][0])
	if err != nil {
		fmt.Println("ERROR GETTING RESTRICTED USER IN EditRestrictedUser, err:", err)
		return
	}
	restrictedUser, err := models.GetRestrictedUser(int64(restrictedUsernameId))
	if err != nil && err != models.ErrUsernameTaken {
		fmt.Println("ERROR GETTING RESTRICTED USER IN EditRestrictedUser, err:", err)
		return
	}
	parentUser, err := restrictedUser.GetParentUser()
	if err != nil {
		fmt.Println("ERROR GETTING PARENT USER IN EditRestrictedUser, err:", err)
		return
	}
	parentUserCampaignNames, err := models.GetCampaignNames(parentUser.Id)
	if err != nil {
		fmt.Println("ERROR GETTING CAMPAIGN NAMES IN EditRestrictedUser, err:", err)
		return
	}
	campaignNames := []string{}
	for _, campaign := range restrictedUser.Campaigns {
		for _, parentCampaign := range parentUserCampaignNames {
			if campaign.CampaignId == parentCampaign.Id {
				campaignNames = append(campaignNames, parentCampaign.Name)
				break
			}
		}
	}
	params := struct {
		Title          string
		Flashes        []interface{}
		RestrictedUser models.RestrictedUser
		CampaignNames  []string
		Token          string
	}{Title: "Edit Restricted User", RestrictedUser: restrictedUser, CampaignNames: campaignNames, Token: csrf.Token(r)}
	session := ctx.Get(r, "session").(*sessions.Session)
	user := ctx.Get(r, "user").(models.User)
	switch {
	case r.Method == "GET":
		params.Flashes = session.Flashes()
		session.Save(r, w)
		templates := template.New("template")
		_, err := templates.ParseFiles("templates/editrestricteduser.html", "templates/flashes.html")
		if err != nil {
			Logger.Println(err)
		}
		template.Must(templates, err).ExecuteTemplate(w, "base", params)
	case r.Method == "POST":
		if r.FormValue("campaignemaildomain") != "" {
			//Attempt to update restricteduser
			succ, err := auth.UpdateRestrictedUserCampaigns(user, restrictedUser, r)
			//If we've registered, redirect to the login page
			if succ {
				session.AddFlash(models.Flash{
					Type:    "success",
					Message: "Saved.",
				})
				session.Save(r, w)
			}
			// Check the error
			if err != nil {
				m := err.Error()
				Logger.Println(err)
				session.AddFlash(models.Flash{
					Type:    "danger",
					Message: m,
				})
			}
			session.Save(r, w)
			http.Redirect(w, r, "/editrestricteduser?usernameid="+strconv.Itoa(int(restrictedUser.Id)), 302)
		} else if r.FormValue("campaignids") != "" {
			campaigns := r.FormValue("campaignids")
			ids := strings.Split(campaigns, ",")
			campaignIds := []models.RestrictedCampaign{}
			for _, id := range ids {
				campaignID, err := strconv.Atoi(id)
				if err != nil {
					fmt.Println(err)
					m := err.Error()
					Logger.Println(err)
					session.AddFlash(models.Flash{
						Type:    "danger",
						Message: m,
					})
					session.Save(r, w)
					http.Redirect(w, r, "/editrestricteduser?usernameid="+strconv.Itoa(int(restrictedUser.Id)), 302)
				}
				restrictedCampaign := models.RestrictedCampaign{CampaignId: int64(campaignID)}
				campaignIds = append(campaignIds, restrictedCampaign)
			}
			for _, campaignID := range campaignIds {
				restrictedUser.Campaigns = append(restrictedUser.Campaigns, campaignID)
			}
			err := models.PutRestrictedUser(&restrictedUser)
			if err != nil {
				m := err.Error()
				Logger.Println(err)
				session.AddFlash(models.Flash{
					Type:    "danger",
					Message: m,
				})
			} else {
				session.AddFlash(models.Flash{
					Type:    "success",
					Message: "Saved.",
				})
			}
			session.Save(r, w)
			http.Redirect(w, r, "/editrestricteduser?usernameid="+strconv.Itoa(int(restrictedUser.Id)), 302)
		} else if r.FormValue("new_password") != "" {
			newPassword := r.FormValue("new_password")
			confirmPassword := r.FormValue("confirm_new_password")
			if err := auth.SetRestrictedUserPassword(restrictedUser, newPassword, confirmPassword); err != nil {
				session.AddFlash(models.Flash{
					Type:    "danger",
					Message: err.Error(),
				})
				session.Save(r, w)
			} else {
				session.AddFlash(models.Flash{
					Type:    "success",
					Message: "Password Set.",
				})
				session.Save(r, w)
			}
			http.Redirect(w, r, "/editrestricteduser?usernameid="+strconv.Itoa(int(restrictedUser.Id)), 302)

		} else {
			session.AddFlash(models.Flash{
				Type:    "danger",
				Message: "Nothing entered",
			})
			session.Save(r, w)
			http.Redirect(w, r, "/editrestricteduser?usernameid="+strconv.Itoa(int(restrictedUser.Id)), 302)
		}
		return
	}
}

// RestrictedUsers lists all restrictedusers
func RestrictedUsers(w http.ResponseWriter, r *http.Request) {
	u := ctx.Get(r, "user").(models.User)
	restrictedUsers := models.GetRestrictedUsersByParent(u.Username)
	params := struct {
		User            models.User
		RestrictedUsers []models.RestrictedUser
		Title           string
		Flashes         []interface{}
		Token           string
	}{Title: "Restricted Users", User: u, RestrictedUsers: restrictedUsers, Token: csrf.Token(r)}
	getTemplate(w, "restrictedusers").ExecuteTemplate(w, "base", params)
}

// Base handles the default path and template execution
func Base(w http.ResponseWriter, r *http.Request) {
	FilteredCampaigns(w, r)
}

// Fix Results status
func FixResultsStatus(w http.ResponseWriter, r *http.Request) {
	allCampaigns, err := models.GetCampaigns(ctx.Get(r, "user").(models.User).Id, false)
	if err != nil {
		log.Fatal("FIXRESULTSSTATUS: Err fetching campaings")
		return
	}
	numUpdated := 0
	// If data was POST'ed, let's record it
	if r.Method == "POST" {
		for _, c := range allCampaigns {
			if err := c.GetDetails(true, false, false, false); err != nil {
				log.Printf("Couldn't update results for campaign \"%s\"\n", fmt.Sprintf("%s (#%d)", c.Name, c.Id))
				continue
			}
			for _, result := range c.Results {
				if result.Status == models.STATUS_SUCCESS {
					submittedData := false
					for _, e := range c.Events {
						if e.Email == result.Email && e.Message == "Submitted Data" {
							submittedData = true
							break
						}
					}
					if !submittedData {
						fmt.Println("Update result status for:", result)
						result.UpdateStatus(models.EVENT_CLICKED)
						numUpdated++
					}
				}
			}
		}
	}
	u := ctx.Get(r, "user").(models.User)
	restrictedUsers := models.GetRestrictedUsersByParent(u.Username)
	params := struct {
		User            models.User
		RestrictedUsers []models.RestrictedUser
		Title           string
		Flashes         []interface{}
		NumUpdated      int
		Posted          bool
		Token           string
	}{Title: "Fix Results Status", User: u, RestrictedUsers: restrictedUsers, Posted: r.Method == "POST", NumUpdated: numUpdated, Token: csrf.Token(r)}
	getTemplate(w, "fixresultsstatus").ExecuteTemplate(w, "base", params)
}

// Campaigns handles the default path and template execution
func Campaigns(w http.ResponseWriter, r *http.Request) {
	if ctx.Get(r, "user") != nil {
		params := struct {
			User    models.User
			Title   string
			Flashes []interface{}
			Token   string
		}{Title: "Campaigns", User: ctx.Get(r, "user").(models.User), Token: csrf.Token(r)}
		getTemplate(w, "campaigns").ExecuteTemplate(w, "base", params)
	} else {
		CampaignsReport(w, r)
	}
}

// CampaignID handles the default path and template execution
func CampaignID(w http.ResponseWriter, r *http.Request) {
	if ctx.Get(r, "user") != nil {
		params := struct {
			User    models.User
			Title   string
			Flashes []interface{}
			Token   string
		}{Title: "Campaign Results - Phishaway", User: ctx.Get(r, "user").(models.User), Token: csrf.Token(r)}
		getTemplate(w, "campaign_results").ExecuteTemplate(w, "base", params)
	} else {
		CampaignIDReport(w, r)
	}
}

func getUser(w http.ResponseWriter, r *http.Request) (models.User, error) {
	var user models.User
	var restrictedUser models.RestrictedUser
	var err error
	if ctx.Get(r, "user") != nil {
		user = ctx.Get(r, "user").(models.User)
	} else {
		restrictedUser = ctx.Get(r, "restricteduser").(models.RestrictedUser)
		user, err = restrictedUser.GetParentUser()
		if err != nil {
			fmt.Println("ERROR getUser:", err)
			return user, err
		}
	}
	return user, nil
}

func getRestrictedUser(w http.ResponseWriter, r *http.Request) (models.RestrictedUser, error) {
	var restrictedUser models.RestrictedUser
	isRestrictedUser := false
	var err error
	if ctx.Get(r, "user") == nil {
		restrictedUser = ctx.Get(r, "restricteduser").(models.RestrictedUser)
		_, err = restrictedUser.GetParentUser()
		isRestrictedUser = true
		if err != nil {
			fmt.Println("ERROR getRestrictedUser:", err)
			return models.RestrictedUser{}, err
		}
	}
	if !isRestrictedUser {
		err = errors.New("Expected restricted user, not full user")
		return models.RestrictedUser{}, err
	}
	return restrictedUser, nil
}

func isRestrictedUser(w http.ResponseWriter, r *http.Request) bool {
	var restrictedUser models.RestrictedUser
	isRestrictedUser := false
	var err error
	if ctx.Get(r, "user") == nil {
		restrictedUser = ctx.Get(r, "restricteduser").(models.RestrictedUser)
		_, err = restrictedUser.GetParentUser()
		isRestrictedUser = true
		if err != nil {
			fmt.Println("ERROR isRestrictedUser:", err)
			return false
		}
	}
	return isRestrictedUser
}

// FilteredCampaigns handles the default path and template execution
func FilteredCampaigns(w http.ResponseWriter, r *http.Request) {
	var user models.User
	var restrictedUser models.RestrictedUser
	isRestrictedUser := false
	var err error
	if ctx.Get(r, "user") != nil {
		user = ctx.Get(r, "user").(models.User)
	} else {
		restrictedUser = ctx.Get(r, "restricteduser").(models.RestrictedUser)
		user, err = restrictedUser.GetParentUser()
		isRestrictedUser = true
		if err != nil {
			fmt.Println("ERROR FilteredCampaigns:", err)
			return
		}
	}
	params := struct {
		User             models.User
		RestrictedUser   models.RestrictedUser
		IsRestrictedUser bool
		Title            string
		Flashes          []interface{}
		Token            string
	}{Title: "Campaigns", User: user, RestrictedUser: restrictedUser, IsRestrictedUser: isRestrictedUser, Token: csrf.Token(r)}
	if isRestrictedUser {
		getRestrictedUserTemplate(w, "filtered_campaigns_results").ExecuteTemplate(w, "restricteduserbase", params)
	} else {
		getTemplate(w, "filtered_campaigns_results").ExecuteTemplate(w, "base", params)
	}
}

// FilterPhishingResults handles the default path and template execution
func FilterPhishingResults(w http.ResponseWriter, r *http.Request) {
	var user models.User
	var err error
	var restrictedUser models.RestrictedUser
	isRestrictedUser := false

	if ctx.Get(r, "user") != nil {
		user = ctx.Get(r, "user").(models.User)
	} else {
		restrictedUser = ctx.Get(r, "restricteduser").(models.RestrictedUser)
		user, err = restrictedUser.GetParentUser()
		isRestrictedUser = true
		if err != nil {
			fmt.Println("ERROR FilterPhishingResults:", err)
			return
		}
	}
	params := struct {
		User             models.User
		RestrictedUser   models.RestrictedUser
		IsRestrictedUser bool
		Title            string
		Flashes          []interface{}
		Token            string
	}{Title: "Phishing Results",
		User:             user,
		RestrictedUser:   restrictedUser,
		IsRestrictedUser: isRestrictedUser,
		Token:            csrf.Token(r)}

	if isRestrictedUser {
		getRestrictedUserTemplate(w, "filterresults").ExecuteTemplate(w, "restricteduserbase", params)
	} else {
		getTemplate(w, "filterresults").ExecuteTemplate(w, "base", params)
	}
}

// Target handles the default path and template execution
func Target(w http.ResponseWriter, r *http.Request) {
	var user models.User
	var err error
	var restrictedUser models.RestrictedUser
	isRestrictedUser := false

	if ctx.Get(r, "user") != nil {
		user = ctx.Get(r, "user").(models.User)
	} else {
		restrictedUser = ctx.Get(r, "restricteduser").(models.RestrictedUser)
		user, err = restrictedUser.GetParentUser()
		isRestrictedUser = true
		if err != nil {
			fmt.Println("ERROR Target:", err)
			return
		}
	}
	var firstName, lastName, email string
	if len(r.URL.Query()["firstname"]) > 0 {
		firstName = r.URL.Query()["firstname"][0]
	}
	if len(r.URL.Query()["lastname"]) > 0 {
		lastName = r.URL.Query()["lastname"][0]
	}
	if len(r.URL.Query()["email"]) > 0 {
		email = r.URL.Query()["email"][0]
	}
	params := struct {
		User             models.User
		RestrictedUser   models.RestrictedUser
		IsRestrictedUser bool
		FirstName        string
		LastName         string
		Email            string
		Title            string
		Flashes          []interface{}
		Token            string
	}{Title: "Target",
		User:             user,
		RestrictedUser:   restrictedUser,
		IsRestrictedUser: isRestrictedUser,
		FirstName:        firstName,
		LastName:         lastName,
		Email:            email,
		Token:            csrf.Token(r)}

	if isRestrictedUser {
		getRestrictedUserTemplate(w, "target").ExecuteTemplate(w, "restricteduserbase", params)
	} else {
		getTemplate(w, "target").ExecuteTemplate(w, "base", params)
	}
}

// RestrictedUserBase handles the default path and template execution
func RestrictedUserBase(w http.ResponseWriter, r *http.Request) {
	restrictedUser := ctx.Get(r, "restricteduser").(models.RestrictedUser)
	user, err := restrictedUser.GetParentUser()
	if err != nil {
		fmt.Println("RestrictedUserBase:", err)
	}
	params := struct {
		User           models.User
		RestrictedUser models.RestrictedUser
		Title          string
		Flashes        []interface{}
		Token          string
	}{Title: "Dashboard", User: user, RestrictedUser: restrictedUser, Token: csrf.Token(r)}
	getRestrictedUserTemplate(w, "restricteduserdashboard").ExecuteTemplate(w, "restricteduserbase", params)
}

// PdfReport exports the campaign results to pdf.
func PdfReport(w http.ResponseWriter, r *http.Request) {
	var restrictedUser models.RestrictedUser
	var user models.User
	var err error
	if ctx.Get(r, "restricteduser") != nil {
		restrictedUser := ctx.Get(r, "restricteduser").(models.RestrictedUser)

		user, err = restrictedUser.GetParentUser()
	} else {
		user = ctx.Get(r, "user").(models.User)
	}
	if err != nil {
		fmt.Println("ExportPdf:", err)
		fmt.Println("restrictedUser.Parent:", restrictedUser.ParentUsername)
	} else {
		campaignScope := r.FormValue("scope")
		if campaignScope == "results" {
			models.PdfReport(w, r, user.Id)
		} else {
			models.ExportRawEvents(w, r, user.Id)
		}
	}
}

func ExportCsv(w http.ResponseWriter, r *http.Request) {
	var restrictedUser models.RestrictedUser
	var user models.User
	var err error
	if ctx.Get(r, "restricteduser") != nil {
		restrictedUser := ctx.Get(r, "restricteduser").(models.RestrictedUser)

		user, err = restrictedUser.GetParentUser()
	} else {
		user = ctx.Get(r, "user").(models.User)
	}
	if err != nil {
		fmt.Println("ExportPdf:", err)
		fmt.Println("restrictedUser.Parent:", restrictedUser.ParentUsername)
	} else {
		models.ExportCsv(w, r, user.Id)
	}
}

// CampaignsReport handles the default path and template execution
func CampaignsReport(w http.ResponseWriter, r *http.Request) {
	restrictedUser := ctx.Get(r, "restricteduser").(models.RestrictedUser)
	user, err := restrictedUser.GetParentUser()
	if err != nil {
		fmt.Println(err)
	}
	params := struct {
		User           models.User
		RestrictedUser models.RestrictedUser
		Title          string
		Flashes        []interface{}
		Token          string
	}{Title: "Campaigns", User: user, RestrictedUser: restrictedUser, Token: csrf.Token(r)}
	if err := getRestrictedUserTemplate(w, "restrictedusercampaigns").ExecuteTemplate(w, "restricteduserbase", params); err != nil {
		fmt.Println(err)
	}
}

// CampaignID handles the default path and template execution
func CampaignIDReport(w http.ResponseWriter, r *http.Request) {
	restrictedUser := ctx.Get(r, "restricteduser").(models.RestrictedUser)
	user, err := restrictedUser.GetParentUser()
	if err != nil {
		fmt.Println(err)
	}
	params := struct {
		User           models.User
		RestrictedUser models.RestrictedUser
		Title          string
		Flashes        []interface{}
		Token          string
	}{Title: "Campaign Results", User: user, RestrictedUser: restrictedUser, Token: csrf.Token(r)}
	if err := getRestrictedUserTemplate(w, "restrictedusercampaign_results").ExecuteTemplate(w, "restricteduserbase", params); err != nil {
		fmt.Println(err)
	}
}

// Templates handles the default path and template execution
func Templates(w http.ResponseWriter, r *http.Request) {
	// Example of using session - will be removed.
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
	}{Title: "Email Templates", User: ctx.Get(r, "user").(models.User), Token: csrf.Token(r)}
	getTemplate(w, "templates").ExecuteTemplate(w, "base", params)
}

// Users handles the default path and template execution
func Users(w http.ResponseWriter, r *http.Request) {
	// Example of using session - will be removed.
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
	}{Title: "Users & Groups", User: ctx.Get(r, "user").(models.User), Token: csrf.Token(r)}
	getTemplate(w, "users").ExecuteTemplate(w, "base", params)
}

// LandingPages handles the default path and template execution
func LandingPages(w http.ResponseWriter, r *http.Request) {
	// Example of using session - will be removed.
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
	}{Title: "Landing Pages", User: ctx.Get(r, "user").(models.User), Token: csrf.Token(r)}
	getTemplate(w, "landing_pages").ExecuteTemplate(w, "base", params)
}

// SendingProfiles handles the default path and template execution
func SendingProfiles(w http.ResponseWriter, r *http.Request) {
	// Example of using session - will be removed.
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
	}{Title: "Sending Profiles", User: ctx.Get(r, "user").(models.User), Token: csrf.Token(r)}
	getTemplate(w, "sending_profiles").ExecuteTemplate(w, "base", params)
}

// Logo is the handler for changing the logo
func Logo(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		params := struct {
			User    models.User
			Title   string
			Flashes []interface{}
			Token   string
			Version string
		}{Title: "Logo", Version: config.Version, User: ctx.Get(r, "user").(models.User), Token: csrf.Token(r)}
		getTemplate(w, "logo").ExecuteTemplate(w, "base", params)
	case r.Method == "POST":
		msg := models.Response{Success: true, Message: "Logo Updated Successfully"}
		var (
			status int
			err    error
		)

		defer func() {
			if nil != err {
				http.Error(w, err.Error(), status)
			}
		}()

		// parse request
		const _24MB = (1 << 20) * 24
		if err = r.ParseMultipartForm(_24MB); nil != err {
			status = http.StatusInternalServerError
			return
		}

		for _, fheaders := range r.MultipartForm.File {
			for _, hdr := range fheaders {
				// open uploaded
				var infile multipart.File
				if infile, err = hdr.Open(); nil != err {
					status = http.StatusInternalServerError
					return
				}

				// open destination
				var outfile *os.File
				if outfile, err = os.Create("./static/images/" + "logo_inv_small.png"); nil != err {
					fmt.Println("ERR:", err)
					status = http.StatusInternalServerError
					return
				}

				// 32K buffer copy
				var written int64
				if written, err = io.Copy(outfile, infile); nil != err {
					status = http.StatusInternalServerError
					return
				}
				fmt.Println("written:", written)
				// w.Write([]byte("uploaded file:" + hdr.Filename + ";length:" + strconv.Itoa(int(written))))
				params := struct {
					User    models.User
					Title   string
					Flashes []interface{}
					Token   string
					Version string
				}{Title: "Logo", Version: config.Version, User: ctx.Get(r, "user").(models.User), Token: csrf.Token(r)}
				getTemplate(w, "logo").ExecuteTemplate(w, "base", params)

			}
		}
		if err != nil {
			msg.Message = err.Error()
			msg.Success = false
			JSONResponse(w, msg, http.StatusBadRequest)
			return
		}
	}
}

// Settings handles the changing of settings
func Settings(w http.ResponseWriter, r *http.Request) {
	if ctx.Get(r, "user") != nil {
		switch {
		case r.Method == "GET":
			params := struct {
				User    models.User
				Title   string
				Flashes []interface{}
				Token   string
				Version string
			}{Title: "Settings", Version: config.Version, User: ctx.Get(r, "user").(models.User), Token: csrf.Token(r)}
			getTemplate(w, "settings").ExecuteTemplate(w, "base", params)
		case r.Method == "POST":
			err := auth.ChangePassword(r)
			msg := models.Response{Success: true, Message: "Settings Updated Successfully"}
			if err == auth.ErrInvalidPassword {
				msg.Message = "Invalid Password"
				msg.Success = false
				JSONResponse(w, msg, http.StatusBadRequest)
				return
			}
			if err != nil {
				msg.Message = err.Error()
				msg.Success = false
				JSONResponse(w, msg, http.StatusBadRequest)
				return
			}
			JSONResponse(w, msg, http.StatusOK)
		}
	} else {
		RestrictedUserSettings(w, r)
	}
}

// About handles the /about page
func About(w http.ResponseWriter, r *http.Request) {
	var user models.User
	var restrictedUser models.RestrictedUser
	var isRestricted bool
	user, err := getUser(w, r)
	if err != nil {
		fmt.Println("ERROR - About():", err)
		return
	}
	if isRestricted = isRestrictedUser(w, r); isRestricted {
		if restrictedUser, err = getRestrictedUser(w, r); err != nil {
			fmt.Println("ERROR - About():", err)
			return
		}
	}
	params := struct {
		User             models.User
		IsRestrictedUser bool
		RestrictedUser   models.RestrictedUser
		Title            string
		Flashes          []interface{}
		Token            string
		Version          string
	}{Title: "About",
		User:             user,
		RestrictedUser:   restrictedUser,
		IsRestrictedUser: isRestricted,
		Token:            csrf.Token(r)}
	if isRestricted {
		getRestrictedUserTemplate(w, "about").ExecuteTemplate(w, "restricteduserbase", params)
	} else {
		getTemplate(w, "about").ExecuteTemplate(w, "base", params)
	}
}

// RestrictedUserSettings handles the changing of settings
func RestrictedUserSettings(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		restrictedUser := ctx.Get(r, "restricteduser").(models.RestrictedUser)
		user, err := restrictedUser.GetParentUser()
		if err != nil {
			fmt.Println(err)
		}
		params := struct {
			User           models.User
			RestrictedUser models.RestrictedUser
			Title          string
			Flashes        []interface{}
			Token          string
			Version        string
		}{Title: "Settings", Version: config.Version, User: user, RestrictedUser: restrictedUser, Token: csrf.Token(r)}
		getRestrictedUserTemplate(w, "restrictedusersettings").ExecuteTemplate(w, "restricteduserbase", params)
	case r.Method == "POST":
		err := auth.RestrictedUserChangePassword(r)
		msg := models.Response{Success: true, Message: "Settings Updated Successfully"}
		if err == auth.ErrInvalidPassword {
			msg.Message = "Invalid Password"
			msg.Success = false
			JSONResponse(w, msg, http.StatusBadRequest)
			return
		}
		if err != nil {
			msg.Message = err.Error()
			msg.Success = false
			JSONResponse(w, msg, http.StatusBadRequest)
			return
		}
		JSONResponse(w, msg, http.StatusOK)
	}
}

// Login handles the authentication flow for a user. If credentials are valid,
// a session is created
func Login(w http.ResponseWriter, r *http.Request) {
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
	}{Title: "Login", Token: csrf.Token(r)}
	session := ctx.Get(r, "session").(*sessions.Session)
	switch {
	case r.Method == "GET":
		params.Flashes = session.Flashes()
		session.Save(r, w)
		templates := template.New("template")
		_, err := templates.ParseFiles("templates/login.html", "templates/flashes.html")
		if err != nil {
			Logger.Println(err)
		}
		template.Must(templates, err).ExecuteTemplate(w, "base", params)
	case r.Method == "POST":
		//Attempt to login
		succ, u, err := auth.Login(r)
		if err != nil {
			Logger.Println("User login: ", err)
			succ2, restrictedUser, err2 := auth.RestrictedUserLogin(r)
			if err2 != nil {
				Logger.Println("Restricted user login:", err2)
			}
			if succ2 {
				session.Values["restricteduserid"] = restrictedUser.Id
				session.Save(r, w)
				http.Redirect(w, r, "/", 302)
			} else {
				Logger.Println("Restricted user login: success = false")
			}

		}
		//If we've logged in, save the session and redirect to the dashboard
		if succ {
			session.Values["id"] = u.Id
			session.Save(r, w)
			http.Redirect(w, r, "/", 302)
		} else {
			Flash(w, r, "danger", "Invalid Username/Password")
			http.Redirect(w, r, "/login", 302)
		}
	}
}

// Logout destroys the current user session
func Logout(w http.ResponseWriter, r *http.Request) {
	// If it is a post request, attempt to register the account
	// Now that we are all registered, we can log the user in
	session := ctx.Get(r, "session").(*sessions.Session)
	delete(session.Values, "id")

	// If logged in as a restricteduser then logout as a restricteduser
	delete(session.Values, "restricteduserid")

	Flash(w, r, "success", "You have successfully logged out")
	http.Redirect(w, r, "/login", 302)
}

// Preview allows for the viewing of page html in a separate browser window
func Preview(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "%s", r.FormValue("html"))
}

// Clone takes a URL as a POST parameter and returns the site HTML
func Clone(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusBadRequest)
		return
	}
	if url, ok := vars["url"]; ok {
		Logger.Println(url)
	}
	http.Error(w, "No URL given.", http.StatusBadRequest)
}

func getTemplate(w http.ResponseWriter, tmpl string) *template.Template {
	templates := template.New("template")
	_, err := templates.ParseFiles("templates/base.html", "templates/"+tmpl+".html", "templates/flashes.html")
	if err != nil {
		Logger.Println(err)
	}
	return template.Must(templates, err)
}

func getRestrictedUserTemplate(w http.ResponseWriter, tmpl string) *template.Template {
	templates := template.New("template")
	_, err := templates.ParseFiles("templates/restricteduserbase.html", "templates/"+tmpl+".html", "templates/flashes.html")
	if err != nil {
		Logger.Println("getRestrictedUserTemplate", err)
	}
	return template.Must(templates, err)
}

// Flash handles the rendering flash messages
func Flash(w http.ResponseWriter, r *http.Request, t string, m string) {
	session := ctx.Get(r, "session").(*sessions.Session)
	session.AddFlash(models.Flash{
		Type:    t,
		Message: m,
	})
	session.Save(r, w)
}
