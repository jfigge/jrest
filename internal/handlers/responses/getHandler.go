package responses

import (
	"fmt"
	"jrest/internal/handlers"
	"jrest/internal/models"
	"net/http"
	"strings"
)

func getHandler(response *models.Response) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		path := ctx.Value(handlers.Path).(string)
		attr := ctx.Value(handlers.Attributes).(map[string]interface{})
		args := attr[handlers.AttrPathArgs].(map[string]string)
		for key, value := range response.Headers {
			w.Header().Set(key, value)
		}

		respData := ""
		if response.Content != nil {
			respData = *response.Content
		} else if response.Query != nil && ctx.Value(handlers.Store) != nil {
			db := ctx.Value(handlers.Store).(*models.Store)
			bs, err := db.Select(response.Query, args)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				handlers.AuditLog(r.Method, path, fmt.Sprintf("%d", response.Status))
				respData = err.Error()
				return
			}
			respData = string(bs)
		}

		if response.Status != 0 {
			w.WriteHeader(response.Status)
			handlers.AuditLog(r.Method, path, fmt.Sprintf("%d", response.Status))
		}

		for k, v := range args {
			respData = strings.ReplaceAll(respData, k, v)
		}
		_, _ = w.Write(append([]byte(respData), []byte("\n")...))
		if response.Status == 0 {
			handlers.AuditLog(r.Method, path, "200")
		}
	})
}
