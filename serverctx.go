package tpl

import (
	"context"
	"net/http"
	"sync"

	"github.com/KarpelesLab/webutil"
)

type serverCtxOp int

const serverCtxReset serverCtxOp = iota

type serverCtx struct {
	context.Context
	request *http.Request

	getV     map[string]interface{}
	postV    map[string]interface{}
	requestV map[string]interface{}

	get, post sync.Once
}

func ServerCtx(request *http.Request) context.Context {
	return &serverCtx{Context: request.Context(), request: request}
}

func ResetServerCtx(ctx context.Context) {
	ctx.Value(serverCtxReset)
}

func (s *serverCtx) Value(key interface{}) interface{} {
	if op, ok := key.(serverCtxOp); ok {
		switch op {
		case serverCtxReset:
			s.get = sync.Once{}
			s.post = sync.Once{}
		}
	}

	keyStr, ok := key.(string)
	if !ok {
		return s.Context.Value(key)
	}
	switch keyStr {
	case "$_get":
		s.get.Do(func() {
			s.getV = webutil.ParsePhpQuery(s.request.URL.RawQuery)
		})
		return s.getV
	case "$_post":
		s.post.Do(func() {
			s.request.ParseMultipartForm(2 * 1024 * 1024)
			s.postV = webutil.ConvertPhpQuery(s.request.PostForm)
			s.requestV = webutil.ConvertPhpQuery(s.request.Form)
		})
		return s.postV
	case "$_request":
		s.post.Do(func() {
			s.request.ParseMultipartForm(2 * 1024 * 1024)
			s.postV = webutil.ConvertPhpQuery(s.request.PostForm)
			s.requestV = webutil.ConvertPhpQuery(s.request.Form)
		})
		return s.requestV
	}
	return s.Context.Value(key)
}
