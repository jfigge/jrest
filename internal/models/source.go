package models

import (
	"encoding/json"
	"fmt"
	"jrest/internal/security"
	"reflect"
	"strconv"
)

type Methods map[string]response

type response struct {
	Bearer      security.Claims   `json:"bearer,omitempty"`
	Credentials security.Claims   `json:"credentials,omitempty"`
	Status      int               `json:"status_code"`
	Data        json.RawMessage   `json:"data"`
	Headers     map[string]string `json:"headers,omitempty"`
}

type Source struct {
	Host    string             `json:"host" default:"127.0.0.1"`
	Base    string             `json:"base" default:"/"`
	Port    int                `json:"port" default:"8080"`
	Timeout int                `json:"timeout" default:"30"`
	Api     map[string]Methods `json:"api,omitempty"`
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

func atoi(val string) int {
	i, err := strconv.Atoi(val)
	if err != nil {
		fmt.Printf("Debug: Invalid default: %s: %v\n", val, err)
		return 0
	}
	return i
}
