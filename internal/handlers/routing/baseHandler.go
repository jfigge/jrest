package routing

import (
	"context"
	"jrest/internal/handlers"
	auth2 "jrest/internal/handlers/authentication"
	"jrest/internal/models"
	"net/http"
	"strings"
)

func BaseHandler(source *models.Source) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, source.Base) {
			w.WriteHeader(http.StatusNotFound)
			handlers.AuditLog(r.Method, r.URL.Path, "Not found")
			return
		}
		ctx := r.Context()
		path := r.URL.Path[len(source.Base)+1:]
		attr := make(map[string]interface{})
		attr[handlers.AttrBase] = source.Base
		attr[handlers.AttrPath] = path
		ctx = context.WithValue(ctx, handlers.Attributes, attr)
		ctx = context.WithValue(ctx, handlers.Path, path)
		auth := source.Authentication
		r = r.WithContext(ctx)
		next := PathHandler(source.Paths)
		if auth != nil {
			auth2.AuthHandler(auth, next).ServeHTTP(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
