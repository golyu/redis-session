package redis_session

import (
	"encoding/base32"
	"errors"
	"github.com/go-redis/redis"
	"github.com/golyu/redis-session/serializer"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"net/http"
	"strings"
	"time"
)

// RedisStore stores sessions in a redis backend.
type RedisStore struct {
	Pool       *redis.Client
	Codecs     []securecookie.Codec
	Options    *sessions.Options // default configuration
	keyPrefix  string            // key前缀
	maxLength  int               // 数据最大长度
	serializer serializer.ISessionSerializer
}

var (
	defaultSessionOptions = &sessions.Options{
		Path:     "/",
		Domain:   "",
		MaxAge:   86400 * 7, // session expire time ,second
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteDefaultMode,
	}
)

// NewRedisStore instantiates a RedisStore with a *redis.Client passed in.
func NewRedisStore(pool *redis.Client, keyPairs ...[]byte) *RedisStore {
	rs := &RedisStore{
		Pool:       pool,
		Codecs:     securecookie.CodecsFromPairs(keyPairs...),
		Options:    defaultSessionOptions,
		maxLength:  4096,
		keyPrefix:  "session_", // session key prefix in redis
		serializer: serializer.GobSerializer{},
	}
	return rs
}

func (s *RedisStore) Close() error {
	return s.Pool.Close()
}

//SetExpireSecond 设置过期秒数
func (s *RedisStore) SetExpireSecond(expire int) *RedisStore {
	if expire >= 0 {
		s.Options.MaxAge = expire
	}
	return s
}

//SetMaxLength 设置session内容最大限制
func (s *RedisStore) SetMaxLength(l int) *RedisStore {
	if l >= 0 {
		s.maxLength = l
	}
	return s
}

// SetSerializer sets the serializer
func (s *RedisStore) SetSerializer(ss serializer.ISessionSerializer) *RedisStore {
	s.serializer = ss
	return s
}

// Get returns a session for the given name after adding it to the registry.
// See gorilla/sessions FilesystemStore.Get().
func (s *RedisStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

// New returns a session for the given name without adding it to the registry.
// see gorilla/sessions FilesystemStore.New().
func (s *RedisStore) New(r *http.Request, name string) (*sessions.Session, error) {
	var (
		err error
		ok  bool
	)
	session := sessions.NewSession(s, name)
	options := *s.Options
	session.Options = &options
	session.IsNew = true
	if c, errNoCookie := r.Cookie(name); errNoCookie == nil {
		// decode function see: github.com/gorilla/securecookie Decode(). 303 row
		err = securecookie.DecodeMulti(name, c.Value, &session.ID, s.Codecs...)
		if err == nil {
			ok, err = s.load(session)
			if err == nil && ok {
				session.IsNew = false
			}
		}
	}
	return session, err
}

//Save Save the sess information to redis, if the request does not exist, generate a new session
func (s *RedisStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	if session.Options.MaxAge <= 0 {
		if err := s.delete(session); err != nil {
			return err
		}
		http.SetCookie(w, sessions.NewCookie(session.Name(), "", session.Options))
		return nil
	}
	if session.ID == "" {
		session.ID = strings.TrimRight(base32.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32)), "=")
	}
	if err := s.save(session); err != nil {
		return err
	}
	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, s.Codecs...)
	if err != nil {
		return err
	}
	http.SetCookie(w, sessions.NewCookie(session.Name(), encoded, session.Options))
	return nil
}

//load read the session from redis.
func (s *RedisStore) load(session *sessions.Session) (bool, error) {
	data, err := s.Pool.Get(s.keyPrefix + session.ID).Result()
	if err != nil {
		return false, err
	}
	if data == "" {
		return false, nil
	}
	return true, s.serializer.DeSerialize([]byte(data), session)
}

//save stores the session in redis.
func (s *RedisStore) save(session *sessions.Session) error {
	bs, err := s.serializer.Serialize(session)
	if err != nil {
		return err
	}
	if s.maxLength != 0 && len(bs) > s.maxLength {
		return errors.New("the value to store is too big")
	}
	age := session.Options.MaxAge
	if age == 0 {
		age = s.Options.MaxAge
	}
	return s.Pool.Set(s.keyPrefix+session.ID, bs, time.Duration(age)*time.Second).Err()
}

// delete removes keys from redis
func (s *RedisStore) delete(session *sessions.Session) error {
	return s.Pool.Del(s.keyPrefix + session.ID).Err()
}
