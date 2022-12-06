package routing

import (
	"jrest/internal/handlers"
	auth2 "jrest/internal/handlers/authentication"
	"jrest/internal/models"
	"net/http"
)

func PathHandler(paths models.Paths) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		path := ctx.Value(handlers.Path).(string)
		api, ok := paths[path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			handlers.AuditLog(r.Method, path, "Not found")
			return
		}
		auth := api.Authentication
		next := MethodHandler(api.Methods)
		r = r.WithContext(ctx)
		if auth != nil {
			auth2.AuthHandler(auth, next).ServeHTTP(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
