package topic

import (
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"strconv"
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
	id, _ := strconv.ParseUint(parts[2], 10, 64)
	if !Exists(id) {
		redirect = "/"
		return
	}
	var u_id = sess.Values["id"].(uint64)
	key := fmt.Sprintf("user:%d:subscribed", u_id)
	_, rerr := shared.RedisClient.Sadd(key, id)
	if rerr != nil {
		log.Println("redis err:", rerr)
		return
	}
	redirect = "/topic/" + parts[2]
	return
}

func Unsubscribe(r *http.Request, sess *sessions.Session) (body *shared.Body, tpl *template.Template, redirect string) {
//	log.Println("route: unsubscribe")
	parts := strings.Split(html.EscapeString(r.URL.Path[1:]), "/")
	if len(parts) < 2 {
		redirect = "/"
		return
	}
	id, _ := strconv.ParseUint(parts[2], 10, 64)
	if !Exists(id) {
		redirect = "/"
		return
	}
	var u_id = sess.Values["id"].(uint64)
	key := fmt.Sprintf("user:%d:subscribed", u_id)
	_, rerr := shared.RedisClient.Srem(key, id)
	if rerr != nil {
		log.Println("redis err:", rerr)
		return
	}
	redirect = "/topic/" + parts[2]
	return
}

func Bump(r *http.Request, sess *sessions.Session) (body *shared.Body, tpl *template.Template, redirect string) {
//	log.Println("route: bump topic")
	var u_id = sess.Values["id"].(uint64)
	parts := strings.Split(html.EscapeString(r.URL.Path[1:]), "/")
	if len(parts) < 3 {
		redirect = "/"
		return
	}
	id, _ := strconv.ParseUint(parts[2], 10, 64)
	if !Exists(id) {
		redirect = "/"
		return
	}
	de, rerr := shared.RedisClient.Get(fmt.Sprintf("topic:%d:d_id", id))
	if rerr != nil {
		redirect = "/"
		return
	}
	redirect = fmt.Sprintf("/topic/%d", id)
	// check if already bumped
	voted, rerr := shared.RedisClient.Sismember(fmt.Sprintf("topic:%d:bumped", id), u_id)
	if rerr != nil || voted {
		// already voted
		return
	}
	go BumpThread(u_id, id, uint64(de.Int64()))
	return
}

func Bury(r *http.Request, sess *sessions.Session) (body *shared.Body, tpl *template.Template, redirect string) {
//	log.Println("route: bury topic")
	var u_id = sess.Values["id"].(uint64)
	parts := strings.Split(html.EscapeString(r.URL.Path[1:]), "/")
	if len(parts) < 3 {
		redirect = "/"
		return
	}
	id, _ := strconv.ParseUint(parts[2], 10, 64)
	if !Exists(id) {
		redirect = "/"
		return
	}
	de, rerr := shared.RedisClient.Get(fmt.Sprintf("topic:%d:d_id", id))
	if rerr != nil || de == nil {
		redirect = "/"
		return
	}
	redirect = fmt.Sprintf("/topic/%d", id)
	// check if already buried
	voted, rerr := shared.RedisClient.Sismember(fmt.Sprintf("topic:%d:buried", id), u_id)
	if rerr != nil || voted {
		// already voted
		return
	}
	go BuryThread(u_id, id, uint64(de.Int64()))
	return
}

func AddTopic(r *http.Request, sess *sessions.Session) (body *shared.Body, tpl *template.Template, redirect string) {
//	log.Println("route: add topic")
	var u_id = sess.Values["id"].(uint64)
	parts := strings.Split(html.EscapeString(r.URL.Path[1:]), "/")
	if len(parts) < 2 {
		redirect = "/"
		return
	}
	parts = parts[2:]
	uri := strings.Join(parts, "/")
	res, rerr := shared.RedisClient.Get("discussions:" + uri)
	if rerr != nil {
		redirect = "/"
		return
	}
	d_id := uint64(res.Int64())
	if r.Method != "POST" {
		body, tpl = AddForm(r, d_id, parts)
	} else {
		id, err := Add(r, u_id)
		if err != nil {
			body, tpl = AddForm(r, d_id, parts)
		} else {
			redirect = fmt.Sprintf("/topic/%d/", id)
		}
	}
	return
}

func ViewTopic(r *http.Request, sess *sessions.Session) (body *shared.Body, tpl *template.Template, redirect string) {
//	log.Println("route: view topic")
	parts := strings.Split(html.EscapeString(r.URL.Path[1:]), "/")
	if len(parts) < 2 {
		redirect = "/"
		return
	}
	id, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		redirect = "/"
		return
	}
	if !Exists(id) {
		redirect = "/"
		return
	}
	t, err := GetById(id)
	body = new(shared.Body)
	labels, uris := shared.GetTopicBreadcrumbs(t.Id)
	labels, uris = append(labels, t.Title), append(uris, "")
	body.Breadcrumbs = &shared.Breadcrumbs{labels, uris}
	body.ContentData = t
	body.Title = t.Title
	go AddThreadView(t.DId, t.Id)
	tpl, _ = viewTpls.Clone()
	tpl.Parse(shared.GetPageTitle(t.Title))
	return
}