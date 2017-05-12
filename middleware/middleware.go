package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gophish/gophish/auth"
	ctx "github.com/gophish/gophish/context"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/csrf"
)

var CSRFExemptPrefixes = []string{
	"/api",
}

func CSRFExceptions(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, prefix := range CSRFExemptPrefixes {
			if strings.HasPrefix(r.URL.Path, prefix) {
				r = csrf.UnsafeSkipCheck(r)
				break
			}
		}
		handler.ServeHTTP(w, r)
	}
}

// GetContext wraps each request in a function which fills in the context for a given request.
// This includes setting the User and Session keys and values as necessary for use in later functions.
func GetContext(handler http.Handler) http.HandlerFunc {
	// Set the context here
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the request form
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing request", http.StatusInternalServerError)
		}
		// v := r.Header.Get("Content-Type")
		// if v != "" {
		// 	d, _, err := mime.ParseMediaType(v)
		// 	if err == nil && d == "multipart/form-data" {
		// 		r.ParseMultipartForm(32 << 20)
		// 	}
		// }

		// Set the context appropriately here.
		// Set the session
		session, _ := auth.Store.Get(r, "gophish")
		// Put the session in the context so that we can
		// reuse the values in different handlers
		r = ctx.Set(r, "session", session)
		if id, ok := session.Values["id"]; ok {
			u, err := models.GetUser(id.(int64))
			if err != nil {
				r = ctx.Set(r, "user", nil)
			} else {
				r = ctx.Set(r, "user", u)
			}
		} else {
			r = ctx.Set(r, "user", nil)
		}
		if restricteduserId, ok := session.Values["restricteduserid"]; ok {
			u, err := models.GetRestrictedUser(restricteduserId.(int64))
			if err != nil {
				r = ctx.Set(r, "restricteduser", nil)
			} else {
				r = ctx.Set(r, "restricteduser", u)
			}
		} else {
			r = ctx.Set(r, "restricteduser", nil)
		}
		handler.ServeHTTP(w, r)
		// Remove context contents
		ctx.Clear(r)
	}
}

func RequireAPIKey(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		// v := r.Header.Get("Content-Type")
		// if v != "" {
		// 	d, _, err := mime.ParseMediaType(v)
		// 	if err == nil && d == "multipart/form-data" {
		// 		r.ParseMultipartForm(32 << 20)
		// 	}
		// }
		ak := r.Form.Get("api_key")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			w.Header().Set("Access-Control-Max-Age", "1000")
			w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
			return
		}
		if ak == "" {
			JSONError(w, 400, "API Key not set")
		} else {
			u, err := models.GetUserByAPIKey(ak)
			if err != nil {
				JSONError(w, 400, "Invalid API Key")
				return
			}
			r = ctx.Set(r, "user_id", u.Id)
			r = ctx.Set(r, "api_key", ak)
			handler.ServeHTTP(w, r)
		}
	}
}

// RequireLogin is a simple middleware which checks to see if the user is currently logged in.
// If not, the function returns a 302 redirect to the login page.
func RequireLogin(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := ctx.Get(r, "user")
		restrictedUser := ctx.Get(r, "restricteduser")
		if u != nil || restrictedUser != nil {
			handler.ServeHTTP(w, r)
		} else {
			http.Redirect(w, r, "/login", 302)
		}
	}
}

// RequireRestrictedUserLogin is a simple middleware which checks to see if the user is currently logged in.
// If not, the function returns a 302 redirect to the login page.
func RequireRestrictedUserLogin(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if u := ctx.Get(r, "restricteduser"); u != nil {
			handler.ServeHTTP(w, r)
		} else {
			http.Redirect(w, r, "/login", 302)
		}
	}
}

// JSONError returns an error in JSON format with the given
// status code and message
func JSONError(w http.ResponseWriter, c int, m string) {
	w.WriteHeader(c)
	w.Header().Set("Content-Type", "application/json")
	cj, _ := json.MarshalIndent(models.Response{Success: false, Message: m}, "", "  ")
	fmt.Fprintf(w, "%s", cj)
}
