package handlers

import (
	"jrest/internal/models"
	"net/http"
)

func PathHandler(paths models.Paths) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		path := ctx.Value(Path).(string)
		api, ok := paths[path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			AuditLog(r.Method, path, "Not found")
			return
		}
		auth := api.Authentication
		next := MethodHandler(api.Methods)
		r = r.WithContext(ctx)
		if auth != nil {
			AuthHandler(auth, next).ServeHTTP(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
