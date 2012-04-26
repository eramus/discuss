package user

import (
	"net/http"
	
	"discuss/shared"
)

type Form  struct {
	Username	string
	Email		string
	Remember	string
}

func RegisterForm(r *http.Request) (body *shared.Body, files []string) {
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
	files = append(files, "./templates/user/register.tpl")
	return
}

func LoginForm(r *http.Request) (body *shared.Body, files []string) {
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
	files = append(files, "./templates/user/login.tpl")
	return
}

func ProfilePage() (body *shared.Body, files []string) {
	body = new(shared.Body)
	files = append(files, "./templates/user/profile.tpl")
	return
}

func FeedPage() (body *shared.Body, files []string) {
	body = new(shared.Body)
	body.NoSidebar = true
	files = append(files, "./templates/user/feed.tpl")
	return
}