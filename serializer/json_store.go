package serializer

import (
	"fmt"
	"github.com/golyu/redis-session/internal/json"
	"github.com/gorilla/sessions"
)

// JSONSerializer json类型的session
type JSONSerializer struct {
}

//Serialize 序列化session数据到json
func (J *JSONSerializer) Serialize(ss *sessions.Session) ([]byte, error) {
	m := make(map[string]interface{}, len(ss.Values))
	for k, v := range ss.Values {
		keyStr, ok := k.(string)
		if !ok {
			err := fmt.Errorf("Non-string key value, cannot serialize session to JSON: %s", k)
			fmt.Printf("redistore.JSONSerializer.serialize() Error: %v", err)
			return nil, err
		}
		m[keyStr] = v
	}
	return json.Marshal(m)
}

//DeSerialize 反序列化json数据到map[string]interface{}
func (J *JSONSerializer) DeSerialize(d []byte, ss *sessions.Session) error {
	m := make(map[string]interface{})
	err := json.Unmarshal(d, &m)
	if err != nil {
		fmt.Printf("redistore.JSONSerializer.deserialize() Error: %v", err)
		return err
	}
	for k, v := range m {
		ss.Values[k] = v
	}
	return nil
}
