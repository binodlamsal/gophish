package controllers

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/binodlamsal/gophish/auth"
	"github.com/binodlamsal/gophish/bakery"
	"github.com/binodlamsal/gophish/config"
	ctx "github.com/binodlamsal/gophish/context"
	log "github.com/binodlamsal/gophish/logger"
	mid "github.com/binodlamsal/gophish/middleware"
	"github.com/binodlamsal/gophish/models"
	"github.com/binodlamsal/gophish/util"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

func init() {
	bakery.SetKey(os.Getenv("SSO_KEY"))
}

// CreateAdminRouter creates the routes for handling requests to the web interface.
// This function returns an http.Handler to be used in http.ListenAndServe().
func CreateAdminRouter() http.Handler {
	router := mux.NewRouter()
	// Base Front-end routes
	router.HandleFunc("/", Use(Base, mid.RequireLogin, mid.SSO))
	router.HandleFunc("/login", SSO_Login)
	router.HandleFunc("/bakery/login", SSO_Login)
	// router.HandleFunc("/sso/mock", SSO_Mock)
	router.HandleFunc("/logout", Use(Logout))
	router.HandleFunc("/campaigns", Use(Campaigns, mid.RequireLogin, mid.SSO))
	router.HandleFunc("/campaigns/{id:[0-9]+}", Use(CampaignID, mid.RequireLogin, mid.SSO))
	router.HandleFunc("/templates", Use(Templates, mid.RequireLogin, mid.SSO))
	router.HandleFunc("/users", Use(Users, mid.RequireLogin, mid.SSO))
	router.HandleFunc("/landing_pages", Use(LandingPages, mid.RequireLogin, mid.SSO))
	router.HandleFunc("/sending_profiles", Use(SendingProfiles, mid.RequireRoles([]int64{models.Administrator}), mid.RequireLogin, mid.SSO))
	router.HandleFunc("/our_domains", Use(SendingDomains, mid.RequireRoles([]int64{models.Administrator}), mid.RequireLogin, mid.SSO))
	router.HandleFunc("/categories", Use(PhishingCategories, mid.RequireRoles([]int64{models.Administrator}), mid.RequireLogin, mid.SSO))
	router.HandleFunc("/register", Use(Register, mid.RequireRoles([]int64{models.Administrator, models.Partner, models.ChildUser}), mid.RequireLogin, mid.SSO))
	router.HandleFunc("/settings", Use(Settings, mid.RequireLogin, mid.SSO))
	router.HandleFunc("/people", Use(People, mid.RequireRoles([]int64{models.Administrator, models.Partner, models.ChildUser}), mid.RequireLogin, mid.SSO))
	router.HandleFunc("/roles", Use(Roles, mid.RequireRoles([]int64{models.Administrator}), mid.RequireLogin, mid.SSO))
	router.HandleFunc("/logo", Use(Logo))
	router.HandleFunc("/avatars/{id:[0-9]+}", Use(Avatars_Id))
	// Create the API routes
	api := router.PathPrefix("/api").Subrouter()
	api = api.StrictSlash(true)
	api.HandleFunc("/reset", Use(API_Reset, mid.RequireAPIKey))
	api.HandleFunc("/campaigns/", Use(API_Campaigns, mid.RequireAPIKey))
	api.HandleFunc("/people", Use(API_Users, mid.RequireAPIKey))
	api.HandleFunc("/signupapi", Use(API_Users, mid.RequireAPIKey))
	api.HandleFunc("/people/partner", Use(API_User_Partners, mid.RequireAPIKey))
	api.HandleFunc("/roles", Use(API_Roles, mid.RequireRoles([]int64{models.Administrator, models.Partner}), mid.RequireAPIKey))
	api.HandleFunc("/roles/{id:[0-9]+}", Use(API_Roles_Id, mid.RequireAPIKey))
	api.HandleFunc("/people/{id:[0-9]+}", Use(API_Users_Id, mid.RequireAPIKey))
	api.HandleFunc("/phishtags/", Use(API_Tags, mid.RequireAPIKey))
	api.HandleFunc("/phishtagssingle/{id:[0-9]+}", Use(API_Tags_Single, mid.RequireAPIKey))
	api.HandleFunc("/campaigns/summary", Use(API_Campaigns_Summary, mid.RequireAPIKey))
	api.HandleFunc("/campaigns/{id:[0-9]+}", Use(API_Campaigns_Id, mid.RequireAPIKey))
	api.HandleFunc("/campaigns/{id:[0-9]+}/results", Use(API_Campaigns_Id_Results, mid.RequireAPIKey))
	api.HandleFunc("/campaigns/{id:[0-9]+}/summary", Use(API_Campaign_Id_Summary, mid.RequireAPIKey))
	api.HandleFunc("/campaigns/{id:[0-9]+}/complete", Use(API_Campaigns_Id_Complete, mid.RequireAPIKey))
	api.HandleFunc("/groups/", Use(API_Groups, mid.RequireAPIKey))
	api.HandleFunc("/groups/summary", Use(API_Groups_Summary, mid.RequireAPIKey))
	api.HandleFunc("/groups/{id:[0-9]+}", Use(API_Groups_Id, mid.RequireAPIKey))
	api.HandleFunc("/groups/{id:[0-9]+}/summary", Use(API_Groups_Id_Summary, mid.RequireAPIKey))
	api.HandleFunc("/groups/{id:[0-9]+}/lms_users", Use(API_Groups_Id_LMS, mid.RequireAPIKey))
	api.HandleFunc(`/groups/{id:[0-9]+}/lms_users/jobs/{jid:[a-f0-9\-]{36}}`, Use(API_Groups_Id_LMS_Jobs_Id, mid.RequireAPIKey))
	api.HandleFunc("/templates/", Use(API_Templates, mid.RequireAPIKey))
	api.HandleFunc("/templates/{id:[0-9]+}", Use(API_Templates_Id, mid.RequireAPIKey))
	api.HandleFunc("/pages/", Use(API_Pages, mid.RequireAPIKey))
	api.HandleFunc("/pages/{id:[0-9]+}", Use(API_Pages_Id, mid.RequireAPIKey))
	api.HandleFunc("/plans/", Use(API_Plans, mid.RequireRoles([]int64{models.Administrator, models.Partner}), mid.RequireAPIKey))
	api.HandleFunc("/subscriptions/", Use(API_Subscriptions, mid.RequireRoles([]int64{models.Administrator}), mid.RequireAPIKey))
	api.HandleFunc("/smtp/", Use(API_SMTP, mid.RequireAPIKey))
	api.HandleFunc("/sendingdomains", Use(API_SMTP_domains, mid.RequireAPIKey))
	api.HandleFunc("/smtp/{id:[0-9]+}", Use(API_SMTP_Id, mid.RequireRoles([]int64{models.Administrator}), mid.RequireAPIKey))
	api.HandleFunc("/util/send_test_email", Use(API_Send_Test_Email, mid.RequireAPIKey))
	api.HandleFunc("/import/group", Use(API_Import_Group, mid.RequireAPIKey))
	api.HandleFunc("/import/email", Use(API_Import_Email, mid.RequireAPIKey))
	api.HandleFunc("/import/site", Use(API_Import_Site, mid.RequireAPIKey))

	// Setup static file serving
	router.PathPrefix("/").Handler(http.FileServer(UnindexedFileSystem{http.Dir("./static/")}))

	// Setup CSRF Protection
	csrfHandler := csrf.Protect([]byte(util.GenerateSecureKey()),
		csrf.FieldName("csrf_token"),
		csrf.Secure(config.Conf.AdminConf.UseTLS))
	csrfRouter := csrfHandler(router)
	return Use(csrfRouter.ServeHTTP, mid.CSRFExceptions, mid.GetContext)
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
		Roles   models.Roles
		Admin   bool
		Token   string
	}{Title: "Register", Admin: false, Token: csrf.Token(r)}

	session := ctx.Get(r, "session").(*sessions.Session)

	switch {
	case r.Method == "GET":
		uid := ctx.Get(r, "user").(models.User).Id
		role, err := models.GetUserRole(uid)

		if err != nil {
			log.Error(err)
		}

		if role.Is(models.Administrator) {
			params.Admin = true
		}

		roles, err := models.GetRoles()

		if err != nil {
			log.Error(err)
		}

		params.Flashes = session.Flashes()
		params.Roles = roles.AvailableFor(role)
		session.Save(r, w)
		templates := template.New("template")

		_, errs := templates.ParseFiles("templates/register.html", "templates/flashes.html")
		if errs != nil {
			log.Error(errs)
		}
		template.Must(templates, errs).ExecuteTemplate(w, "base", params)
	case r.Method == "POST":
		//Attempt to register
		succ, err := auth.Register(r)
		//If we've registered, redirect to the login page
		if succ {
			Flash(w, r, "success", "Registration successful!")
			session.Save(r, w)
			http.Redirect(w, r, "/people", 302)
			return
		}
		// Check the error
		m := err.Error()
		log.Error(err)
		Flash(w, r, "danger", m)
		session.Save(r, w)
		http.Redirect(w, r, "/register", 302)
		return
	}
}

// Base handles the default path and template execution
func Base(w http.ResponseWriter, r *http.Request) {
	params := struct {
		User    models.User
		Role    string
		Title   string
		Flashes []interface{}
		Token   string
	}{Title: "Dashboard", User: ctx.Get(r, "user").(models.User), Role: "", Token: csrf.Token(r)}
	role, err := models.GetUserRole(params.User.Id)

	if err != nil {
		log.Error(err)
	}

	params.Role = role.Name()
	getTemplate(r, w, "dashboard").ExecuteTemplate(w, "base", params)
}

// Campaigns handles the default path and template execution
func Campaigns(w http.ResponseWriter, r *http.Request) {
	// Example of using session - will be removed.
	params := struct {
		User    models.User
		Role    string
		Title   string
		Flashes []interface{}
		Token   string
	}{
		Title: "Campaigns",
		User:  ctx.Get(r, "user").(models.User),
		Role:  "",
		Token: csrf.Token(r),
	}

	role, err := models.GetUserRole(params.User.Id)

	if err != nil {
		log.Error(err)
	}

	params.Role = role.Name()
	getTemplate(r, w, "campaigns").ExecuteTemplate(w, "base", params)
}

// People handles the default path and template execution
func People(w http.ResponseWriter, r *http.Request) {
	// Example of using session - will be removed.
	params := struct {
		User                   models.User
		Role                   string
		Title                  string
		Flashes                []interface{}
		Token                  string
		CanManageSubscriptions bool
	}{
		Title: "People",
		User:  ctx.Get(r, "user").(models.User),
		Role:  "", Token: csrf.Token(r),
		CanManageSubscriptions: false,
	}

	role, err := models.GetUserRole(params.User.Id)

	if err != nil {
		log.Error(err)
	}

	params.Role = role.Name()
	params.CanManageSubscriptions = params.User.CanManageSubscriptions()
	getTemplate(r, w, "people").ExecuteTemplate(w, "base", params)
}

// Roles handles the default path and template execution
func Roles(w http.ResponseWriter, r *http.Request) {
	// Example of using session - will be removed.
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
	}{Title: "Roles", User: ctx.Get(r, "user").(models.User), Token: csrf.Token(r)}
	getTemplate(r, w, "roles").ExecuteTemplate(w, "base", params)
}

// CampaignID handles the default path and template execution
func CampaignID(w http.ResponseWriter, r *http.Request) {
	// Example of using session - will be removed.
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
	}{Title: "Campaign Results", User: ctx.Get(r, "user").(models.User), Token: csrf.Token(r)}
	getTemplate(r, w, "campaign_results").ExecuteTemplate(w, "base", params)
}

// Templates handles the default path and template execution
func Templates(w http.ResponseWriter, r *http.Request) {
	// Example of using session - will be removed.
	params := struct {
		User    models.User
		Role    string
		Title   string
		Flashes []interface{}
		Token   string
	}{
		Title: "Email Templates",
		User:  ctx.Get(r, "user").(models.User),
		Role:  "",
		Token: csrf.Token(r),
	}

	role, err := models.GetUserRole(params.User.Id)

	if err != nil {
		log.Error(err)
	}

	params.Role = role.Name()
	getTemplate(r, w, "templates").ExecuteTemplate(w, "base", params)
}

// Users handles the default path and template execution
func Users(w http.ResponseWriter, r *http.Request) {
	// Example of using session - will be removed.
	params := struct {
		User    models.User
		Role    string
		Title   string
		Flashes []interface{}
		Token   string
	}{
		Title: "Users & Groups",
		User:  ctx.Get(r, "user").(models.User),
		Role:  "",
		Token: csrf.Token(r),
	}

	role, err := models.GetUserRole(params.User.Id)

	if err != nil {
		log.Error(err)
	}

	params.Role = role.Name()
	getTemplate(r, w, "users").ExecuteTemplate(w, "base", params)
}

// LandingPages handles the default path and template execution
func LandingPages(w http.ResponseWriter, r *http.Request) {
	// Example of using session - will be removed.
	params := struct {
		User    models.User
		Role    string
		Title   string
		Flashes []interface{}
		Token   string
	}{
		Title: "Landing Pages",
		User:  ctx.Get(r, "user").(models.User),
		Role:  "",
		Token: csrf.Token(r),
	}

	role, err := models.GetUserRole(params.User.Id)

	if err != nil {
		log.Error(err)
	}

	params.Role = role.Name()
	getTemplate(r, w, "landing_pages").ExecuteTemplate(w, "base", params)
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
	getTemplate(r, w, "sending_profiles").ExecuteTemplate(w, "base", params)
}

// Replancememnt of SendingProfiles by sendingdomains in our application a nornal user can use the profile/domains created by
// the administrator handles the default path and template execution
func SendingDomains(w http.ResponseWriter, r *http.Request) {
	// Example of using session - will be removed.
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
	}{Title: "Sending Domains", User: ctx.Get(r, "user").(models.User), Token: csrf.Token(r)}
	getTemplate(r, w, "sending_domains").ExecuteTemplate(w, "base", params)
}

// Replancememnt of SendingProfiles by sendingdomains in our application a nornal user can use the profile/domains created by
// the administrator handles the default path and template execution
func PhishingCategories(w http.ResponseWriter, r *http.Request) {
	// Example of using session - will be removed.
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
	}{Title: "Phishing Categories", User: ctx.Get(r, "user").(models.User), Token: csrf.Token(r)}
	getTemplate(r, w, "phishing_categories").ExecuteTemplate(w, "base", params)
}

// Settings handles the changing of settings
func Settings(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		params := struct {
			User    models.User
			Title   string
			Flashes []interface{}
			Token   string
			Version string
		}{Title: "Settings", Version: config.Version, User: ctx.Get(r, "user").(models.User), Token: csrf.Token(r)}
		getTemplate(r, w, "settings").ExecuteTemplate(w, "base", params)
	case r.Method == "POST":
		msg := models.Response{Success: true, Message: "Settings Updated Successfully"}
		err := auth.ChangeLogo(r)

		if err != nil {
			msg.Message = err.Error()
			msg.Success = false
			JSONResponse(w, msg, http.StatusBadRequest)
			return
		}

		err = auth.ChangePassword(r)

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

// Logo serves custom logo image (if any) or the default logo
func Logo(w http.ResponseWriter, r *http.Request) {
	u, ok := ctx.Get(r, "user").(models.User)

	if ok {
		l := u.GetLogo()

		if l != nil {
			l.Serve(w)
			return
		}
	}

	http.Redirect(w, r, "/images/logo_inv_small.png", 302)
}

// Avatars_Id serves avatar image by the given user id or the default avatar
func Avatars_Id(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	user, err := models.GetUser(id)

	if err == nil {
		if user.Avatar != "" {
			user.ServeAvatar(w)
			return
		}
	}

	http.Redirect(w, r, "/images/noavatar.png", 302)
}

// SSO_Login handles Bakery Single Sign-On authentication flow for a user.
// If credentials are valid, a session is created.
func SSO_Login(w http.ResponseWriter, r *http.Request) {
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
	}{Title: "Login", Token: csrf.Token(r)}

	session := ctx.Get(r, "session").(*sessions.Session)

	switch {
	case r.Method == "GET":
		if _, err := r.Cookie("CHOCOLATECHIPSSL"); err == nil {
			http.Redirect(w, r, "/", 302)
			return
		}

		if cookie, err := r.Cookie("OATMEALSSL"); err == nil {
			c, err := bakery.ParseCookie(cookie.Value)

			if err != nil {
				log.Error(err)
			} else if c.Error != "" {
				log.Error(c.Error)
				Flash(w, r, "danger", c.Error)
			}

			cookie.Value = ""
			cookie.Expires = time.Unix(0, 0)
			cookie.MaxAge = -1
			cookie.Domain = auth.SSODomain
			cookie.Path = "/"
			cookie.Secure = true
			http.SetCookie(w, cookie)
		}

		params.Flashes = session.Flashes()
		session.Save(r, w)
		templates := template.New("template")

		_, err := templates.ParseFiles("templates/login.html", "templates/flashes.html")

		if err != nil {
			log.Error(err)
		}

		template.Must(templates, err).ExecuteTemplate(w, "base", params)
	case r.Method == "POST":
		username, password := r.FormValue("username"), r.FormValue("password")
		cookie, err := bakery.CreateOatmealCookie(username, password, "login", auth.SSOSlaveURL)

		if err != nil {
			Flash(w, r, "danger", err.Error())
			params.Flashes = session.Flashes()
			session.Save(r, w)
			templates := template.New("template")

			if _, err := templates.ParseFiles("templates/login.html", "templates/flashes.html"); err != nil {
				log.Error(err)
			}

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusUnauthorized)
			template.Must(templates, err).ExecuteTemplate(w, "base", params)
		} else {
			http.SetCookie(
				w, &http.Cookie{
					Name:    "OATMEALSSL",
					Value:   cookie,
					Domain:  auth.SSODomain,
					Path:    "/",
					Expires: time.Now().Add(1 * time.Hour),
				},
			)

			http.Redirect(w, r, auth.SSOMasterLoginURL, 302)
		}
	}
}

// func SSO_Mock(w http.ResponseWriter, r *http.Request) {
// 	authenticated := true

// 	if authenticated {
// 		_, err := bakery.CreateChocolatechipCookie("nonexistentcustomer@test.com", "Security Awareness User")

// 		if err != nil {
// 			w.WriteHeader(http.StatusInternalServerError)
// 		} else {
// 			http.SetCookie(
// 				w, &http.Cookie{
// 					Name:    "CHOCOLATECHIPSSL",
// 					Value:   "OGY1OTVlZjZjNjAzMDlmZmY4N2ZkMDg2MDZiMTgyZmYwZjhkNzFjMDY5NDEzNzFkN2NjYTNhNjU3NWE5MGQxMOBEgMt30umvtAH1pS5C%2FwQP9Y0HXHrQSbFgQMu3Ut4Omz%2Bwrc5Au69%2FUEehZ38oqDCezCkzlI8FR%2ByU5s4U4UaBmQyNMns5vz5aoMQ93vV63ZFQ0wAFyL7%2FN8WGnZwMxBr9tcvQRxUYMVKNrdJGumVVCI6XFxFVhyE4V9jKh2Vhqa9NB7OEu%2FEBYJZ4TMzTWfntkVFrXl2AgDers7lUHD6nebe%2BNWGYZZ8bfHwuI2gxuVWclDV7ieARhbxaljz%2FXwL8ZYG3Wn%2FJNKicCWG8%2BsWEt6t0MIuCbNw4422g7qDS3lo2Vnt63Y77LuBZuzca70ahrZ9KWAoiHsRR2WBh%2FbblIWYLDDrOSXCb2gtycNlP%2FvqxLbyjTkEm8rou6VIEHyQHNPqrbt5Kx%2FtVs4W90M6HEZfw7D%2FYuCn0ilVDgAgdmTU5oPORjdvrvi1sVtKQOPklzE6lehTjHUi3ZoCl8VE%3D",
// 					Domain:  ".localhost",
// 					Expires: time.Now().Add(1 * time.Hour),
// 					Path:    "/",
// 				},
// 			)

// 			if cookie, err := r.Cookie("OATMEALSSL"); err == nil {
// 				cookie.Value = ""
// 				cookie.Expires = time.Unix(0, 0)
// 				cookie.MaxAge = -1
// 				http.SetCookie(w, cookie)
// 			}
// 		}
// 	} else {
// 		http.SetCookie(
// 			w, &http.Cookie{
// 				Name:    "OATMEALSSL",
// 				Value:   `OTdjNmY1NzdiMjQ4YThjYzFlMjgzNjhjOTc3ZTUzNGVkN2RiN2I2YjllNzllMWZkZDIwZmY4YWViOWM1ZTIwMqWAFEmG%2FNutdJ93u4DxZKCMaMv1iB5au61d7RxCfvmj9gqjP5spZ4DzTnw3xpyvQUgiHaNlZbsI69quyt7hnqVNP2jq5Ev%2FsSvpFWno6KeyisZkPc7hs7LwfXeng7aYEMNbSl8O9j90G9eNYMVi8nTpqTF%2F3B4d2IBBIjlj2ym1wlWuJIuAs2pLU8vyb5wQkK5%2BaqQsNImTuC8CItkVYEqXKPRU4obtUy4%2FqpYqM04mO5%2FUtIW1QgzltHgPpsrmvvOw8NmOuAzLhJqp1aX1FWubum9TTCrWkNyHGkGdg8oZnh90Cu8WzTx%2F8Zsh63iPiV3U7FYz2oAQgV0d4TJtCGlnt95j1tukOvNYmNI1WRj6GaUcKthHhyqD3zU6WyBuiYYrlWcjuM4d%2FXHzs7dSc4AlUKCCaMPFgaOrAMzw4I9ROqlLQUDv3QGiGb24TWyvJw%3D%3D`,
// 				Domain:  ".localhost",
// 				Expires: time.Now().Add(1 * time.Hour),
// 				Path:    "/",
// 			},
// 		)
// 	}

// 	http.Redirect(w, r, "https://localhost:3333/bakery/login", 302)
// }

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
			log.Error(err)
		}
		template.Must(templates, err).ExecuteTemplate(w, "base", params)
	case r.Method == "POST":
		//Attempt to login
		succ, u, err := auth.Login(r)
		if err != nil {
			log.Error(err)
		}
		//If we've logged in, save the session and redirect to the dashboard
		if succ {
			session.Values["id"] = u.Id
			session.Save(r, w)
			next := "/"
			url, err := url.Parse(r.FormValue("next"))
			if err == nil {
				path := url.Path
				if path != "" {
					next = path
				}
			}
			http.Redirect(w, r, next, 302)
		} else {
			Flash(w, r, "danger", "Invalid Username/Password")
			params.Flashes = session.Flashes()
			session.Save(r, w)
			templates := template.New("template")
			_, err := templates.ParseFiles("templates/login.html", "templates/flashes.html")
			if err != nil {
				log.Error(err)
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusUnauthorized)
			template.Must(templates, err).ExecuteTemplate(w, "base", params)
		}
	}
}

// Logout destroys the current user session and deletes the SSO cookies (if any)
func Logout(w http.ResponseWriter, r *http.Request) {
	if session, ok := ctx.Get(r, "session").(*sessions.Session); ok {
		if _, ok := session.Values["id"]; ok {
			delete(session.Values, "id")
			Flash(w, r, "success", "You have successfully logged out")
			session.Save(r, w)
		}
	}

	for _, c := range r.Cookies() {
		if c.Name == "CHOCOLATECHIPSSL" ||
			strings.HasPrefix(c.Name, "SESS") ||
			strings.HasPrefix(c.Name, "SSESS") {
			c.Value = ""
			c.Expires = time.Unix(0, 0)
			c.MaxAge = -1
			c.Domain = auth.SSODomain
			c.Path = "/"
			c.Secure = true
			c.HttpOnly = true
			http.SetCookie(w, c)
		}
	}

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
		log.Error(url)
	}
	http.Error(w, "No URL given.", http.StatusBadRequest)
}

func getTemplate(r *http.Request, w http.ResponseWriter, tmpl string) *template.Template {
	templates := template.New("template").Funcs(template.FuncMap{
		"page": func() string {
			return tmpl
		},

		"role": func() string {
			role, err := models.GetUserRole(ctx.Get(r, "user").(models.User).Id)

			if err != nil {
				log.Error(err)
			}

			return role.Name()
		},
	})

	_, err := templates.ParseFiles("templates/base.html", "templates/"+tmpl+".html", "templates/flashes.html", "templates/sidebar.html")

	if err != nil {
		log.Error(err)
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
}
