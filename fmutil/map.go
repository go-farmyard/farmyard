package fmutil

type Map map[string]any

func (m Map) JsonData() any {
	return m
}

func (m Map) GetString(key string, defs ...string) string {
	if s, ok := m[key]; ok {
		return AsString(s, defs...)
	}
	return DefZero(defs)
}

func (m Map) GetBool(key string, defs ...bool) bool {
	if s, ok := m[key]; ok {
		return AsBool(s, defs...)
	}
	return DefZero(defs)
}

func (m Map) GetInt64(key string, defs ...int64) int64 {
	if s, ok := m[key]; ok {
		return AsInt64(s, defs...)
	}
	return DefZero(defs)
}
