package shared

import (
	"log"
	"net/http"
	"regexp"

	"code.google.com/p/gorilla/sessions"
)

type requestHandler func(*http.Request, *sessions.Session) (*Body, []string, string)
type responseHandler func(http.ResponseWriter, *http.Request, *sessions.Session) error

type route struct {
	match		*regexp.Regexp
	request		requestHandler
	response	responseHandler
	requireLogin	bool
}

type Router struct {
	routes		[]route
	notFound	requestHandler
	auth		requestHandler
}

func (r *Router) AddRoute(regex *regexp.Regexp, handler requestHandler, callback responseHandler, requireLogin bool) {
	r.routes = append(r.routes, route{regex, handler, callback, requireLogin})
}

func (r *Router) SetNotFound(handler requestHandler) {
	r.notFound = handler
}

func (r *Router) SetAuth(handler requestHandler) {
	r.auth = handler
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	sess, err := getSession(req, "auth")
	if err != nil {
		log.Println("session err:", err)
		return
	}
	var reqhandler requestHandler
	var respHandler responseHandler
	var u_id uint64
	if u, ok := sess.Values["id"]; ok {
		u_id = u.(uint64)
	}
	if u_id == 0 {
		log.Println("regen?")
		u_id, _ = Regen(req)
		if u_id > 0 {
			sess.Values["id"] = u_id
		}
	}
	for _, rt := range r.routes {
		if rt.match.MatchString(req.URL.Path) {
			reqhandler = rt.request
			respHandler = rt.response
			if !rt.requireLogin {
				break
			}
			if u_id == 0 {
				sess.AddFlash(req.URL.Path, "redirect")
				reqhandler = r.auth
				respHandler = nil
			}
			break
		}
	}
	if reqhandler == nil {
		w.WriteHeader(404)
		reqhandler = r.notFound
	}
	body, files, redirect := reqhandler(req, sess)
	visited(sess)
//	sess.AddFlash(req.URL.Path, "last")
	if respHandler != nil {
		log.Println("run callback")
		respHandler(w, req, sess)
	}
	if redirect != "" {
		sess.Save(req, w)
		http.Redirect(w, req, redirect, 302)
		return
	}
	if body == nil {
		body = new(Body)
	}
	body.UserData = getUserData(sess)
	if !body.NoSidebar {
		// setup side bar
		body.Subscribed = getSubscribed(sess)
	}
	page := new(Page)
	page.Body = body
	wrapper, err := Wrapper.Clone()
	if err != nil {
		log.Println("template err:", err)
		return
	}
	_, err = wrapper.ParseFiles(files...)
	if err != nil {
		log.Println("template err:", err)
		return
	}
	sess.Save(req, w)
	err = wrapper.Execute(w, page)
	if err != nil {
		log.Println("output err:", err)
	}
}