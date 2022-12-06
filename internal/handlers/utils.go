package handlers

import "log"

const (
	Attributes = 1
	Authorized = 2
	Path       = 3
)

const (
	AttrBase   = "base"
	AttrPath   = "path"
	AttrAuth   = "authorized"
	AttrUser   = "user"
	AttrMethod = "method"
)

func AuditLog(method, path, status string) {
	log.Printf("Serving: %s:%s -> %s\n", method, path, status)
}
