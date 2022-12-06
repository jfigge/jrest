package handlers

import (
	"jrest/internal/models"
	"net/http"
)

func MethodHandler(methods models.Methods) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		path := ctx.Value(Path).(string)
		response, ok := methods[r.Method]
		if !ok {
			w.WriteHeader(http.StatusMethodNotAllowed)
			AuditLog(r.Method, path, "Not found")
			return
		}
		attr := ctx.Value(Attributes).(map[string]interface{})
		attr[AttrMethod] = r.Method
		auth := response.Authentication
		next := ResponseHandler(response)
		if auth != nil {
			AuthHandler(auth, next).ServeHTTP(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
