package user

import (
	"fmt"
	_ "log"
	"net/http"
	"time"

	_ "github.com/bmizerany/noeq.go"
	"github.com/dchest/passwordhash"

	. "discuss/shared"
)

// global
const (
	G_LOGIN = iota
)

// topic
const (
	T_VIEW = iota
	T_ADD
	T_UPDATE
	T_DELETE
	T_MODERATE
)

/*

USER RECORD
user:<username>
user:<u_id>:username
user:<u_id>:password
user:<u_id>:salt
user:<u_id>:email
user:<u_id>:lastlogin
user:<u_id>:lastvisit

PERMISSIONS -- BITMASKS

user:<u_id>:permissions:global
user:<u_id>:permissions:discussion:<d_id>
user:<u_id>:permissions:topic:<t_id>
...
etc

*/

func add(r *http.Request) (uint64, error) {
	// get an id
	id, err := NoeqClient.GenOne()
	if err != nil {
		return 0, err
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	confirm := r.FormValue("confirm_password")
	email := r.FormValue("email")
	if username == "" || password == "" || confirm == "" || password != confirm || email == "" {
		return 0, fmt.Errorf("invalid registration details")
	}

	// try to set it
	res, rerr := RedisClient.Setnx("users:"+username, id)
	if rerr != nil {
		return 0, rerr
	} else if !res {
		// someone beat us to it
		return 0, fmt.Errorf("username already exists")
	}

	// set the discussion record
	ph := passwordhash.New(password)
	var userData = make(map[string]string)
	userData[fmt.Sprintf("user:%d:username", id)] = username
	userData[fmt.Sprintf("user:%d:password", id)] = string(ph.Hash)
	userData[fmt.Sprintf("user:%d:salt", id)] = string(ph.Salt)
	userData[fmt.Sprintf("user:%d:email", id)] = email
	rerr = RedisClient.Mset(userData)
	if rerr != nil {
		// roll it back
		return 0, rerr
	}
	return id, nil
}

func checkRemember(r *http.Request) (uint64, error) {
	return 0, nil
}

func authenticate(r *http.Request) (uint64, error) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	if username == "" || password == "" {
		return 0, fmt.Errorf("bad login")
	}
	// get the u_id
	ue, rerr := RedisClient.Get("users:" + username)
	id := uint64(ue.Int64())
	if rerr != nil {
		return 0, rerr
	} else if id == 0 {
		return 0, fmt.Errorf("username not found")
	}
	// get credentials
	se, rerr := RedisClient.Get(fmt.Sprintf("user:%d:salt", id))
	if rerr != nil {
		return 0, rerr
	}
	pe, rerr := RedisClient.Get(fmt.Sprintf("user:%d:password", id))
	if rerr != nil {
		return 0, rerr
	}
	ph := passwordhash.PasswordHash{
		passwordhash.DefaultIterations,
		se.Bytes(),
		pe.Bytes(),
	}
	if password == "" {
		return 0, fmt.Errorf("invalid password")
	}
	// check it
	if !ph.EqualToPassword(password) {
		return 0, fmt.Errorf("invalid password")
	}
	rerr = RedisClient.Set(fmt.Sprintf("user:%d:lastlogin", id), time.Now().Unix())
	if rerr != nil {
		return 0, rerr
	}
	return id, nil
}

func get(parts []string) string {
	return fmt.Sprintf("DO A USER LOOKUP: %q\n", parts)
}

func getById(id uint64) string {
	return fmt.Sprintf("DO A USER ID LOOKUP: %q\n", id)
}
