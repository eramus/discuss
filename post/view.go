package post

import (
	"fmt"
	"net/http"

	"discuss/shared"
)

type Post struct {
	Id uint64
	TId uint64
	UId uint64
	Username string
	Post string
	Posts []Post
	Timestamp uint64
	FTimestamp	string
}

type Form struct {
	TId		uint64
	PId		uint64
	Post	string
}

func AddForm(r *http.Request, t_id, p_id uint64) (body *shared.Body, files []string) {
	t, rerr := shared.RedisClient.Get(fmt.Sprintf("topic:%d:title", t_id))
	if rerr != nil {
		return
	}
	labels, uris := shared.GetTopicBreadcrumbs(t_id)
	if p_id > 0 {
		labels, uris = append(labels, t.String()), append(uris, fmt.Sprintf("/topic/%d#%d", t_id, p_id))
	} else {
		labels, uris = append(labels, t.String()), append(uris, fmt.Sprintf("/topic/%d", t_id))
	}
	labels, uris = append(labels, "Add Post"), append(uris, "")
	
	body = new(shared.Body)
	f := new(Form)
	f.TId = t_id
	if p_id > 0 {
		f.PId = p_id
	}
	if r.Method == "POST" {
		f.Post = r.FormValue("post")
	}
	body.Breadcrumbs = &shared.Breadcrumbs{labels, uris}
	body.ContentData = f

	files = append(files, "./templates/post/form.tpl")
	return
}