package discussion

import (
	"net/http"

	"discuss/topic"
	"discuss/shared"
)

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

func AddForm(r *http.Request) (body *shared.Body, files []string) {
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
	files = append(files, "./templates/discussion/form.tpl")
	return
}