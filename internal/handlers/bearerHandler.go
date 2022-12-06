package handlers

import (
	"context"
	"jrest/internal/security"
	"net/http"
)

func BearerHandler(claims security.Claims, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		authorized, tokenClaims := security.BearerAuthorized(r, claims)
		if !authorized {
			w.WriteHeader(http.StatusUnauthorized)
			path := ctx.Value(Path).(string)
			AuditLog(r.Method, path, "Not authorized")
			return
		}
		attr := ctx.Value(Attributes).(map[string]interface{})
		attr[AttrAuth] = true
		attr[AttrUser] = tokenClaims
		ctx = context.WithValue(ctx, Authorized, true)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
