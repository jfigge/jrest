package responses

import (
	"fmt"
	"jrest/internal/handlers"
	"jrest/internal/models"
	"net/http"
)

func getHandler(response *models.Response) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		path := ctx.Value(handlers.Path).(string)
		attr := ctx.Value(handlers.Attributes).(map[string]interface{})
		fmt.Printf("%v\n", attr)
		for key, value := range response.Headers {
			w.Header().Set(key, value)
		}
		if response.Status != 0 {
			w.WriteHeader(response.Status)
			handlers.AuditLog(r.Method, path, fmt.Sprintf("%d", response.Status))
		}
		_, _ = w.Write(append([]byte(response.Content), []byte("\n")...))
		if response.Status == 0 {
			handlers.AuditLog(r.Method, path, "200")
		}
	})
}
