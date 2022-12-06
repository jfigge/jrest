package handlers

import (
	"context"
	"jrest/internal/security"
	"net/http"
)

func CredentialsHandler(claims security.Claims, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		authorized, credentialClaims := security.CredentialsAuthorized(r, claims)
		if !authorized {
			w.WriteHeader(http.StatusUnauthorized)
			path := ctx.Value(Path).(string)
			AuditLog(r.Method, path, "Not authorized")
			return
		}
		attr := ctx.Value(Attributes).(map[string]interface{})
		attr[AttrAuth] = true
		attr[AttrUser] = credentialClaims
		ctx = context.WithValue(ctx, Authorized, true)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
