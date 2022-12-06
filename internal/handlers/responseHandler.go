package handlers

import (
	"fmt"
	"jrest/internal/models"
	"net/http"
)

func ResponseHandler(response *models.Response) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		path := ctx.Value(Path).(string)
		attr := ctx.Value(Attributes).(map[string]interface{})
		fmt.Printf("%v\n", attr)
		for key, value := range response.Headers {
			w.Header().Set(key, value)
		}
		if response.Status != 0 {
			w.WriteHeader(response.Status)
			AuditLog(r.Method, path, fmt.Sprintf("%d", response.Status))
		}
		_, _ = w.Write(append(response.Content, []byte("\n")...))
		if response.Status == 0 {
			AuditLog(r.Method, path, "200")
		}
	})
}
