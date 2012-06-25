package post

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
//	log.Println("route: bump post")
	var u_id = sess.Values["id"].(uint64)
	parts := strings.Split(html.EscapeString(r.URL.Path[1:]), "/")
	if len(parts) < 3 {
		redirect = "/"
		return
	}
	id, _ := strconv.ParseUint(parts[2], 10, 64)
	e, _ := shared.RedisClient.Get(fmt.Sprintf("post:%d:t_id", id))
	redirect = fmt.Sprintf("/topic/%d", uint64(e.Int64()))
	// check if already bumped
	key := fmt.Sprintf("post:%d:bumped", id)
	voted, rerr := shared.RedisClient.Sismember(key, u_id)
	if rerr != nil || voted {
		// already voted
		return
	}
	go BumpPost(u_id, id)
	return
}

func Bury(r *http.Request, sess *sessions.Session) (body *shared.Body, tpl *template.Template, redirect string) {
//	log.Println("route: bury post")
	var u_id = sess.Values["id"].(uint64)
	parts := strings.Split(html.EscapeString(r.URL.Path[1:]), "/")
	if len(parts) < 3 {
		redirect = "/"
		return
	}
	id, _ := strconv.ParseUint(parts[2], 10, 64)
	e, _ := shared.RedisClient.Get(fmt.Sprintf("post:%d:t_id", id))
	redirect = fmt.Sprintf("/topic/%d", uint64(e.Int64()))
	// check if already buried
	key := fmt.Sprintf("post:%d:buried", id)
	voted, rerr := shared.RedisClient.Sismember(key, u_id)
	if rerr != nil || voted {
		// already voted
		return
	}
	go BuryPost(u_id, id)
	return
}

func AddPost(r *http.Request, sess *sessions.Session) (body *shared.Body, tpl *template.Template, redirect string) {
//	log.Println("route: add post")
	var u_id = sess.Values["id"].(uint64)
	parts := strings.Split(html.EscapeString(r.URL.Path[1:]), "/")
	if len(parts) < 2 {
		redirect = "/"
		return
	}
	t_id, _ := strconv.ParseUint(parts[2], 10, 64)
	var p_id uint64
	if len(parts) > 3 {
		p_id, _ = strconv.ParseUint(parts[3], 10, 64)
	}
	if r.Method != "POST" {
		body, tpl = AddForm(r, t_id, p_id)
	} else {
		id, err := Add(r, u_id)
		if err != nil {
			body, tpl = AddForm(r, t_id, p_id)
		} else {
			redirect = fmt.Sprintf("/topic/%d#%d", t_id, id)
		}
	}
	return
}