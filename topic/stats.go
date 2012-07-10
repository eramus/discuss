package topic

import (
	"fmt"
	"log"
	"time"

	"discuss/shared"
)

const runUpdate = 3
const forceUpdate = 120

type topicView struct {
	d uint64
	t uint64
}

type score struct {
	uid uint64
	tid uint64
	did uint64
	up  bool
}

var topicViews chan topicView
var topicVotes chan score

func init() {
	topicViews = make(chan topicView)
	topicVotes = make(chan score)
	go handleTopicviews()
}

func handleTopicviews() {
	t := time.NewTicker(30 * time.Second)
	done := make(chan bool)
	working := false
	for {
		select {
		case ids, _ := <-topicViews:
			go addTopicView(ids)
		case ids, _ := <-topicVotes:
			go scoreTopic(ids)
		case <-t.C:
			if !working {
				working = true
				go updateViews(done)
			}
		case <-done:
			working = false
		}
	}
}

func addView(d_id, t_id uint64) {
	topicViews <- topicView{d_id, t_id}
}

func addTopicView(ids topicView) {
	key := fmt.Sprintf("discussion:%d:views", ids.d)
	shared.RedisClient.Rpush(key, ids.t)
}

func updateViews(done chan bool) {
	//	log.Println("update thread views")
	ds, _ := shared.RedisClient.Keys("discussion:*:views")
	if len(ds) > 0 {
		var (
			err         error
			viewMap     = make(map[uint64]int64)
			l           int64
			id          uint64
			ok          bool
			key, newKey string
			now         = time.Now().Unix()
		)
		for _, key := range ds {
			l, _ = shared.RedisClient.Llen(key)
			if l == 0 {
				continue
			}
			if l < runUpdate {
				e, _ := shared.RedisClient.Get(key + ":updated")
				if e != nil && now-e.Int64() < forceUpdate {
					continue
				}
			}
			// rename, get & del
			newKey = key + ":updating"
			err = shared.RedisClient.Rename(key, newKey)
			if err != nil {
				// hmm
				log.Println("hmm")
				continue
			}
			shared.RedisClient.Set(key+":updated", now)
			l, _ = shared.RedisClient.Llen(newKey)
			views, _ := shared.RedisClient.Lrange(newKey, 0, int(l))
			shared.RedisClient.Del(newKey)
			if views == nil {
				continue
			}
			// process them
			for _, e := range views.Elems {
				id = uint64(e.Elem.Int64())
				_, ok = viewMap[id]
				if ok {
					viewMap[id]++
				} else {
					viewMap[id] = 1
				}
			}
		}
		//		log.Println("viewMap:", viewMap)
		for id, cnt := range viewMap {
			key = fmt.Sprintf("topic:%d:views", id)
			shared.RedisClient.Incrby(key, cnt)
		}
	}
	done <- true
}

func bumpTopic(u_id, t_id, d_id uint64) {
	topicVotes <- score{u_id, t_id, d_id, true}
}

func buryTopic(u_id, t_id, d_id uint64) {
	topicVotes <- score{u_id, t_id, d_id, false}
}

func scoreTopic(ids score) {
	var to, from string
	if ids.up {
		to, from = "bumped", "buried"
	} else {
		to, from = "buried", "bumped"
	}
	// try to add user
	key := fmt.Sprintf("topic:%d:%s", ids.tid, to)
	added, _ := shared.RedisClient.Sadd(key, ids.uid)
	if !added {
		// already voted
		return
	}
	shared.RedisClient.Sadd(fmt.Sprintf("user:%d:voted:discussion:%d", ids.uid, ids.did), ids.tid)
	// try to remove from other
	key = fmt.Sprintf("topic:%d:%s", ids.tid, from)
	removed, _ := shared.RedisClient.Srem(key, ids.uid)
	var move float64 = 1
	if removed {
		move++
	}
	if !ids.up {
		move = -move
	}
	// move scores
	key = fmt.Sprintf("discussion:%d:topics", ids.did)
	shared.RedisClient.Zincrby(key, move, ids.tid)
	shared.RedisClient.Incrby(fmt.Sprintf("topic:%d:score", ids.tid), int64(move))
}
