package responses

import (
	"jrest/internal/handlers"
	"jrest/internal/models"
	"net/http"
)

func ResponseHandler(response *models.Response) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		path := ctx.Value(handlers.Path).(string)
		switch r.Method {
		case http.MethodGet:
			getHandler(response).ServeHTTP(w, r)
		default:
			w.WriteHeader(http.StatusGone)
			handlers.AuditLog(r.Method, path, "Not implemented")
			return
		}
	})
}
