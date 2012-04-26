package topic

import (
	"html/template"
//	"io"
//	"log"
	"net/http"
	"strings"

	"discuss/post"
	"discuss/shared"
)

var addTpls = template.Must(template.ParseFiles(
		append(shared.Templates, "./templates/topic/form.tpl")...
))
var viewTpls = template.Must(template.ParseFiles(
		append(shared.Templates, "./templates/topic/view.tpl", "./templates/topic/posts.tpl")...
))

type Topic struct {
	Id uint64
	DId uint64
	Uri string
	Title string
	Posts []*post.Post
	Comments int
	LastPost string
	LastPostId uint64
	NumPosts int64
}

type Form struct {
	Uri		string
	DId		uint64
	Title		string
	Post	string
}

func AddForm(r *http.Request, d_id uint64, parts[]string) (body *shared.Body, tpl *template.Template) {
	body = new(shared.Body)
	labels, uris := shared.GetDiscussionBreadcrumbs(d_id, false)
	labels, uris = append(labels, "Add Topic"), append(uris, "")
	body.Breadcrumbs = &shared.Breadcrumbs{labels, uris}
	
	f := new(Form)
	f.Uri = strings.Join(parts, "/")
	f.DId = d_id
	
	if r.Method == "POST" {
		f.Title = r.FormValue("title")
		f.Post = r.FormValue("post")

	}
	body.ContentData = f
	tpl = addTpls
	return
}