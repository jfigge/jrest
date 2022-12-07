package authentication

import (
	"context"
	"jrest/internal/handlers"
	"jrest/internal/models"
	"net/http"
)

func AuthHandler(auth *models.Authentication, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if auth != nil && ctx.Value(handlers.Authorized) == nil {
			if auth.Bearer != nil {
				BearerHandler(auth.Bearer, next).ServeHTTP(w, r.WithContext(ctx))
				return
			} else if auth.Credentials != nil {
				CredentialsHandler(auth.Credentials, next).ServeHTTP(w, r.WithContext(ctx))
				return
			} else {
				attr := ctx.Value(handlers.Attributes).(map[string]interface{})
				attr[handlers.AttrAuth] = true
				r = r.WithContext(context.WithValue(ctx, handlers.Authorized, true))
			}
		}
		next.ServeHTTP(w, r)
	})
}
