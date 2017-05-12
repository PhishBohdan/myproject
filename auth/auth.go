package auth

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"crypto/rand"

	ctx "github.com/gophish/gophish/context"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

//init registers the necessary models to be saved in the session later
func init() {
	gob.Register(&models.User{})
	gob.Register(&models.Flash{})
	Store.Options.HttpOnly = true
	// This sets the maxAge to 5 days for all cookies
	Store.MaxAge(86400 * 5)
}

// Store contains the session information for the request
var Store = sessions.NewCookieStore(
	[]byte(securecookie.GenerateRandomKey(64)), //Signing key
	[]byte(securecookie.GenerateRandomKey(32)))

// ErrInvalidPassword is thrown when a user provides an incorrect password.
var ErrInvalidPassword = errors.New("Invalid Password")

// ErrEmptyPassword is thrown when a user provides a blank password to the register
// or change password functions
var ErrEmptyPassword = errors.New("Password cannot be blank")

// ErrPasswordMismatch is thrown when a user provides passwords that do not match
var ErrPasswordMismatch = errors.New("Passwords must match")

// Login attempts to login the user given a request.
func Login(r *http.Request) (bool, models.User, error) {
	username, password := r.FormValue("username"), r.FormValue("password")
	u, err := models.GetUserByUsername(username)
	if err != nil && err != models.ErrUsernameTaken {
		return false, models.User{}, err
	}
	//If we've made it here, we should have a valid user stored in u
	//Let's check the password
	err = bcrypt.CompareHashAndPassword([]byte(u.Hash), []byte(password))
	if err != nil {
		return false, models.User{}, ErrInvalidPassword
	}
	return true, u, nil
}

// RestrictedUserLogin attempts to login the user given a request.
func RestrictedUserLogin(r *http.Request) (bool, models.RestrictedUser, error) {
	username, password := r.FormValue("username"), r.FormValue("password")
	u, err := models.GetRestrictedUserByUsername(username)
	if err != nil && err != models.ErrUsernameTaken {
		return false, models.RestrictedUser{}, err
	}
	//If we've made it here, we should have a valid user stored in u
	//Let's check the password
	err = bcrypt.CompareHashAndPassword([]byte(u.Hash), []byte(password))
	if err != nil {
		return false, models.RestrictedUser{}, ErrInvalidPassword
	}
	return true, u, nil
}

// Register attempts to register the user given a request.
func Register(r *http.Request) (bool, error) {
	username := r.FormValue("username")
	newPassword := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")
	u, err := models.GetUserByUsername(username)
	// If we have an error which is not simply indicating that no user was found, report it
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	u = models.User{}
	// If we've made it here, we should have a valid username given
	// Check that the passsword isn't blank
	if newPassword == "" {
		return false, ErrEmptyPassword
	}
	// Make sure passwords match
	if newPassword != confirmPassword {
		return false, ErrPasswordMismatch
	}
	// Let's create the password hash
	h, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return false, err
	}
	u.Username = username
	u.Hash = string(h)
	u.ApiKey = GenerateSecureKey()
	err = models.PutUser(&u)
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	return true, nil
}

// AddRestrictedUser attempts to add the restricteduser given a request.
func AddRestrictedUser(parent models.User, r *http.Request) (bool, error) {
	username := r.FormValue("username")
	parentUsername := parent.Username
	newPassword := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")
	u, err := models.GetRestrictedUserByUsername(username)
	// If we have an error which is not simply indicating that no user was found, report it
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	u = models.RestrictedUser{}
	// If we've made it here, we should have a valid username given
	// Check that the passsword isn't blank
	if newPassword == "" {
		return false, ErrEmptyPassword
	}
	// Make sure passwords match
	if newPassword != confirmPassword {
		return false, ErrPasswordMismatch
	}
	// Let's create the password hash
	h, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return false, err
	}
	u.Username = username
	u.Campaigns = []models.RestrictedCampaign{}
	u.ParentUsername = parentUsername
	u.Hash = string(h)
	err = models.PutRestrictedUser(&u)
	if err != nil {
		fmt.Println("ERROR ADDRESTRICTEDUSER:", err)
		return false, err
	}
	return true, nil
}

// UpdateRestrictedUserCampaigns attempts to update the restricteduser given a request.
func UpdateRestrictedUserCampaigns(parent models.User, ruser models.RestrictedUser, r *http.Request) (bool, error) {
	u := ruser
	emailDomain := r.FormValue("campaignemaildomain")
	campaignIds := []models.RestrictedCampaign{}
	campaigns, err := models.GetCampaigns(parent.Id, false)
	if err != nil {
		fmt.Println("ERROR UPDATERESTRICTEDUSER:", err)
		return false, err
	}
	for _, c := range campaigns {
		if err := c.GetDetails(false, false, false, false); err != nil {
			fmt.Println("ERROR UPDATERESTRICTEDUSER:", err)
			return false, err
		}
		add := false
		for _, r := range c.Results {
			if strings.Contains(r.Email, emailDomain) {
				add = true
				break
			}
		}
		if add {
			restrictedCampaign := models.RestrictedCampaign{RestrictedUserID: int(u.Id), CampaignId: c.Id}
			campaignIds = append(campaignIds, restrictedCampaign)
		}
	}
	u.Campaigns = campaignIds
	err = models.PutRestrictedUser(&u)
	if err != nil {
		fmt.Println("ERROR UPDATERESTRICTEDUSER:", err)
		return false, err
	}
	return true, nil
}

// GenerateSecureKey creates a secure key to use
// as an API key
func GenerateSecureKey() string {
	// Inspired from gorilla/securecookie
	k := make([]byte, 32)
	io.ReadFull(rand.Reader, k)
	return fmt.Sprintf("%x", k)
}

func ChangePassword(r *http.Request) error {
	u := ctx.Get(r, "user").(models.User)
	currentPw := r.FormValue("current_password")
	newPassword := r.FormValue("new_password")
	confirmPassword := r.FormValue("confirm_new_password")
	// Check the current password
	err := bcrypt.CompareHashAndPassword([]byte(u.Hash), []byte(currentPw))
	if err != nil {
		return ErrInvalidPassword
	}
	// Check that the new password isn't blank
	if newPassword == "" {
		return ErrEmptyPassword
	}
	// Check that new passwords match
	if newPassword != confirmPassword {
		return ErrPasswordMismatch
	}
	// Generate the new hash
	h, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Hash = string(h)
	if err = models.PutUser(&u); err != nil {
		return err
	}
	return nil
}

func RestrictedUserChangePassword(r *http.Request) error {
	restrictedUser := ctx.Get(r, "restricteduser").(models.RestrictedUser)
	currentPw := r.FormValue("current_password")
	newPassword := r.FormValue("new_password")
	confirmPassword := r.FormValue("confirm_new_password")
	// Check the current password
	err := bcrypt.CompareHashAndPassword([]byte(restrictedUser.Hash), []byte(currentPw))
	if err != nil {
		return ErrInvalidPassword
	}
	// Check that the new password isn't blank
	if newPassword == "" {
		return ErrEmptyPassword
	}
	// Check that new passwords match
	if newPassword != confirmPassword {
		return ErrPasswordMismatch
	}
	// Generate the new hash
	h, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	restrictedUser.Hash = string(h)
	if err = models.PutRestrictedUser(&restrictedUser); err != nil {
		return err
	}
	return nil
}

func SetRestrictedUserPassword(restrictedUser models.RestrictedUser, newPassword, confirmPassword string) error {
	// Check that the new password isn't blank
	if newPassword == "" {
		return ErrEmptyPassword
	}
	// Check that new passwords match
	if newPassword != confirmPassword {
		return ErrPasswordMismatch
	}
	// Generate the new hash
	h, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	restrictedUser.Hash = string(h)
	if err = models.PutRestrictedUser(&restrictedUser); err != nil {
		return err
	}
	return nil
}
