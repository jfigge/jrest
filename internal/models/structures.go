package models

import (
	"github.com/hashicorp/go-memdb"
	"jrest/internal/security"
)

type Paths map[string]*Path
type Methods map[string]*Response
type Schema map[string]interface{}
type Entities map[string]Entity
type Entity map[string]DataType

type Store struct {
	Entities Entities `json:"entities" yaml:"entities"`
	Schema   Schema   `json:"schema" yaml:"schema"`
}
type Response struct {
	Authentication *Authentication   `json:"auth" yaml:"auth"`
	Status         int               `json:"status_code" yaml:"status_code"`
	Content        string            `json:"content" yaml:"content"`
	Headers        map[string]string `json:"headers,omitempty" yaml:"headers"`
}

type Authentication struct {
	Bearer      security.Claims `json:"bearer,omitempty" yaml:"bearer"`
	Credentials security.Claims `json:"credentials,omitempty" yaml:"credentials"`
}

type Path struct {
	Authentication *Authentication `json:"auth" yaml:"auth"`
	Methods        Methods         `json:"methods" yaml:"methods"`
}

type Source struct {
	Host           string          `json:"host" yaml:"host" default:"127.0.0.1"`
	Base           string          `json:"base" yaml:"base" default:"/"`
	Port           int             `json:"port" yaml:"port" default:"8080"`
	Timeout        int             `json:"timeout" yaml:"timeout" default:"30"`
	Authentication *Authentication `json:"auth" yaml:"auth"`
	Paths          Paths           `json:"paths" yaml:"paths"`
	Storage        Store           `json:"storage" yaml:"storage"`
	DB             *memdb.MemDB
}
