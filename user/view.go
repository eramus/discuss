package user

import (
	"html/template"
	"net/http"
	
	"discuss/shared"
)

var registerTpls = template.Must(template.ParseFiles(
		append(shared.Templates, "./templates/user/register.tpl")...
))
var loginTpls = template.Must(template.ParseFiles(
		append(shared.Templates, "./templates/user/login.tpl")...
))
var profileTpls = template.Must(template.ParseFiles(
		append(shared.Templates, "./templates/user/profile.tpl")...
))
var feedTpls = template.Must(template.ParseFiles(
		append(shared.Templates, "./templates/user/feed.tpl")...
))

type Form  struct {
	Username	string
	Email		string
	Remember	string
}

func RegisterForm(r *http.Request) (body *shared.Body, tpl *template.Template) {
	body = new(shared.Body)
	if r.Method == "POST" {
		f := new(Form)
		f.Username = r.FormValue("username")
		f.Email = r.FormValue("email")

		body.ContentData = f
	}
	body.Breadcrumbs = &shared.Breadcrumbs {
		Labels: []string{"Register"},
		Uris: []string{""},
	}
	tpl = registerTpls
	return
}

func LoginForm(r *http.Request) (body *shared.Body, tpl *template.Template) {
	body = new(shared.Body)
	if r.Method == "POST" {
		f := new(Form)
		f.Username = r.FormValue("username")
		f.Remember = r.FormValue("remember")

		body.ContentData = f
	}
	body.Breadcrumbs = &shared.Breadcrumbs {
		Labels: []string{"Login"},
		Uris: []string{""},
	}
	tpl = loginTpls
	return
}

func ProfilePage() (body *shared.Body, tpl *template.Template) {
	body = new(shared.Body)
	tpl = profileTpls
	return
}

func FeedPage() (body *shared.Body, tpl *template.Template) {
	body = new(shared.Body)
	body.NoSidebar = true
	tpl = feedTpls
	return
}