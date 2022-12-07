package authentication

import (
	"context"
	"jrest/internal/handlers"
	"jrest/internal/security"
	"net/http"
)

func CredentialsHandler(claims security.Claims, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		authorized, credentialClaims := security.CredentialsAuthorized(r, claims)
		if !authorized {
			w.WriteHeader(http.StatusUnauthorized)
			path := ctx.Value(handlers.Path).(string)
			handlers.AuditLog(r.Method, path, "Not authorized")
			return
		}
		attr := ctx.Value(handlers.Attributes).(map[string]interface{})
		attr[handlers.AttrAuth] = true
		attr[handlers.AttrUser] = credentialClaims
		ctx = context.WithValue(ctx, handlers.Authorized, true)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
