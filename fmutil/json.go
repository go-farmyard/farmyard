package fmutil

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
)

type JsonDataProvider interface {
	JsonData() any
}

type AnyJsonDataProvider struct {
	v any
}

func (a *AnyJsonDataProvider) JsonData() any {
	return a.v
}

func AsJsonDataProvider(v any) JsonDataProvider {
	return &AnyJsonDataProvider{v}
}

type JsonValue struct {
	v any
}

func (jv *JsonValue) JsonData() any {
	return jv.AsInterface()
}

var jsonValueNull = &JsonValue{}

func AsInt64(n any, defs ...int64) int64 {
	if n == nil {
		return DefZero(defs)
	}
	switch n := n.(type) {
	case int:
		return int64(n)
	case int8:
		return int64(n)
	case int16:
		return int64(n)
	case int32:
		return int64(n)
	case int64:
		return n
	case uint:
		return int64(n)
	case uintptr:
		return int64(n)
	case uint8:
		return int64(n)
	case uint16:
		return int64(n)
	case uint32:
		return int64(n)
	case uint64:
		return int64(n)

	case float32:
		return int64(n)
	case float64:
		return int64(n)

	case string:
		v, _ := strconv.ParseInt(n, 10, 64)
		return v
	}
	return 0
}

func AsFloat64(n any, defs ...float64) float64 {
	if n == nil {
		return DefZero(defs)
	}
	switch n := n.(type) {
	case int:
		return float64(n)
	case int8:
		return float64(n)
	case int16:
		return float64(n)
	case int32:
		return float64(n)
	case int64:
		return float64(n)
	case uint:
		return float64(n)
	case uintptr:
		return float64(n)
	case uint8:
		return float64(n)
	case uint16:
		return float64(n)
	case uint32:
		return float64(n)
	case uint64:
		return float64(n)

	case float32:
		return float64(n)
	case float64:
		return n

	case string:
		v, _ := strconv.ParseFloat(n, 64)
		return v
	}
	return 0
}

func AsString(v any, defs ...string) string {
	if v == nil {
		return DefZero(defs)
	}
	switch v := v.(type) {
	case string:
		return v
	case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64:
		return strconv.FormatInt(AsInt64(v, 0), 10)
	case float32, float64:
		return strconv.FormatFloat(AsFloat64(v, 0), 'f', -1, 64)
	default:
		return fmt.Sprintf("%T", v)
	}
}

func AsBool(v any, defs ...bool) bool {
	if v == nil {
		return DefZero(defs)
	}
	switch v := v.(type) {
	case bool:
		return v
	case string:
		b, _ := strconv.ParseBool(v)
		return b
	case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64:
		return v != 0
	case float32, float64:
		return v != 0
	default:
		return v != nil
	}
}

func (jv *JsonValue) AsInterface() any {
	if jv == nil {
		return nil
	}
	return jv.v
}

func (jv *JsonValue) AsMap() Map {
	if jv == nil {
		return nil
	}
	if m, ok := jv.v.(map[string]any); ok {
		return m
	}
	return nil
}

func (jv *JsonValue) AsString(defs ...string) string {
	if jv == nil {
		return AsString(jsonValueNull, defs...)
	}
	return AsString(jv.v, defs...)
}

func (jv *JsonValue) AsInt64(defs ...int64) int64 {
	if jv == nil {
		return AsInt64(jsonValueNull, defs...)
	}
	return AsInt64(jv.v, defs...)
}

func (jv *JsonValue) AsFloat64(defs ...float64) float64 {
	if jv == nil {
		return AsFloat64(jsonValueNull, defs...)
	}
	return AsFloat64(jv.v, defs...)
}

func (jv *JsonValue) GetString(key string, defs ...string) string {
	if jv != nil {
		if m, ok := jv.v.(map[string]any); ok {
			if v, ok := m[key]; ok {
				return AsString(v, defs...)
			}
		}
	}
	if len(defs) > 0 {
		return defs[0]
	}
	return ""
}

func (jv *JsonValue) GetInt64(key string, defs ...int64) int64 {
	if jv != nil {
		if m, ok := jv.v.(map[string]any); ok {
			if v, ok := m[key]; ok {
				return AsInt64(v, defs...)
			}
		}
	}
	if len(defs) > 0 {
		return defs[0]
	}
	return 0
}

func (jv *JsonValue) Get(key string) *JsonValue {
	if jv != nil {
		if m, ok := jv.v.(map[string]any); ok {
			return &JsonValue{v: m[key]}
		}
	}
	return &JsonValue{v: nil}
}

func JsonDecodeBytesWithError(buf []byte) (*JsonValue, error) {
	m := map[string]any{}
	err := json.Unmarshal(buf, &m)
	if err != nil {
		return nil, err
	}
	return &JsonValue{v: m}, nil
}

func JsonDecodeReader(r io.Reader) (*JsonValue, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return JsonDecodeBytesWithError(buf)
}

func JsonDecodeString(s string) *JsonValue {
	jv, _ := JsonDecodeBytesWithError([]byte(s))
	return jv
}

func JsonEncode(v any) []byte {
	buf, _ := json.Marshal(v)
	return buf
}

func JsonEncodeString(v any) string {
	return string(JsonEncode(v))
}
