package routing

import (
	"jrest/internal/handlers"
	auth2 "jrest/internal/handlers/authentication"
	"jrest/internal/handlers/responses"
	"jrest/internal/models"
	"net/http"
)

func MethodHandler(methods models.Methods) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		path := ctx.Value(handlers.Path).(string)
		response, ok := methods[r.Method]
		if !ok {
			w.WriteHeader(http.StatusMethodNotAllowed)
			handlers.AuditLog(r.Method, path, "Not found")
			return
		}
		attr := ctx.Value(handlers.Attributes).(map[string]interface{})
		attr[handlers.AttrMethod] = r.Method
		auth := response.Authentication
		next := responses.ResponseHandler(response)
		auth2.AuthHandler(auth, next).ServeHTTP(w, r)
	})
}
