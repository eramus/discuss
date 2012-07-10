package shared

import (
	"html/template"
	"log"
	"net/http"
	"net/http/pprof"
	"regexp"

	"code.google.com/p/gorilla/sessions"
)

type requestHandler func(*http.Request, *sessions.Session) (*Body, *template.Template, string)
type responseHandler func(http.ResponseWriter, *http.Request, *sessions.Session) error

type route struct {
	flat         string
	regex        *regexp.Regexp
	request      requestHandler
	response     responseHandler
	requireLogin bool
}

type Router struct {
	routes       []route
	flatRoutes   []route
	staticRoutes []*regexp.Regexp
	notFound     requestHandler
	auth         requestHandler
	profilier    *regexp.Regexp
}

func (r *Router) AddProfilier() {
	r.profilier = regexp.MustCompile(`^/debug/pprof(/(heap|symbol|profile|cmdline)|/|)`)
}

func (r *Router) AddStatic(rt *regexp.Regexp) {
	r.staticRoutes = append(r.staticRoutes, rt)
}

func (r *Router) AddRoute(rt interface{}, handler requestHandler, callback responseHandler, requireLogin bool) {
	if rr, ok := rt.(*regexp.Regexp); ok {
		r.routes = append(r.routes, route{"", rr, handler, callback, requireLogin})
	} else if sr, ok := rt.(string); ok {
		r.flatRoutes = append(r.flatRoutes, route{sr, nil, handler, callback, requireLogin})
	}
}

func (r *Router) SetNotFound(handler requestHandler) {
	r.notFound = handler
}

func (r *Router) SetAuth(handler requestHandler) {
	r.auth = handler
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// check if this is a profilier request
	if r.profilier != nil {
		if r.profilier.MatchString(req.URL.Path) {
			res := r.profilier.FindAllStringSubmatch(req.URL.Path, -1)
			switch res[0][2] {
			case "cmdline":
				pprof.Cmdline(w, req)
			case "symbol":
				pprof.Symbol(w, req)
			case "profile":
				pprof.Profile(w, req)
			default:
				pprof.Index(w, req)
			}
			return
		}
	}
	// check for static routes first
	for _, rt := range r.staticRoutes {
		if rt.MatchString(req.URL.Path) {
			http.ServeFile(w, req, "./static"+req.URL.Path)
			return
		}
	}
	go countRequest()
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
		u_id, _ = Regen(req)
		if u_id > 0 {
			sess.Values["id"] = u_id
		}
	}
	for _, rt := range r.flatRoutes {
		if req.URL.Path == rt.flat || req.URL.Path == rt.flat+"/" {
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
		for _, rt := range r.routes {
			if rt.regex.MatchString(req.URL.Path) {
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
	}
	if reqhandler == nil {
		w.WriteHeader(404)
		reqhandler = r.notFound
	}
	body, tpl, redirect := reqhandler(req, sess)
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
	sess.Save(req, w)
	if tpl != nil {
		err = tpl.Execute(w, page)
		if err != nil {
			log.Println("output err:", err)
		}
	}
}
