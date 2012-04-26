package home

import (
	"fmt"
	"log"
	"net/http"

	"code.google.com/p/gorilla/sessions"

	"discuss/discussion"
	"discuss/shared"
)

type Results struct {
	NumFound	int
	Start		int
	Discussions	[]shared.DiscussionDoc
	Posts		[]shared.PostDoc
	Query		string
}

func Index(r *http.Request, sess *sessions.Session) (body *shared.Body, files []string, redirect string) {
//	log.Println("route: index")
	files = append(files, "./templates/home/home.tpl")
	return
}

func Search(r *http.Request, sess *sessions.Session) (body *shared.Body, files []string, redirect string) {
//	log.Println("route: search")
	if r.Method == "POST" {
		query := r.FormValue("search")
		if r.Method != "POST" || query == "" {
			redirect = "/"
			return
		}
		q := fmt.Sprintf(`keywords:"%s" title:"%s" uri:"%s" desc:"%s"`, query, query, query, query)
		dr, err := shared.SolrDiscuss.Query(q)
		if err != nil {
			log.Println("solr error:", err)
			redirect = "/"
			return
		}
		q = fmt.Sprintf(`post:"%s" title:"%s"`, query, query)
		pr, err := shared.SolrPosts.Query(q)
		if err != nil {
			log.Println("solr error:", err)
			redirect = "/"
			return
		}
		res := Results{
			NumFound: dr.Response.NumFound + pr.Response.NumFound,
			Start: dr.Response.Start,
			Discussions: discussion.ParseDiscussions(dr),
			Posts: discussion.ParsePosts(pr),
			Query: query,
		}
		body = new(shared.Body)
		body.Breadcrumbs = &shared.Breadcrumbs {
			Labels: []string{"Results"},
			Uris: []string{""},
		}
		body.ContentData = res
		files = append(files, "./templates/home/search.tpl")
	} else {
		redirect = "/"
	}
	return
}