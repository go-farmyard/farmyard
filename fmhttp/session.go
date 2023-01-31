package fmhttp

import (
	"github.com/go-farmyard/farmyard/fmutil"
	"github.com/gorilla/sessions"
	"net/http"
)

type Session struct {
	session *sessions.Session
	changed bool
}

func NewSession(session *sessions.Session) *Session {
	return &Session{session: session}
}

func (ws *Session) Get(key string) any {
	return ws.session.Values[key]
}

func (ws *Session) GetString(key string) string {
	return fmutil.AsString(ws.session.Values[key])
}

func (ws *Session) GetInt64(key string) int64 {
	return fmutil.AsInt64(ws.session.Values[key])
}

func (ws *Session) Set(key string, val any) {
	ws.session.Values[key] = val
	ws.changed = true
}

func (ws *Session) Remove(key string) {
	delete(ws.session.Values, key)
	ws.changed = true
}

func (ws *Session) Destroy() {
	opts := *ws.session.Options
	opts.MaxAge = -1
	ws.session.Options = &opts
	ws.changed = true
}

func (ws *Session) SetChanged() {
	ws.changed = true
}

func (ws *Session) AutoSave(w http.ResponseWriter, r *http.Request) error {
	if ws.changed {
		ws.changed = false
		return ws.session.Save(r, w)
	}
	return nil
}
