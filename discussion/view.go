package discussion

import (
	"html/template"
	"net/http"

	"discuss/topic"
	"discuss/shared"
)

var addTpls = template.Must(template.ParseFiles(
		append(shared.Templates, "./templates/discussion/form.tpl")...
))

var listTpls = template.Must(template.ParseFiles(
		append(shared.Templates, "./templates/discussion/listing.tpl")...
))

type List struct {
	Id		uint64
	Uri		string
	Title		string
	Topics	[]topic.Topic
}

type Form struct {
	Uri			string
	Title			string
	Description	string
	Keywords	string
}

func AddForm(r *http.Request) (body *shared.Body, tpl *template.Template) {
	body = new(shared.Body)
	if r.Method == "POST" {
		f := new(Form)
		f.Uri = r.FormValue("uri")
		f.Title = r.FormValue("title")
		f.Description = r.FormValue("description")
		f.Keywords = r.FormValue("keywords")

		body.ContentData = f
	}
	body.Breadcrumbs = &shared.Breadcrumbs {
		Labels: []string{"Add Discussion"},
		Uris: []string{""},
	}
	tpl = addTpls
	return
}