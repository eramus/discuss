package discussion

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/simonz05/godis"

	. "discuss/shared"
	"discuss/topic"
)

const (
	VIEW         = iota // 0
	BANNED              // 1
	UPDATE              // 2
	MODERATE            // 3
	ADD_TOPIC           // 4
	UPDATE_TOPIC        // 5
	DELETE_TOPIC        // 6
	ADD_POST            // 7
	UPDATE_POST         // 8
	DELETE_POST         // 9
)

func getId(uri string) (uint64, error) {
	e, err := RedisClient.Get("discussions:" + uri)
	if err != nil {
		return 0, err
	}
	id := uint64(e.Int64())
	if id > 0 {
		// discussion found
		return id, nil
	}
	return 0, nil
}

func getUri(id uint64) (string, error) {
	ie, rerr := RedisClient.Get(fmt.Sprintf("discussion:%d:uri", id))
	if rerr != nil {
		return "", rerr
	}
	return ie.String(), nil
}

func add(r *http.Request, u_id uint64) (uint64, error) {
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
	res, rerr := RedisClient.Setnx("discussions:"+uri, id)
	if rerr != nil {
		log.Println("redis err", rerr)
		return 0, rerr
	} else if !res {
		// someone beat us to it
		log.Println("exists err", err)
		return getId(uri)
	}
	// add to solr
	ds := make([]interface{}, 1)
	ds[1] = &DiscussionDoc{
		Id:          id,
		Uri:         strings.Split(uri, "/"),
		Title:       title,
		Description: description,
		Keywords:    keywords,
	}
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
	err = defaultCreator(prc, u_id, id)
	if err != nil {
		prc.Discard()
		RedisClient.Del("discussions:" + uri)
		// roll back solr too
		return 0, rerr
	}
	prc.Exec()
	return id, nil
}

func topics(id uint64) ([]*topic.Topic, error) {
	ts, err := RedisClient.Zrevrangebyscore(fmt.Sprintf("discussion:%d:topics", id), 10000, 800, "limit", "0", "10")
	if err != nil {
		return nil, err
	}
	var topics = make([]*topic.Topic, len(ts.Elems))
	var keys = make([]string, 7)
	for i, e := range ts.Elems {
		t_id := uint64(e.Elem.Int64())

		keys[0] = fmt.Sprintf("topic:%d:title", t_id)
		keys[1] = fmt.Sprintf("topic:%d:lastpost", t_id)
		keys[2] = fmt.Sprintf("topic:%d:u_id", t_id)
		keys[3] = fmt.Sprintf("topic:%d:last_p_id", t_id)
		keys[4] = fmt.Sprintf("topic:%d:last_u_id", t_id)
		keys[5] = fmt.Sprintf("topic:%d:numposts", t_id)
		keys[6] = fmt.Sprintf("topic:%d:score", t_id)
		fs, rerr := RedisClient.Mget(keys...)
		if rerr != nil {
			return nil, rerr
		}
		uc, rerr := RedisClient.Scard(fmt.Sprintf("topic:%d:users", t_id))
		if rerr != nil {
			return nil, rerr
		}
		t := &topic.Topic{
			Id:         t_id,
			DId:        id,
			Title:      fs.Elems[0].Elem.String(),
			LastPost:   FormatTime(fs.Elems[1].Elem.Int64()),
			LastPostId: uint64(fs.Elems[2].Elem.Int64()),
			NumPosts:   fs.Elems[5].Elem.Int64(),
			Score:      fs.Elems[6].Elem.Int64(),
			Users:      uc,
		}
		topics[i] = t
	}
	return topics, nil
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

func defaultCreator(prc *godis.PipeClient, u_id, id uint64) error {
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
