package shared

import (
	"fmt"
	"net/http"
	"sync"
	"strconv"

	"code.google.com/p/gorilla/securecookie"
	"code.google.com/p/gorilla/sessions"
	"github.com/dchest/passwordhash"
)

func GetSession(r *http.Request, key string) (*sessions.Session, error) {
	return getSession(r, key)
}

func getSession(r *http.Request, key string) (*sessions.Session, error) {
	s, err := sessionStore.Get(r, key)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func visited(s *sessions.Session) error {
	key := "session:" + s.ID
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	_, rerr := RedisClient.Expire(key, int64(sessionExpire))
	if rerr != nil {
		return rerr
	}
	return nil
}

func killSession(r *http.Request, w http.ResponseWriter, s *sessions.Session) error {
	var opts = *sessionStore.Options
	opts.MaxAge = -1
	s.Options = &opts
	// values
	for k, _ := range s.Values {
		delete(s.Values, k)
	}
	key := "session:" + s.ID
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	_, rerr := RedisClient.Del(key)
	if rerr != nil {
		return rerr
	}
	s.Save(r, w)
	return nil
}

func Remember(r *http.Request, w http.ResponseWriter, id uint64) error {
	if id == 0 {
		return nil
	}
	session, err := sessionStore.New(r, "remember")
	if err != nil {
		return err
	}
	se, rerr := RedisClient.Get(fmt.Sprintf("user:%d:password", id))
	if rerr != nil {
		return rerr
	}
	ph := passwordhash.NewSaltIter(se.String(), rememberKey, passwordhash.DefaultIterations)
	values := make([]interface{}, 2)
	values[0] = id
	values[1] = string(ph.Hash)
	encoded, err := securecookie.EncodeMulti(session.Name(), values, sessionStore.Codecs...)
	if err != nil {
		return err
	}
	cookie := &http.Cookie{
		Name:     session.Name(),
		Value:    encoded,
		Path:     rememberOpts.Path,
		Domain:   rememberOpts.Domain,
		MaxAge:   rememberOpts.MaxAge,
		Secure:   rememberOpts.Secure,
		HttpOnly: rememberOpts.HttpOnly,
	}
	http.SetCookie(w, cookie)
	return nil
}

func Regen(r *http.Request) (uint64, error) {
	name := "remember"
	c, err := r.Cookie(name)
	if err != nil {
		if err == http.ErrNoCookie {
			return 0, nil
		}
		fmt.Println("cookie err", err)
		return 0, err
	}
	vals := make([]interface{}, 2)
	err = securecookie.DecodeMulti(name, c.Value, &vals, sessionStore.Codecs...)
	if err != nil {
		return 0, err
	}
	id := vals[0].(uint64)
	se, rerr := RedisClient.Get(fmt.Sprintf("user:%d:password", id))
	if rerr != nil {
		return 0, rerr
	}
	ph := passwordhash.NewSaltIter(se.String(), rememberKey, passwordhash.DefaultIterations)
	if string(ph.Hash) == vals[1].(string) {
		return id, nil
	}
	return 0, nil
}

// redisStore ------------------------------------------------------------

var sessionMutex sync.RWMutex
var sessionExpire = 300
var rememberExpire = 31536000
var rememberOpts = &sessions.Options{
	Path:   "/",
	MaxAge: rememberExpire,
	HttpOnly: true,
}

var storeKey = []byte("")
var rememberKey = []byte("")

var sessionStore	= newRedisStore(storeKey)

func newRedisStore(keyPairs ...[]byte) *redisStore {
	return &redisStore{
		Codecs: securecookie.CodecsFromPairs(keyPairs...),
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: sessionExpire,
		},
	}
}

type redisStore struct {
	Codecs  []securecookie.Codec
	Options *sessions.Options // default configuration
}

// Get returns a session for the given name after adding it to the registry.
//
// See CookieStore.Get().
func (s *redisStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

// New returns a session for the given name without adding it to the registry.
//
// See CookieStore.New().
func (s *redisStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(s, name)
	session.IsNew = true
	var err error
	if c, errCookie := r.Cookie(name); errCookie == nil {
		err = securecookie.DecodeMulti(name, c.Value, &session.ID, s.Codecs...)
		if err == nil {
			err = s.load(session)
			if err == nil {
				session.IsNew = false
			}
		}
	}
	return session, err
}

// Save adds a single session to the response.
func (s *redisStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	if session.ID == "" {
		i, err := NoeqClient.GenOne()
		if err != nil {
			return err
		}
		session.ID = strconv.FormatUint(i, 10)
	}
	if err := s.save(session); err != nil {
		return err
	}
	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, s.Codecs...)
	if err != nil {
		return err
	}
	options := s.Options
	if session.Options != nil {
		options = session.Options
	}
	cookie := &http.Cookie{
		Name:     session.Name(),
		Value:    encoded,
		Path:     options.Path,
		Domain:   options.Domain,
		MaxAge:   options.MaxAge,
		Secure:   options.Secure,
		HttpOnly: options.HttpOnly,
	}
	http.SetCookie(w, cookie)
	return nil
}

// save writes encoded session.Values to a file.
func (s *redisStore) save(session *sessions.Session) error {
	if session.Name() == "remember" {
		return nil
	}
	if len(session.Values) == 0 {
		// Don't need to write anything.
		return nil
	}
	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values, s.Codecs...)
	if err != nil {
		return err
	}
	key := "session:" + session.ID
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	rerr := RedisClient.Setex(key, int64(sessionExpire), encoded)
	if rerr != nil {
		return rerr
	}
	return nil
}

// load reads a file and decodes its content into session.Values.
func (s *redisStore) load(session *sessions.Session) error {
	if session.Name() == "remember" {
		return nil
	}
	key := "session:" + session.ID
	se, rerr := RedisClient.Get(key)
	if rerr != nil {
		return rerr
	}
	ss := se.String()
	if ss == "" {
		return nil
	}
	err := securecookie.DecodeMulti(session.Name(), ss, &session.Values, s.Codecs...)
	if err != nil {
		return err
	}
	return nil
}