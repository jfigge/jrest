package handlers

import (
	"context"
	"jrest/internal/models"
	"net/http"
)

func AuthHandler(auth *models.Authentication, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if ctx.Value(Authorized) == nil {
			if auth.Bearer != nil {
				BearerHandler(auth.Bearer, next).ServeHTTP(w, r.WithContext(ctx))
				return
			} else if auth.Credentials != nil {
				CredentialsHandler(auth.Credentials, next).ServeHTTP(w, r.WithContext(ctx))
				return
			} else {
				attr := ctx.Value(Attributes).(map[string]interface{})
				attr[AttrAuth] = true
				ctx = context.WithValue(ctx, Authorized, true)
			}
		}
		next.ServeHTTP(w, r)
	})
}
