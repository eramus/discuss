package discussion

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/simonz05/godis"
	"solr"
			
	. "discuss/shared"
	"discuss/topic"
)

const (
	VIEW = iota		// 0
	BANNED			// 1
	UPDATE			// 2
	MODERATE		// 3
	ADD_TOPIC		// 4
	UPDATE_TOPIC	// 5
	DELETE_TOPIC	// 6
	ADD_POST		// 7
	UPDATE_POST	// 8
	DELETE_POST	// 9
)

func GetId(uri string) (uint64, error) {
	e, err := RedisClient.Get("discussions:" + uri)
	if err != nil {
		return 0, err
	}
	id := e.Int64()
	if id > 0 {
		// discussion found
		return uint64(id), nil
	}
	return 0, nil
}

func GetUri(id uint64) (string, error) {
	ie, rerr := RedisClient.Get(fmt.Sprintf("discussion:%d:uri", id))
	if rerr != nil {
		return "", rerr
	}
	return ie.String(), nil
}

func Add(r *http.Request, u_id uint64) (uint64, error) {
	// get an id
	id, err := NoeqClient.GenOne()
	if err != nil {
		log.Println("noeq err", err)
		return 0, err
	}
	// get the data
	uri := r.FormValue("uri")
	title := r.FormValue("title")
	description := r.FormValue("description")
	keywords := r.FormValue("keywords")
	if len(keywords) == 0 {
		keywords = uri
	}
	// try to set it
	res, rerr := RedisClient.Setnx("discussions:" + uri, id)
	if rerr != nil {
		log.Println("redis err", rerr)
		return 0, rerr
	} else if !res {
		// someone beat us to it
		log.Println("exists err", err)
		return GetId(uri)
	}
	// add to solr
	var ds []interface{}
	d := &DiscussionDoc {
		Id: id,
		Uri: strings.Split(uri, "/"),
		Title: title,
		Description: description,
		Keywords: keywords,
	}
	ds = append(ds, d)
	_, err = SolrDiscuss.Update(ds)
	if err != nil {
		// roll it back
		log.Println("solr err", err)
		RedisClient.Del("discussions:" + uri)
		return 0, rerr
	}
	prc := godis.NewPipeClientFromClient(RedisClient)
	prc.Multi()
	// set the discussion record
	var discussionData = make(map[string]string)
	discussionData[fmt.Sprintf("discussion:%d:uri", id)] = uri
	discussionData[fmt.Sprintf("discussion:%d:u_id", id)] = strconv.FormatUint(u_id, 10)
	discussionData[fmt.Sprintf("discussion:%d:title", id)] = title
	discussionData[fmt.Sprintf("discussion:%d:description", id)] = description
	discussionData[fmt.Sprintf("discussion:%d:keywords", id)] = keywords
	rerr = prc.Mset(discussionData)
	if rerr != nil {
		// roll it back
		prc.Discard()
		RedisClient.Del("discussions:" + uri)
		// roll back solr too
		return 0, rerr
	}
	err = defaultDiscussion(prc, id)
	if err != nil {
		prc.Discard()
		RedisClient.Del("discussions:" + uri)
		// roll back solr too
		return 0, rerr
	}
	err = defaultCreator(prc, id, u_id)
	if err != nil {
		prc.Discard()
		RedisClient.Del("discussions:" + uri)
		// roll back solr too
		return 0, rerr
	}
	prc.Exec()
	return id, nil
}

func Topics(id uint64) ([]topic.Topic, error) {
	ts, err := RedisClient.Zrevrangebyscore(fmt.Sprintf("discussion:%d:topics", id), 10000, 800, "limit", "0", "10")
	if err != nil {
		return nil, err
	}
	var topics []topic.Topic
	for _, e := range ts.Elems {
		t_id := uint64(e.Elem.Int64())

		var keys = make([]string, 6)
		keys[0] = fmt.Sprintf("topic:%d:title", t_id)
		keys[1] = fmt.Sprintf("topic:%d:lastpost", t_id)
		keys[2] = fmt.Sprintf("topic:%d:u_id", t_id)
		keys[3] = fmt.Sprintf("topic:%d:last_p_id", t_id)
		keys[4] = fmt.Sprintf("topic:%d:last_u_id", t_id)
		keys[5] = fmt.Sprintf("topic:%d:numposts", t_id)
		fs, rerr := RedisClient.Mget(keys...)
		if rerr != nil {
			return nil, rerr
		}
		t := topic.Topic{
			Id: t_id,
			DId: id,
			Title: fs.Elems[0].Elem.String(),
			LastPost: FormatTime(fs.Elems[1].Elem.Int64()),
			LastPostId: uint64(fs.Elems[2].Elem.Int64()),
			NumPosts: fs.Elems[5].Elem.Int64(),
		}
		topics = append(topics, t)
	}
	return topics, nil
}

func ParseDiscussions(sr *solr.SolrResponse) (docs []DiscussionDoc) {
	if len(sr.Response.Docs) > 0 {
		for _, i := range sr.Response.Docs {
			doc := i.(map[string]interface{})
			id, _ := strconv.ParseInt(doc["id"].(string), 10, 64)
			d := DiscussionDoc {
				Id: uint64(id),
				Title: doc["title"].(string),
			}
			for _, u := range doc["uri"].([]interface{}) {
				d.Uri = append(d.Uri, u.(string))
			}
			keys := make([]string, 3)
			keys[0] = fmt.Sprintf("discussion:%d:description", uint64(id))
			keys[1] = fmt.Sprintf("discussion:%d:numtopics", uint64(id))
			keys[2] = fmt.Sprintf("discussion:%d:subscribed", uint64(id))
			fs, rerr := RedisClient.Mget(keys...)
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

func ParsePosts(sr *solr.SolrResponse) (docs []PostDoc) {
	if len(sr.Response.Docs) > 0 {
		for _, i := range sr.Response.Docs {
			doc := i.(map[string]interface{})
			id, _ := strconv.ParseInt(doc["id"].(string), 10, 64)
			keys := make([]string, 2)
			keys[0] = fmt.Sprintf("post:%d:post", uint64(id))
			keys[1] = fmt.Sprintf("post:%d:t_id", uint64(id))
			fs, rerr := RedisClient.Mget(keys...)
			if rerr != nil {
				return
			}
			d := PostDoc {
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

func defaultDiscussion(prc *godis.PipeClient, id uint64) error {
	key := fmt.Sprintf("permission:%d", id)
	_, rerr := prc.Setbit(key, VIEW, 1)
	if rerr != nil {
		return rerr
	}
	_, rerr = prc.Setbit(key, ADD_TOPIC, 1)
	if rerr != nil {
		return rerr
	}
	_, rerr = prc.Setbit(key, UPDATE_TOPIC, 1)
	if rerr != nil {
		return rerr
	}
	_, rerr = prc.Setbit(key, DELETE_TOPIC, 1)
	if rerr != nil {
		return rerr
	}
	_, rerr = prc.Setbit(key, ADD_POST, 1)
	if rerr != nil {
		return rerr
	}
	_, rerr = prc.Setbit(key, UPDATE_POST, 1)
	if rerr != nil {
		return rerr
	}
	_, rerr = prc.Setbit(key, DELETE_POST, 1)
	if rerr != nil {
		return rerr
	}
	return nil
}

func defaultCreator(prc *godis.PipeClient, id, u_id uint64) error {
	key := fmt.Sprintf("user:%d:permission:%d", u_id, id)
	_, rerr := prc.Setbit(key, VIEW, 1)
	if rerr != nil {
		return rerr
	}
	_, rerr = prc.Setbit(key, UPDATE, 1)
	if rerr != nil {
		return rerr
	}
	_, rerr = prc.Setbit(key, MODERATE, 1)
	if rerr != nil {
		return rerr
	}
	_, rerr = prc.Setbit(key, ADD_TOPIC, 1)
	if rerr != nil {
		return rerr
	}
	_, rerr = prc.Setbit(key, UPDATE_TOPIC, 1)
	if rerr != nil {
		return rerr
	}
	_, rerr = prc.Setbit(key, DELETE_TOPIC, 1)
	if rerr != nil {
		return rerr
	}
	_, rerr = prc.Setbit(key, ADD_POST, 1)
	if rerr != nil {
		return rerr
	}
	_, rerr = prc.Setbit(key, UPDATE_POST, 1)
	if rerr != nil {
		return rerr
	}
	_, rerr = prc.Setbit(key, DELETE_POST, 1)
	if rerr != nil {
		return rerr
	}
	return nil
}