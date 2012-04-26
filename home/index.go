package home

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"code.google.com/p/gorilla/sessions"
	"solr"

	"discuss/shared"
)

var indexTpls = template.Must(template.ParseFiles(
		append(shared.Templates, "./templates/home/home.tpl")...
))
var searchTpls = template.Must(template.ParseFiles(
		append(shared.Templates, "./templates/home/search.tpl")...
))

type Results struct {
	NumFound	int
	Start		int
	Discussions	[]shared.DiscussionDoc
	Posts		[]shared.PostDoc
	Query		string
}

func Index(r *http.Request, sess *sessions.Session) (body *shared.Body, tpl *template.Template, redirect string) {
//	log.Println("route: index")
	tpl = indexTpls
	return
}

func Search(r *http.Request, sess *sessions.Session) (body *shared.Body, tpl *template.Template, redirect string) {
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
			Discussions: parseDiscussions(dr),
			Posts: parsePosts(pr),
			Query: query,
		}
		body = new(shared.Body)
		body.Breadcrumbs = &shared.Breadcrumbs {
			Labels: []string{"Results"},
			Uris: []string{""},
		}
		body.ContentData = res
		tpl = searchTpls
	} else {
		redirect = "/"
	}
	return
}

func parseDiscussions(sr *solr.SolrResponse) (docs []shared.DiscussionDoc) {
	if len(sr.Response.Docs) > 0 {
		keys := make([]string, 3)
		for _, i := range sr.Response.Docs {
			doc := i.(map[string]interface{})
			id, _ := strconv.ParseInt(doc["id"].(string), 10, 64)
			d := shared.DiscussionDoc {
				Id: uint64(id),
				Title: doc["title"].(string),
			}
			for _, u := range doc["uri"].([]interface{}) {
				d.Uri = append(d.Uri, u.(string))
			}
			keys[0] = fmt.Sprintf("discussion:%d:description", uint64(id))
			keys[1] = fmt.Sprintf("discussion:%d:numtopics", uint64(id))
			keys[2] = fmt.Sprintf("discussion:%d:subscribed", uint64(id))
			fs, rerr := shared.RedisClient.Mget(keys...)
			if rerr != nil {
				return
			}
			d.Description = fs.Elems[0].Elem.String()
			d.Topics = fs.Elems[1].Elem.Int64()
			d.Subscribed = fs.Elems[2].Elem.Int64()
			docs = append(docs, d)
		}
	}
	return
}

func parsePosts(sr *solr.SolrResponse) (docs []shared.PostDoc) {
	if len(sr.Response.Docs) > 0 {
		keys := make([]string, 2)
		for _, i := range sr.Response.Docs {
			doc := i.(map[string]interface{})
			id, _ := strconv.ParseInt(doc["id"].(string), 10, 64)
			keys[0] = fmt.Sprintf("post:%d:post", uint64(id))
			keys[1] = fmt.Sprintf("post:%d:t_id", uint64(id))
			fs, rerr := shared.RedisClient.Mget(keys...)
			if rerr != nil {
				return
			}
			d := shared.PostDoc {
				Id: uint64(id),
				TId: uint64(fs.Elems[1].Elem.Int64()),
				Title: doc["title"].(string),
				Post: fs.Elems[0].Elem.String(),
			}
			docs = append(docs, d)
		}
	}
	return
}