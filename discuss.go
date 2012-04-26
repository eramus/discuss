package main

import (
	"log"
	"net/http"
	"regexp"

	"code.google.com/p/gorilla/sessions"

	"discuss/discussion"
	"discuss/home"
	"discuss/post"
	"discuss/shared"
	"discuss/topic"
	"discuss/user"
)

func main() {
	r := new(shared.Router)
	r.SetNotFound(home.Index)
	r.SetAuth(user.Login)
	
	loginCallback := func(w http.ResponseWriter, r *http.Request, sess *sessions.Session) error {
		if idi, ok := sess.Values["id"]; ok {
			id := idi.(uint64)
			if r.FormValue("remember") != "1" || id == 0 {
				log.Println("no remember")
				return nil
			}
			log.Println("remember")
			return shared.Remember(r, w, id)
		}
		return nil
	}
	logoutCallback := func(w http.ResponseWriter, r *http.Request, sess *sessions.Session) error {
		cookie := &http.Cookie{
			Name:     "remember",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
		}
		http.SetCookie(w, cookie)
		return nil
	}

	r.AddRoute(regexp.MustCompile(`^/register`), user.Register, nil, false)
	r.AddRoute(regexp.MustCompile(`^/login`), user.Login, loginCallback, false)
	r.AddRoute(regexp.MustCompile(`^/logout`), user.Logout, logoutCallback, false)
	r.AddRoute(regexp.MustCompile(`^/user/.+`), user.Profile, nil, false)
	r.AddRoute(regexp.MustCompile(`^/new/post`), post.AddPost, nil, true)
	r.AddRoute(regexp.MustCompile(`^/new/topic`), topic.AddTopic, nil, true)
	r.AddRoute(regexp.MustCompile(`^/new/discussion`), discussion.AddDiscussion, nil, true)
	r.AddRoute(regexp.MustCompile(`^/topic`), topic.ViewTopic, nil, false)
	r.AddRoute(regexp.MustCompile(`^/discuss`), discussion.ViewDiscussion, nil, false)
	r.AddRoute(regexp.MustCompile(`^/bump/post/[0-9]+`), post.Bump, nil, true)
	r.AddRoute(regexp.MustCompile(`^/bury/post/[0-9]+`), post.Bury, nil, true)
	r.AddRoute(regexp.MustCompile(`^/bump/topic/[0-9]+`), topic.Bump, nil, true)
	r.AddRoute(regexp.MustCompile(`^/bury/topic/[0-9]+`), topic.Bury, nil, true)
	r.AddRoute(regexp.MustCompile(`^/subscribe`), discussion.Subscribe, nil, true)
	r.AddRoute(regexp.MustCompile(`^/unsubscribe`), discussion.Unsubscribe, nil, true)
	r.AddRoute(regexp.MustCompile(`^/search`), home.Search, nil, false)
	r.AddRoute(regexp.MustCompile(`^/$`), home.Index, nil, false)
	log.Fatal("ListenAndServe: ", http.ListenAndServe(":8081", r))
}
