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
		body, ok := paths.MatchPath(path)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			handlers.AuditLog(r.Method, path, "Not found")
			return
		}
		auth := body.Authentication
		next := MethodHandler(body.Methods)
		auth2.AuthHandler(auth, next).ServeHTTP(w, r)
	})
}
