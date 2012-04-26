package shared

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"code.google.com/p/gorilla/sessions"
	"github.com/simonz05/godis"
	"github.com/bmizerany/noeq.go"

	"solr"
)

var (
	RedisClient	= godis.New("", 0, "")
	NoeqClient	*noeq.Client
	SolrDiscuss	= solr.New("localhost", 8080, "discuss")
	SolrPosts	= solr.New("localhost", 8080, "posts")
	Templates	= []string{
		"./templates/wrapper.tpl",
		"./templates/header.tpl",
		"./templates/body.tpl",
		"./templates/footer.tpl",
	}
)

type Page struct {
	Header string
	Body *Body
	Footer string
}

type Body struct {
	Breadcrumbs *Breadcrumbs
	UserData *UserData
	Subscribed *Subscribed
	ContentData interface{}
	NoSidebar bool
}

type Breadcrumbs struct {
	Labels	[]string
	Uris		[]string
}

type Subscribed struct {
	Labels	[]string
	Uris		[]string
}

type UserData struct {
	Id			uint64
	Username	string
}

type DiscussionDoc struct {
	Id			uint64	`json:"id"`
	Uri			[]string	`json:"uri"`
	Title			string 	`json:"title"`
	Description	string	`json:"desc"`
	Keywords	string	`json:"keywords"`
	Subscribed	int64 	`json:"subscribed"`
	Topics		int64 	`json:"topics"`
}

type PostDoc struct {
	Id		uint64	`json:"id"`
	DId		uint64	`json:"d_id"`
	TId		uint64	`json:"t_id"`
	PId		uint64	`json:"p_id,omitempty"`
	Title		string	`json:"title"`
	Post	string	`json:"post"`
}

func init() {
	runtime.GOMAXPROCS(2)
	var err error
	NoeqClient, err = noeq.New("", ":4444")
	if err != nil {
		log.Println ("failed to create noeq client", err)
		os.Exit(1)
	}
	_, err = NoeqClient.GenOne()
	if err != nil {
		log.Println ("failed to create noeq client", err)
		os.Exit(1)
	}
}

func getSubscribed(sess *sessions.Session) *Subscribed {
	var sub = new(Subscribed)
	var id uint64
	if _, ok := sess.Values["id"]; ok {
		id = sess.Values["id"].(uint64)
	}
	if id == 0 {
		return sub
	}
	key := fmt.Sprintf("user:%d:subscribed", id)
	se, rerr := RedisClient.Smembers(key)
	if rerr != nil {
		return sub
	}
	for _, d := range se.IntArray() {
		t, _ := RedisClient.Get(fmt.Sprintf("discussion:%d:title", d))
		u, _ := RedisClient.Get(fmt.Sprintf("discussion:%d:uri", d))
		if t != nil && u != nil {
			sub.Labels, sub.Uris = append(sub.Labels, t.String()), append(sub.Uris, "/discuss/" + u.String())
		}
	}
	return sub
}

func getUserData(sess *sessions.Session) *UserData {
	var data = new(UserData)
	var id uint64
	if _, ok := sess.Values["id"]; ok {
		id = sess.Values["id"].(uint64)
	}
	if id == 0 {
		return data
	}
	ue, rerr := RedisClient.Get(fmt.Sprintf("user:%d:username", id))
	if rerr != nil {
		log.Println("redis err", rerr)
		return data
	}
	data.Id = id
	data.Username = ue.String()
	return data
}

func GetTopicBreadcrumbs(id uint64) ([]string, []string) {
	d, rerr := RedisClient.Get(fmt.Sprintf("topic:%d:d_id", id))
	if rerr != nil {
		return nil, nil
	}
	d_id := uint64(d.Int64())
	return uriTree(d_id)
}

func GetDiscussionBreadcrumbs(id uint64, removeLast bool) ([]string, []string) {
	ls, us := uriTree(id)
	l := len(us)
	if removeLast && l > 0 {
		us[len(us)-1] = ""
	}
	return ls, us
}

func uriTree(id uint64) (titles, uris []string) {
	title, rerr := RedisClient.Get(fmt.Sprintf("discussion:%d:title", id))
	if rerr != nil || title.String() == "" {
		return
	}
	uri, rerr := RedisClient.Get(fmt.Sprintf("discussion:%d:uri", id))
	if rerr != nil {
		return
	}
	top_title := title.String()
	top_uri := uri.String()
	parts := strings.Split(uri.String(), "/")
	if len(parts) <= 1 {
		titles = append(titles, top_title)
		uris = append(uris, "/discuss/" + top_uri)
		return
	}
	for i := 1; i < len(parts); i++ {
		d, rerr := RedisClient.Get(fmt.Sprintf("discussions:%s", strings.Join(parts[:i], "/")))
		if rerr != nil {
			return
		}
		id := uint64(d.Int64())
		if id > 0 {
			uri, rerr := RedisClient.Get(fmt.Sprintf("discussion:%d:uri", id))
			if rerr != nil {
				return
			}
			title, rerr := RedisClient.Get(fmt.Sprintf("discussion:%d:title", id))
			if rerr != nil {
				return
			}
			titles = append(titles, title.String())
			uris = append(uris, "/discuss/" + uri.String())
		}
	}
	titles = append(titles, top_title)
	uris = append(uris, "/discuss/" + top_uri)
	return
}

func GetPermission(id, o uint64) uint64 {
	key := fmt.Sprintf("user:%d:permission:%d", id, o)
	p, _ := RedisClient.Get(key)
	return uint64(p.Int64())
}

func CanDo(id, o uint64, action int) bool {
	key := fmt.Sprintf("user:%d:permission:%d", id, o)
	e, _  := RedisClient.Exists(key)
	if !e {
		// set the default
		d, _ := RedisClient.Get(fmt.Sprintf("permission:%d", o))
		RedisClient.Setnx(key, d.String())
	}
	cando, rerr := RedisClient.Getbit(key, action)
	if rerr != nil || cando == 0 {
		return false
	}
	return true
}