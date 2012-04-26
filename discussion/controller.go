package discussion

import (
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"strings"

	"code.google.com/p/gorilla/sessions"

	"discuss/shared"
)

func Subscribe(r *http.Request, sess *sessions.Session) (body *shared.Body, tpl *template.Template, redirect string) {
//	log.Println("route: subscribe")
	parts := strings.Split(html.EscapeString(r.URL.Path[1:]), "/")
	if len(parts) < 2 {
		redirect = "/"
		return
	}
	parts = parts[1:]
	var uri = strings.Join(parts, "/")
	id, rerr := GetId(uri)
	if rerr != nil {
		log.Println("redis err:", rerr)
		return
	}
	var u_id = sess.Values["id"].(uint64)
	key := fmt.Sprintf("user:%d:subscribed", u_id)
	added, rerr := shared.RedisClient.Sadd(key, id)
	if rerr != nil {
		log.Println("redis err:", rerr)
		return
	}
	if added {
		key = fmt.Sprintf("discussion:%d:subscribed", id)
		_, rerr := shared.RedisClient.Incr(key)
		if rerr != nil {
			log.Println("redis err:", rerr)
			return
		}
	}
	redirect = "/discuss/" + uri
	return
}

func Unsubscribe(r *http.Request, sess *sessions.Session) (body *shared.Body, tpl *template.Template, redirect string) {
//	log.Println("route: unsubscribe")
	parts := strings.Split(html.EscapeString(r.URL.Path[1:]), "/")
	if len(parts) < 2 {
		redirect = "/"
		return
	}
	parts = parts[1:]
	var uri = strings.Join(parts, "/")
	id, rerr := GetId(uri)
	if rerr != nil {
		log.Println("redis err:", rerr)
		return
	}
	var u_id = sess.Values["id"].(uint64)
	key := fmt.Sprintf("user:%d:subscribed", u_id)
	removed, rerr := shared.RedisClient.Srem(key, id)
	if rerr != nil {
		log.Println("redis err:", rerr)
		return
	}
	if removed {
		key = fmt.Sprintf("discussion:%d:subscribed", id)
		_, rerr := shared.RedisClient.Decr(key)
		if rerr != nil {
			log.Println("redis err:", rerr)
			return
		}
	}
	redirect = "/discuss/" + uri
	return
}

func AddDiscussion(r *http.Request, sess *sessions.Session) (body *shared.Body, tpl *template.Template, redirect string) {
//	log.Println("route: add discussion")
	var u_id = sess.Values["id"].(uint64)
	if r.Method != "POST" {
		body, tpl = AddForm(r)
	} else {
		id, err := Add(r, u_id)
		if err != nil {
			log.Println("add err:", err)
			body, tpl = AddForm(r)
		} else {
			redirect, _ = GetUri(id)
			redirect = "/discuss/" + redirect
		}
	}
	return
}

func ViewDiscussion(r *http.Request, sess *sessions.Session) (body *shared.Body, tpl *template.Template, redirect string) {
//	log.Println("route: view discussion")
	var u_id uint64
	if _, ok := sess.Values["id"]; ok {
		u_id = sess.Values["id"].(uint64)
	}
	// figure out where we are
	parts := strings.Split(html.EscapeString(r.URL.Path[1:]), "/")
	if len(parts) < 2 {
		redirect = "/"
		return
	}
	parts = parts[1:]
	var uri = strings.Join(parts, "/")
	id, err := GetId(uri)
	if err != nil {
		// uhh
		redirect = "/"
	} else if id > 0 {
		// show current topics -- w pagination
/*		if !shared.CanDo(u_id, id, VIEW) {
			log.Println("no permission")
			redirect = "/"
		} else {*/
			ts, _ := Topics(id)
			key := fmt.Sprintf("discussion:%d:title", id)
			te, rerr := shared.RedisClient.Get(key)
			if rerr != nil {
				log.Println("redis err:", rerr)
				redirect = "/"
			} else {
				labels, uris := shared.GetDiscussionBreadcrumbs(id, true)
				key := fmt.Sprintf("user:%d:subscribed", u_id)
				im, rerr := shared.RedisClient.Sismember(key, id)
				if rerr != nil {
					log.Println("redis err:", rerr)
					redirect = "/"
				}
				if im {
					labels, uris = append(labels, "Unsubscribe"), append(uris, "/unsubscribe/" + uri)
				} else {
					labels, uris = append(labels, "Subscribe"), append(uris, "/subscribe/" + uri)
				}
				body = &shared.Body{
					Breadcrumbs: &shared.Breadcrumbs{labels, uris},
					ContentData: &List{
						Id: id,
						Uri: uri,
						Title: te.String(),
						Topics: ts,
					},
				}
				tpl = listTpls
			}
//		}
	} else {
		// want to add?
		body, tpl = AddForm(r)
		if err != nil {
			log.Println("no page:", err)
		}
	}
	return
}