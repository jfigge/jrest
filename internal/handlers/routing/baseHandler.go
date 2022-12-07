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
		path := r.URL.Path[len(source.Base)+1:]
		attr := make(map[string]interface{})
		attr[handlers.AttrBase] = source.Base
		attr[handlers.AttrPath] = path

		ctx := r.Context()
		ctx = context.WithValue(ctx, handlers.Attributes, attr)
		ctx = context.WithValue(ctx, handlers.Path, path)
		if source.DB != nil {
			ctx = context.WithValue(ctx, handlers.Store, source.DB)
		}

		auth := source.Authentication
		next := PathHandler(source.Paths)
		r = r.WithContext(ctx)
		auth2.AuthHandler(auth, next).ServeHTTP(w, r)
	})
}
