package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golyu/redis-session/internal/json"
	"github.com/gorilla/sessions"
)

const (
	SessionKey = "session-key"
	ValueKey   = "pan-key"
)

type SessionInfo struct {
	Account string
	Name    string
	Age     int
}

type SessionInfoJsonSerialize struct {
}

func (s *SessionInfoJsonSerialize) Serialize(ss *sessions.Session) ([]byte, error) {
	value, ok := ss.Values[ValueKey]
	if !ok {
		return nil, errors.New("No key registered by this platform")
	}
	return json.Marshal(value)
}

func (s *SessionInfoJsonSerialize) DeSerialize(d []byte, ss *sessions.Session) error {
	var info SessionInfo
	err := json.Unmarshal(d, &info)
	if err != nil {
		return err
	}
	ss.Values[ValueKey] = &info
	return nil
}

func SaveSession(info *SessionInfo, ctx *gin.Context) error {
	session, _ := store.New(ctx.Request, SessionKey)
	session.Values[ValueKey] = info
	return session.Save(ctx.Request, ctx.Writer)
}

func GetSession(ctx *gin.Context) *SessionInfo {
	value, ok := ctx.Get(ValueKey)
	if !ok {
		panic("need session info,please register in router group 'CheckLoginMiddleware'")
	}
	return value.(*SessionInfo)
}
