package handlers

import "log"

const (
	Attributes = 1
	Authorized = 2
	Path       = 3
	Store      = 4
)

const (
	AttrBase     = "url.base"
	AttrPath     = "url.path"
	AttrMethod   = "url.method"
	AttrAuth     = "_.authorized"
	AttrUser     = "auth.user"
	AttrPathArgs = "url.args"
)

func AuditLog(method, path, status string) {
	log.Printf("Serving: %s:%s -> %s\n", method, path, status)
}
