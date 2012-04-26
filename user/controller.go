package user

import (
	"html"
	"log"
	"net/http"
	"strings"

	"code.google.com/p/gorilla/sessions"

	"discuss/shared"
)

func Register(r *http.Request, sess *sessions.Session) (body *shared.Body, files []string, redirect string) {
//	log.Println("route: register")
	var id uint64
	if _, ok := sess.Values["id"]; ok {
		id = sess.Values["id"].(uint64)
	}
	if id == 0 {
		if r.Method != "POST" {
			body, files = RegisterForm(r)
		} else {
			id, err := Add(r)
			if err != nil {
				body, files = RegisterForm(r)
			} else {
				sess.Values["id"] = id
				redirect = "/"
			}
		}
	} else {
		redirect = "/"
	}
	return
}

func Login(r *http.Request, sess *sessions.Session) (body *shared.Body, files []string, redirect string) {
//	log.Println("route: login")
	var id uint64
	if _, ok := sess.Values["id"]; ok {
		id = sess.Values["id"].(uint64)
	}
	if id == 0 {
		if r.Method != "POST" {
			body, files = LoginForm(r)
		} else {
			id, err := Authenticate(r)
			if err != nil {
				body, files = LoginForm(r)
			} else {
				sess.Values["id"] = id
				fs := sess.Flashes("redirect")
				redirect = "/"
				if len(fs) > 0 {
					redirect = fs[0].(string)
				}
				// check for remember flag
/*				if r.FormValue("remember") == "1" {
					shared.Remember(r, )
				}*/
			}
		}
	} else {
		redirect = "/"
	}
	return
}

func Logout(r *http.Request, sess *sessions.Session) (body *shared.Body, files []string, redirect string) {
	log.Println("route: logout")
	sess.Values["id"] = uint64(0)
	redirect = "/"
	fs := sess.Flashes("last")
	if len(fs) != 0 {
		redirect = fs[0].(string)
	}
	return
}

func Profile(r *http.Request, sess *sessions.Session) (body *shared.Body, files []string, redirect string) {
	// figure out where we are
	parts := strings.Split(html.EscapeString(r.URL.Path[1:]), "/")
	if len(parts) < 2 {
		redirect = "/"
		return
	}
	username := parts[1]
	u_id, _ := shared.RedisClient.Get("user:" + username)
	if u_id == nil {
		redirect = "/"
		return
	}
	id := uint64(u_id.Int64())
	var logged_in_id uint64
	if v, ok := sess.Values["id"]; ok {
		logged_in_id = v.(uint64)
	}
	if logged_in_id == id {
		// show own user view
		body, files = FeedPage()
		return
	}
	// show other user view
	body, files = ProfilePage()
	return
}























