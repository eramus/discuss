package topic

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/simonz05/godis"

	"discuss/post"
	"discuss/shared"
)

func get(uri string) (uint64, error) {
	fmt.Printf("DO A TOPIC PERMALINK LOOK: %q\n", uri)
	return 0, nil
}

func getById(id uint64) (*Topic, error) {
	keys := make([]string, 3)
	keys[0] = fmt.Sprintf("topic:%d:d_id", uint64(id))
	keys[1] = fmt.Sprintf("topic:%d:title", uint64(id))
	keys[2] = fmt.Sprintf("topic:%d:score", uint64(id))
	fs, rerr := shared.RedisClient.Mget(keys...)
	if rerr != nil {
		return nil, rerr
	}
	uc, rerr := shared.RedisClient.Scard(fmt.Sprintf("topic:%d:users", id))
	if rerr != nil {
		return nil, rerr
	}
	pse, rerr := shared.RedisClient.Zrevrangebyscore(fmt.Sprintf("topic:%d:posts", id), 10000, 800, "limit", "0", "10")
	if rerr != nil {
		return nil, rerr
	}
	res := pse.IntArray()
	l := len(res)
	var posts []*post.Post
	if l > 0 {
		posts = make([]*post.Post, l)
		for i, pe := range res {
			p_id := uint64(pe)
			p, err := post.GetById(p_id, id, 5)
			if err != nil {
				log.Println("err 5", err)
				return nil, err
			}
			posts[i] = p
		}
	}
	return &Topic{
		Id:    id,
		DId:   uint64(fs.Elems[0].Elem.Int64()),
		Title: fs.Elems[1].Elem.String(),
		Score: fs.Elems[2].Elem.Int64(),
		Posts: posts,
		Users: uc,
	}, nil
}

func exists(id uint64) bool {
	val, err := shared.RedisClient.Get(fmt.Sprintf("topic:%d:title", id))
	if err != nil || val.String() == "" {
		return false
	}
	return true
}

func add(r *http.Request, u_id uint64) (uint64, error) {
	// get an id
	ids, err := shared.NoeqClient.Gen(2)
	if err != nil {
		return 0, err
	}
	t_id := ids[0]
	p_id := ids[1]
	ts := strconv.FormatInt(time.Now().Unix(), 10)

	// get the data
	ds_id := r.FormValue("d_id")
	title := r.FormValue("title")
	post := r.FormValue("post")

	if ds_id == "" {
		return 0, fmt.Errorf("Need a discussion id")
	} else if post == "" {
		return 0, fmt.Errorf("Need a post")
	}
	d_id, err := strconv.ParseUint(ds_id, 10, 64)
	if err != nil {
		return 0, err
	}

	// add to solr
	var ps []interface{}
	p := &shared.PostDoc{
		Id:    p_id,
		DId:   d_id,
		TId:   t_id,
		Title: title,
		Post:  post,
	}
	ps = append(ps, p)
	_, err = shared.SolrPosts.Update(ps)
	if err != nil {
		log.Println("SOLR ERROR:", err)
		// roll it back
		return 0, err
	}

	// set the discussion record
	prc := godis.NewPipeClientFromClient(shared.RedisClient)
	prc.Multi()

	var topicData = make(map[string]string)
	topicData[fmt.Sprintf("topic:%d:d_id", t_id)] = ds_id
	topicData[fmt.Sprintf("topic:%d:u_id", t_id)] = strconv.FormatUint(u_id, 10)
	topicData[fmt.Sprintf("topic:%d:last_u_id", t_id)] = strconv.FormatUint(u_id, 10)
	topicData[fmt.Sprintf("topic:%d:title", t_id)] = title
	topicData[fmt.Sprintf("topic:%d:last_p_id", t_id)] = strconv.FormatUint(p_id, 10)
	topicData[fmt.Sprintf("topic:%d:lastpost", t_id)] = ts
	topicData[fmt.Sprintf("topic:%d:score", t_id)] = "1000"
	topicData[fmt.Sprintf("post:%d:d_id", p_id)] = ds_id
	topicData[fmt.Sprintf("post:%d:t_id", p_id)] = strconv.FormatUint(t_id, 10)
	topicData[fmt.Sprintf("post:%d:u_id", p_id)] = strconv.FormatUint(u_id, 10)
	topicData[fmt.Sprintf("post:%d:post", p_id)] = post
	topicData[fmt.Sprintf("post:%d:ts", p_id)] = ts
	topicData[fmt.Sprintf("post:%d:score", p_id)] = "1000"
	rerr := prc.Mset(topicData)
	if rerr != nil {
		// roll it back
		prc.Discard()
		return 0, rerr
	}
	// go routine?
	_, rerr = prc.Zadd(fmt.Sprintf("topic:%d:posts", t_id), 1000, p_id)
	if rerr != nil {
		// roll it back
		prc.Discard()
		return 0, rerr
	}
	_, rerr = prc.Incr(fmt.Sprintf("topic:%d:numposts", t_id))
	if rerr != nil {
		// roll it back
		prc.Discard()
		return 0, rerr
	}
	_, rerr = prc.Zadd(fmt.Sprintf("discussion:%d:topics", d_id), 1000, t_id)
	if rerr != nil {
		// roll it back
		prc.Discard()
		return 0, rerr
	}
	_, rerr = prc.Incr(fmt.Sprintf("discussion:%d:numtopics", d_id))
	if rerr != nil {
		// roll it back
		prc.Discard()
		return 0, rerr
	}
	prc.Sadd(fmt.Sprintf("topic:%d:users", t_id), u_id)
	// update some user stats
	prc.Sadd(fmt.Sprintf("user:%d:topics:%d", u_id, d_id), t_id)
	prc.Sadd(fmt.Sprintf("user:%d:posts:%d", u_id, t_id), p_id)
	// /go routine?
	prc.Exec()
	return t_id, nil
}
