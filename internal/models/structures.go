package models

import (
	"encoding/json"
	"fmt"
	"jrest/internal/security"
	"log"
	"reflect"
	"strconv"
	"strings"
)

type Paths map[string]*Path
type Methods map[string]*Response

type Response struct {
	Authentication *Authentication   `json:"auth" yaml:"auth"`
	Status         int               `json:"status_code" yaml:"status_code"`
	Content        json.RawMessage   `json:"content" yaml:"content"`
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
}

func (s *Source) ApplyDefaults() {
	v := reflect.ValueOf(s)
	t := reflect.TypeOf(*s)
	for i := 0; i < t.NumField(); i++ {
		fv := v.Elem().Field(i)
		if fv.CanSet() {
			if val, ok := t.Field(i).Tag.Lookup("default"); ok && val != "" {
				switch x := fv.Interface().(type) {
				case string:
					if x == "" {
						fv.Set(reflect.ValueOf(val))
					}
				case int:
					if x == 0 {
						fv.Set(reflect.ValueOf(atoi(val)))
					}
				default:
				}
			}
		}
	}
}

func (s *Source) Cleanse() {
	// wrap base is slashes
	if !strings.HasPrefix(s.Base, "/") {
		s.Base = fmt.Sprintf("/%s", s.Base)
	}
	if strings.HasSuffix(s.Base, "/") {
		s.Base = s.Base[:len(s.Base)-1]
	}

	// strip slashes from base of paths
	for path, body := range s.Paths {
		if strings.HasPrefix(path, "/") {
			delete(s.Paths, path)
			s.Paths[path[1:]] = body
		}
	}
}

func (s *Source) LogPaths() {
	fmt.Println("Supported paths:")
	for path, body := range s.Paths {
		for method := range body.Methods {
			log.Printf("  %s: %s\n", method, path)
		}
	}
}

func atoi(val string) int {
	i, err := strconv.Atoi(val)
	if err != nil {
		fmt.Printf("Debug: Invalid default: %s: %v\n", val, err)
		return 0
	}
	return i
}
