package handlers

import "log"

const (
	Attributes = 1
	Authorized = 2
	Path       = 3
	Store      = 4
)

const (
	AttrBase     = "base"
	AttrPath     = "path"
	AttrAuth     = "authorized"
	AttrUser     = "user"
	AttrMethod   = "method"
	AttrPathArgs = "pathArgs"
)

func AuditLog(method, path, status string) {
	log.Printf("Serving: %s:%s -> %s\n", method, path, status)
}
