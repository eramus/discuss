package topic

import (
	"fmt"
	"html"
	"html/template"
//	"log"
	"net/http"
	"strconv"
	"strings"

	"code.google.com/p/gorilla/sessions"

	"discuss/shared"
)

func Bump(r *http.Request, sess *sessions.Session) (body *shared.Body, tpl *template.Template, redirect string) {
//	log.Println("route: bump topic")
	var u_id = sess.Values["id"].(uint64)
	parts := strings.Split(html.EscapeString(r.URL.Path[1:]), "/")
	if len(parts) < 3 {
		redirect = "/"
		return
	}
	t_id, _ := strconv.ParseInt(parts[2], 10, 64)
	id := uint64(t_id)
	key := fmt.Sprintf("topic:%d:d_id", id)
	de, rerr := shared.RedisClient.Get(key)
	if rerr != nil {
		redirect = "/"
		return
	}
	d_id := uint64(de.Int64())
	key = fmt.Sprintf("discussion:%d:uri", d_id)
	ue, rerr := shared.RedisClient.Get(key)
	if rerr != nil {
		redirect = "/"
		return
	}
	redirect = fmt.Sprintf("/discuss/%s", ue.String())
	// check if already bumped
	key = fmt.Sprintf("topic:%d:bumped", id)
	voted, rerr := shared.RedisClient.Sismember(key, u_id)
	if rerr != nil || voted {
		// already voted
		return
	}
	go BumpThread(u_id, id, d_id)
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
	t_id, _ := strconv.ParseInt(parts[2], 10, 64)
	id := uint64(t_id)
	key := fmt.Sprintf("topic:%d:d_id", id)
	de, rerr := shared.RedisClient.Get(key)
	if rerr != nil {
		redirect = "/"
		return
	}
	d_id := uint64(de.Int64())
	key = fmt.Sprintf("discussion:%d:uri", d_id)
	ue, rerr := shared.RedisClient.Get(key)
	if rerr != nil {
		redirect = "/"
		return
	}
	redirect = fmt.Sprintf("/discuss/%s", ue.String())
	// check if already buried
	key = fmt.Sprintf("topic:%d:buried", id)
	voted, rerr := shared.RedisClient.Sismember(key, u_id)
	if rerr != nil || voted {
		// already voted
		return
	}
	go BuryThread(u_id, id, d_id)
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
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		redirect = "/"
		return
	}
	t, err := GetById(uint64(id))
	body = new(shared.Body)
	labels, uris := shared.GetTopicBreadcrumbs(t.Id)
	labels, uris = append(labels, t.Title), append(uris, "")
	body.Breadcrumbs = &shared.Breadcrumbs{labels, uris}
	body.ContentData = t
	go AddThreadView(t.DId, t.Id)
	tpl = viewTpls
	return
}