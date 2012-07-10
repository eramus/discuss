package user

import (
	"html/template"
	"net/http"

	"discuss/shared"
)

var registerTpls = template.Must(template.ParseFiles(
	append(shared.Templates, "./templates/user/register.tpl")...,
))
var loginTpls = template.Must(template.ParseFiles(
	append(shared.Templates, "./templates/user/login.tpl")...,
))
var profileTpls = template.Must(template.ParseFiles(
	append(shared.Templates, "./templates/user/profile.tpl")...,
))
var feedTpls = template.Must(template.ParseFiles(
	append(shared.Templates, "./templates/user/feed.tpl")...,
))

type Form struct {
	Username string
	Email    string
	Remember string
}

func registerForm(r *http.Request) (body *shared.Body, tpl *template.Template) {
	body = new(shared.Body)
	if r.Method == "POST" {
		f := new(Form)
		f.Username = r.FormValue("username")
		f.Email = r.FormValue("email")

		body.ContentData = f
	}
	body.Breadcrumbs = &shared.Breadcrumbs{
		Labels: []string{"Register"},
		Uris:   []string{""},
	}
	body.Title = "Register"
	tpl, _ = registerTpls.Clone()
	tpl.Parse(shared.GetPageTitle("Register"))
	return
}

func loginForm(r *http.Request) (body *shared.Body, tpl *template.Template) {
	body = new(shared.Body)
	if r.Method == "POST" {
		f := new(Form)
		f.Username = r.FormValue("username")
		f.Remember = r.FormValue("remember")

		body.ContentData = f
	}
	body.Breadcrumbs = &shared.Breadcrumbs{
		Labels: []string{"Login"},
		Uris:   []string{""},
	}
	body.Title = "Login"
	tpl, _ = loginTpls.Clone()
	tpl.Parse(shared.GetPageTitle("Login"))
	return
}

func profilePage() (body *shared.Body, tpl *template.Template) {
	body = new(shared.Body)
	body.Title = "Profile"
	tpl, _ = profileTpls.Clone()
	tpl.Parse(shared.GetPageTitle("Profile"))
	return
}

func feedPage() (body *shared.Body, tpl *template.Template) {
	body = new(shared.Body)
	body.NoSidebar = true
	body.Title = "Feed"
	tpl, _ = feedTpls.Clone()
	tpl.Parse(shared.GetPageTitle("Feed"))
	return
}
