package redis_session

import (
	"encoding/base64"
	"encoding/gob"
	"github.com/golyu/redis-session/data"
	"github.com/golyu/redis-session/serializer"
	"github.com/gorilla/sessions"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"testing"
)

type Message struct {
	MInt    int
	MString string
	MFloat  float64
}

func init() {
	gob.Register(&Message{})
	_ = data.InitRedis()
}

func genStore() (*RedisStore, *http.Request) {
	var (
		req   *http.Request
		store *RedisStore
		err   error
	)
	Convey("test new store", func() {
		store = NewRedisStore(data.RedisConn, []byte("hash-key")).SetExpireSecond(data.SessionExpireTime)
		So(err, ShouldBeNil)
		req, err = http.NewRequest("GET", "http://localhost:8080/", nil)
		So(err, ShouldBeNil)
	})
	return store, req
}

func TestRedisStore(t *testing.T) {
	var (
		store   *RedisStore
		req     *http.Request
		resp    *httptest.ResponseRecorder
		ok      bool
		cookies []string
		session *sessions.Session
		flashes []interface{}
		err     error
	)
	Convey("test session add", t, func() {
		store, req = genStore()
		resp = httptest.NewRecorder()
		session, err := store.Get(req, "session-key")
		So(err, ShouldBeNil)
		flashes = session.Flashes()
		So(len(flashes), ShouldEqual, 0)
		session.AddFlash("name")
		session.AddFlash("sex", "custom_key")
		err = sessions.Save(req, resp)
		So(err, ShouldBeNil)
		cookies, ok = resp.Header()["Set-Cookie"]
		So(ok, ShouldBeTrue)
		So(len(cookies), ShouldEqual, 1)
	})
	Convey("test get flashes,_flash or custom key", t, func() {
		req.Header.Add("Cookie", cookies[0])
		session, err = store.Get(req, "session-key")
		So(err, ShouldBeNil)
		flashes = session.Flashes()
		So(len(flashes), ShouldEqual, 1)
		So(flashes[0], ShouldEqual, "name")
		flashes = session.Flashes()
		So(len(flashes), ShouldEqual, 0)
		flashes = session.Flashes("custom_key")
		So(len(flashes), ShouldEqual, 1)
		So(flashes[0], ShouldEqual, "sex")
		flashes = session.Flashes("custom_key")
		So(len(flashes), ShouldEqual, 0)
		// delete session by Set MaxAge -1
		session.Options.MaxAge = -1
		err = sessions.Save(req, resp)
		So(err, ShouldBeNil)
		session, err = store.Get(req, "session-key")
		So(err, ShouldBeNil)
		flashes = session.Flashes()
		So(len(flashes), ShouldEqual, 0)
		flashes = session.Flashes("custom_key")
		So(len(flashes), ShouldEqual, 0)
	})
	Convey("test save struct", t, func() {
		resp = httptest.NewRecorder()
		session, err = store.Get(req, "session-key")
		So(err, ShouldBeNil)
		flashes = session.Flashes()
		So(len(flashes), ShouldEqual, 0)
		session.AddFlash(&Message{66, "str", 66.66666})
		err = sessions.Save(req, resp)
		So(err, ShouldBeNil)
		cookies, ok = resp.Header()["Set-Cookie"]
		So(ok, ShouldBeTrue)
		So(len(cookies), ShouldEqual, 1)
	})
	Convey("test get struct,", t, func() {
		req.Header.Add("Cookie", cookies[0])
		resp = httptest.NewRecorder()
		session, err = store.Get(req, "session-key")
		So(err, ShouldBeNil)
		flashes = session.Flashes()
		custom, ok := flashes[0].(*Message)
		So(ok, ShouldBeTrue)
		So(custom.MInt, ShouldEqual, 66)
		So(custom.MString, ShouldEqual, "str")
		So(custom.MFloat, ShouldEqual, 66.66666)
		// Set MaxAge to -1 to mark for deletion.
		session.Options.MaxAge = -1
		err = sessions.Save(req, resp)
		So(err, ShouldBeNil)
	})
}

func TestRedisStore_SetMaxLength(t *testing.T) {
	var (
		req     *http.Request
		store   *RedisStore
		session *sessions.Session
		err     error
	)
	Convey("TestRedisStore_SetMaxLength", t, func() {
		store, req = genStore()
		resp := httptest.NewRecorder()
		session, err = store.New(req, "session-key")
		So(err, ShouldBeNil)
		session.Values["data"] = make([]byte, base64.StdEncoding.DecodedLen(4096*2))
		err = session.Save(req, resp)
		So(err, ShouldNotBeNil)
		cookies, ok := resp.Header()["Set-Cookie"]
		So(ok, ShouldBeFalse)
		So(len(cookies), ShouldEqual, 0)
		store.SetMaxLength(4096 * 2)
		err = session.Save(req, resp)
		So(err, ShouldBeNil)
		cookies, ok = resp.Header()["Set-Cookie"]
		So(ok, ShouldBeTrue)
		So(len(cookies), ShouldEqual, 1)
	})
}

func TestRedisStore_SetSerializer(t *testing.T) {
	var (
		req     *http.Request
		store   *RedisStore
		session *sessions.Session
		err     error
	)
	Convey("TestRedisStore_SetSerializer", t, func() {
		store, req = genStore()
		store.SetSerializer(&serializer.JSONSerializer{})
		session, err = store.Get(req, "session-key")
		So(err, ShouldBeNil)
		flashes := session.Flashes()
		So(len(flashes), ShouldEqual, 0)
		session.AddFlash("name")
		resp := httptest.NewRecorder()
		err = sessions.Save(req, resp)
		So(err, ShouldBeNil)
		header := resp.Header()
		cookies, ok := header["Set-Cookie"]
		So(ok, ShouldBeTrue)
		So(len(cookies), ShouldEqual, 1)
		req.Header.Add("Cookie", cookies[0])
		session, err = store.Get(req, "session-key")
		So(err, ShouldBeNil)
		flashes = session.Flashes()
		So(len(flashes), ShouldEqual, 1)
		So(flashes[0], ShouldEqual, "name")
	})
}
