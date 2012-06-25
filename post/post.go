package post

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/simonz05/godis"	

	. "discuss/shared"
)

func GetById(id, t_id uint64, lvl int) (*Post, error) {
	keys := make([]string, 4)
	keys[0] = fmt.Sprintf("post:%d:post", uint64(id))
	keys[1] = fmt.Sprintf("post:%d:u_id", uint64(id))
	keys[2] = fmt.Sprintf("post:%d:ts", uint64(id))
	keys[3] = fmt.Sprintf("post:%d:score", uint64(id))
	fs, rerr := RedisClient.Mget(keys...)
	if rerr != nil {
		return nil, rerr
	}
	ue, rerr := RedisClient.Get(fmt.Sprintf("user:%d:username", uint64(fs.Elems[1].Elem.Int64())))
	if rerr != nil {
		return nil, rerr
	}
	ple, rerr := RedisClient.Zcount(fmt.Sprintf("post:%d:posts", id), 800, 10000)
	if rerr != nil {
		return nil, rerr
	}
	var posts []*Post
	if ple > 0 && lvl > 0 {
		pse, rerr := RedisClient.Zrevrangebyscore(fmt.Sprintf("post:%d:posts", id), 10000, 800, "limit", "0", "10")
		if rerr != nil {
			return nil, rerr
		}
		lvl--
		posts = make([]*Post, len(pse.Elems))
		for i, p_id := range pse.Elems {
			ps, err := GetById(uint64(p_id.Elem.Int64()), t_id, lvl)
			if err != nil {
				return nil, err
			}
			posts[i] = ps
		}
	}
	return &Post{
		Id: id,
		TId: t_id,
		UId: uint64(fs.Elems[1].Elem.Int64()),
		Username: ue.String(),
		Post: fs.Elems[0].Elem.String(),
		Posts: posts,
		Timestamp: uint64(fs.Elems[2].Elem.Int64()),
		FTimestamp: time.Unix(fs.Elems[2].Elem.Int64(), 0).Format("January 01, 2006 @ 15:04"),
		RTimestamp: FormatTime(fs.Elems[2].Elem.Int64()),
		Score: fs.Elems[3].Elem.Int64(),
	}, nil
}

func Add(r *http.Request, u_id uint64) (uint64, error) {
	// get an id
	id, err := NoeqClient.GenOne()
	if err != nil {
		return 0, err
	}
	// get the data
	ts_id := r.FormValue("t_id")
	ps_id := r.FormValue("p_id")
	post := r.FormValue("post")
	if ts_id == "" {
		return 0, fmt.Errorf("Need a thread id")
	} else if post == "" {
		return 0, fmt.Errorf("Need a post")
	}
	var (
		t_id uint64 = 0
		p_id uint64 = 0
	)
	t_id, err = strconv.ParseUint(ts_id, 10, 64)
	if err != nil {
		return 0, err
	}
	if ps_id != "" {
		p_id, err = strconv.ParseUint(ps_id, 10, 64)
		if err != nil {
			return 0, err
		}
	}
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	// get the discussion id for completeness
	de, rerr := RedisClient.Get(fmt.Sprintf("topic:%d:d_id", t_id))
	if rerr != nil {
		return 0, rerr
	}
	d_id := uint64(de.Int64())
	// add to solr
	var ps []interface{}
	p := &PostDoc {
		Id: id,
		DId: d_id,
		TId: t_id,
		Post: post,
	}
	if p_id != 0 {
		p.PId = p_id
	}
	ps = append(ps, p)
	_, err = SolrPosts.Update(ps)
	if err != nil {
		log.Println("SOLR ERROR:", err)
		// roll it back
		return 0, rerr
	}

	// set the discussion record
	prc := godis.NewPipeClientFromClient(RedisClient)
	prc.Multi()

	var postData = make(map[string]string)
	postData[fmt.Sprintf("topic:%d:last_p_id", t_id)] = strconv.FormatUint(id, 10)
	postData[fmt.Sprintf("topic:%d:last_u_id", t_id)] = strconv.FormatUint(u_id, 10)
	postData[fmt.Sprintf("topic:%d:lastpost", t_id)] = ts
	postData[fmt.Sprintf("post:%d:d_id", id)] = de.String()
	postData[fmt.Sprintf("post:%d:t_id", id)] = ts_id
	postData[fmt.Sprintf("post:%d:u_id", id)] = strconv.FormatUint(u_id , 10)
	postData[fmt.Sprintf("post:%d:post", id)] = post
	postData[fmt.Sprintf("post:%d:ts", id)] = ts
	postData[fmt.Sprintf("post:%d:score", id)] = "1000"
	if p_id == 0 {
		// top level post
		_, rerr := prc.Zadd(fmt.Sprintf("topic:%d:posts", t_id), 1000, id)
		if rerr != nil {
			// roll it back
			prc.Discard()
			return 0, rerr
		}
	} else {
		// reply to a post
		postData[fmt.Sprintf("post:%d:p_id", id)] = ps_id
		_, rerr := prc.Zadd(fmt.Sprintf("post:%d:posts", p_id), 1000, id)
		if rerr != nil {
			// roll it back
			prc.Discard()
			return 0, rerr
		}
		_, rerr = prc.Incr(fmt.Sprintf("post:%d:numposts", p_id))
		if rerr != nil {
			// roll it back
			prc.Discard()
			return 0, rerr
		}
		// figure out where to bump this guy
		ks := make([]string, 2)
		ks[0] = fmt.Sprintf("post:%d:t_id", p_id)
		ks[1] = fmt.Sprintf("post:%d:p_id", p_id)
		fs, _ := RedisClient.Mget(ks...)
		pt_id := uint64(fs.Elems[0].Elem.Int64())
		pp_id := uint64(fs.Elems[1].Elem.Int64())
		if pp_id > 0 {
			// a reply also
			prc.Zincrby(fmt.Sprintf("posts:%d:posts", pp_id), 1, p_id)
		} else {
			// top level
			prc.Zincrby(fmt.Sprintf("topic:%d:posts", pt_id), 1, p_id)
		}
	}
	prc.Zincrby(fmt.Sprintf("discussion:%d:topics", d_id), 1, t_id)
	prc.Sadd(fmt.Sprintf("user:%d:posts:%d", u_id, t_id), id)
	prc.Sadd(fmt.Sprintf("topic:%d:users", t_id), u_id)
	_, rerr = prc.Incr(fmt.Sprintf("topic:%d:numposts", t_id))
	if rerr != nil {
		// roll it back
		prc.Discard()
		return 0, rerr
	}
	rerr = prc.Mset(postData)
	if rerr != nil {
		// roll it back
		prc.Discard()
		return 0, rerr
	}
	prc.Exec()
	return id, nil
}