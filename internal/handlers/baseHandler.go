package handlers

import (
	"context"
	"jrest/internal/models"
	"net/http"
	"strings"
)

func BaseHandler(source *models.Source) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, source.Base) {
			w.WriteHeader(http.StatusNotFound)
			AuditLog(r.Method, r.URL.Path, "Not found")
			return
		}
		ctx := r.Context()
		path := r.URL.Path[len(source.Base)+1:]
		attr := make(map[string]interface{})
		attr[AttrBase] = source.Base
		attr[AttrPath] = path
		ctx = context.WithValue(ctx, Attributes, attr)
		ctx = context.WithValue(ctx, Path, path)
		auth := source.Authentication
		r = r.WithContext(ctx)
		next := PathHandler(source.Paths)
		if auth != nil {
			AuthHandler(auth, next).ServeHTTP(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
