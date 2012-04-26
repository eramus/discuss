package post

import (
	"fmt"
//	"log"

	"discuss/shared"
)

type score struct {
	uid	uint64
	pid	uint64
	up	bool
}

var postVotes chan score

func init() {
	postVotes = make(chan score)
	go handleScores()
}

func handleScores() {
	for {
		select {
		case ids, _ := <-postVotes:
			go scorePost(ids)
		}
	}
}

func BumpPost(u_id, p_id  uint64) {
	postVotes <- score{u_id, p_id, true}
}

func BuryPost(u_id, p_id uint64) {
	postVotes <- score{u_id, p_id, false}
}

func scorePost(ids score) {
//	log.Println("score post:", ids)
	keys := make([]string, 2)
	keys[0] = fmt.Sprintf("post:%d:t_id", ids.pid)
	keys[1] = fmt.Sprintf("post:%d:p_id", ids.pid)
	fs, rerr := shared.RedisClient.Mget(keys...)
	if rerr != nil {
		return
	}
	t_id := uint64(fs.Elems[0].Elem.Int64())
	keyType := "topic"
	keyId := t_id
	if fs.Elems[1].Elem.Int64() != 0 {
		keyType = "post"
		keyId = uint64(fs.Elems[1].Elem.Int64())
	}
	// try to add user buried
	var to, from string
	if ids.up {
		to, from = "bumped", "buried"
	} else {
		to, from = "buried", "bumped"
	}
	// try to add user
	key := fmt.Sprintf("post:%d:%s", ids.pid, to)
	added, _ := shared.RedisClient.Sadd(key, ids.uid)
	if !added {
		// already voted
		return
	}
	shared.RedisClient.Sadd(fmt.Sprintf("user:%d:voted:topic:%d", ids.uid, t_id), ids.pid)
	// try to remove from other
	key = fmt.Sprintf("post:%d:%s", ids.pid, from)
	removed, _ := shared.RedisClient.Srem(key, ids.uid)
	var move float64 = 1
	if removed {
		move++
	}
	if !ids.up {
		move = -move
	}
	// move scores
	key = fmt.Sprintf("%s:%d:posts", keyType, keyId)
	shared.RedisClient.Zincrby(key, move, ids.pid)
	shared.RedisClient.Incrby(fmt.Sprintf("post:%d:score", ids.pid), int64(move))
}