package serializer

import "github.com/gorilla/sessions"

//ISessionSerializer 序列化操作接口
type ISessionSerializer interface {
	Serialize(ss *sessions.Session) ([]byte, error)
	DeSerialize(d []byte, ss *sessions.Session) error
}
