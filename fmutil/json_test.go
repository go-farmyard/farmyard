package fmutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMap(t *testing.T) {
	jv := JsonDecodeString(`{"a":{"b":123}}`)
	assert.EqualValues(t, 123, jv.Get("a").AsMap()["b"])
	assert.EqualValues(t, "123", jv.Get("a").Get("b").AsString())
}

func TestJsonValue(t *testing.T) {
	s := `{"i":1,"s":"test","m":{}}`
	jv := JsonDecodeString(s)
	assert.EqualValues(t, 1, jv.GetInt64("i"))
	assert.EqualValues(t, "test", jv.GetString("s"))
	assert.EqualValues(t, map[string]any{}, jv.Get("m").AsInterface())
}
