package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/yaml.v2"
)

type ConfigMap struct {
	data     map[string]interface{}
	dataType string
}

func (s *ConfigMap) SetConfigType(dataType string) {
	s.dataType = dataType
}

func (s *ConfigMap) parseBuffer(buff []byte, v interface{}) {
	var (
		err error
	)
	switch s.dataType {
	case "yaml":
		err = yaml.Unmarshal(buff, v)
	case "json":
		err = json.Unmarshal(buff, v)
	default:
		panic("Unknown config type")
	}
	if err != nil {
		panic(err)
	}
}

func (s *ConfigMap) SetConfigBuffer(buff string) {
	s.data = make(map[string]interface{})
	s.parseBuffer([]byte(buff), &s.data)
}

func (s *ConfigMap) MapResult(buff string, v interface{}) {
	s.parseBuffer([]byte(buff), v)
}

func (s *ConfigMap) GetByPathName(path []string) interface{} {
	if s.data == nil {
		return nil
	}

	var v interface{}
	var m map[interface{}]interface{}

	for index, key := range path {
		key = strings.TrimSpace(key)
		if len(key) <= 0 {
			continue
		}

		if m == nil {
			v = s.data[key]
		} else {
			v = m[key]
		}
		if index < (len(path) - 1) {
			if v != nil &&
				reflect.ValueOf(v).Type() == reflect.TypeOf(v) {
				m = v.(map[interface{}]interface{})
			}
		}
	}

	return v
}

func (s *ConfigMap) GetInt(keys string) int {
	path := strings.Split(keys, ".")
	v := s.GetByPathName(path)
	if v == nil || reflect.TypeOf(v).Kind() != reflect.Int {
		return 0
	}
	return v.(int)
}

func (s *ConfigMap) GetInt8(keys string) int8 {
	path := strings.Split(keys, ".")
	v := s.GetByPathName(path)
	if v == nil || reflect.TypeOf(v).Kind() != reflect.Int8 {
		return 0
	}
	return v.(int8)
}

func (s *ConfigMap) GetInt32(keys string) int32 {
	path := strings.Split(keys, ".")
	v := s.GetByPathName(path)
	if v == nil || reflect.TypeOf(v).Kind() != reflect.Int32 {
		return 0
	}
	return v.(int32)
}

func (s *ConfigMap) GetInt64(keys string) int64 {
	path := strings.Split(keys, ".")
	v := s.GetByPathName(path)
	if v == nil || reflect.TypeOf(v).Kind() != reflect.Int64 {
		return 0
	}
	return v.(int64)
}

func (s *ConfigMap) GetString(keys string) string {
	path := strings.Split(keys, ".")
	v := s.GetByPathName(path)
	if v == nil || reflect.TypeOf(v).Kind() != reflect.String {
		return ""
	}
	return v.(string)
}

func (s *ConfigMap) GetBool(keys string) bool {
	path := strings.Split(keys, ".")
	v := s.GetByPathName(path)
	if v == nil || reflect.TypeOf(v).Kind() != reflect.Bool {
		return false
	}
	return v.(bool)
}

func (s *ConfigMap) GetStringArray(keys string) []string {
	path := strings.Split(keys, ".")
	v := s.GetByPathName(path)
	if v == nil ||
		(reflect.TypeOf(v).Kind() != reflect.Slice && reflect.TypeOf(v).Kind() != reflect.Array) {
		return []string{}
	}
	var r []string
	for _, val := range v.([]interface{}) {
		switch reflect.TypeOf(val).Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			r = append(r, fmt.Sprintf("%d", val))
		case reflect.Float32, reflect.Float64:
			r = append(r, fmt.Sprintf("%f", val))
		case reflect.Bool:
			if val.(bool) {
				r = append(r, fmt.Sprintf("1"))
			} else {
				r = append(r, fmt.Sprintf("0"))
			}
		case reflect.String:
			r = append(r, val.(string))
		default:
			panic(fmt.Sprintf("unsupported type:[%d]", reflect.TypeOf(val)))
		}
	}
	return r
}

func TestGet() {
	v := `name: hello
a:
    b:
        c:
        - cc
        - false
        - 1
        - 1.11
        d: dd`

	c := ConfigMap{}
	c.SetConfigType("yaml")
	c.SetConfigBuffer(v)

	t := c.GetString("a.b.d")
	fmt.Printf("%v", t)
}
