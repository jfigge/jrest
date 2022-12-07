package models

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-memdb"
	"jrest/internal/security"
	"log"
	"reflect"
	"strconv"
	"strings"
)

type Paths map[string]*Path
type Methods map[string]*Response

type Response struct {
	Authentication *Authentication   `json:"auth"`
	Status         int               `json:"status_code"`
	Content        json.RawMessage   `json:"content"`
	Headers        map[string]string `json:"headers,omitempty"`
}

type Authentication struct {
	Bearer      security.Claims `json:"bearer,omitempty"`
	Credentials security.Claims `json:"credentials,omitempty"`
}

type Path struct {
	Authentication *Authentication `json:"auth"`
	Methods        Methods         `json:"methods,omitempty"`
}

type Source struct {
	Host           string          `json:"host" default:"127.0.0.1"`
	Base           string          `json:"base" default:"/"`
	Port           int             `json:"port" default:"8080"`
	Timeout        int             `json:"timeout" default:"30"`
	Authentication *Authentication `json:"auth"`
	Paths          Paths           `json:"paths,omitempty"`
	db             *memdb.MemDB
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
	for key, api := range s.Paths {
		if strings.HasPrefix(key, "/") {
			delete(s.Paths, key)
			s.Paths[key[1:]] = api
		}
	}
}

func (s *Source) Audit() {
	fmt.Println("Supported APIs:")
	for key, api := range s.Paths {
		for method := range api.Methods {
			fmt.Printf("  %s: %s\n", method, key)
		}
	}
}

func (s *Source) ConfigureMemDB() {

	// Create a new database
	var err error
	s.db, err = memdb.NewMemDB(schema)
	if err != nil {
		log.Fatalf("Unable to start database: %v", err)
		return
	}

	// Create a write transaction
	txn := s.db.Txn(true)
	defer txn.Abort()

	// Insert some people
	people := []*Person{
		&Person{"joe@aol.com", "Joe", 30},
		&Person{"lucy@aol.com", "Lucy", 35},
		&Person{"tariq@aol.com", "Tariq", 21},
		&Person{"dorothy@aol.com", "Dorothy", 53},
	}
	for _, p := range people {
		if err = txn.Insert("person", p); err != nil {
			panic(err)
		}
	}

	// Commit the transaction
	txn.Commit()

	// Create read-only transaction
	txn = s.db.Txn(false)
	defer txn.Abort()

	// Lookup by email
	raw, err := txn.First("person", "id", "joe@aol.com")
	if err != nil {
		panic(err)
	}

	// Say hi!
	fmt.Printf("Hello %s!\n", raw.(*Person).Name)

	// List all the people
	it, err := txn.Get("person", "id")
	if err != nil {
		panic(err)
	}

	fmt.Println("All the people:")
	for obj := it.Next(); obj != nil; obj = it.Next() {
		p := obj.(*Person)
		fmt.Printf("  %s\n", p.Name)
	}

	// Range scan over people with ages between 25 and 35 inclusive
	it, err = txn.LowerBound("person", "age", 25)
	if err != nil {
		panic(err)
	}

	fmt.Println("People aged 25 - 35:")
	for obj := it.Next(); obj != nil; obj = it.Next() {
		p := obj.(*Person)
		if p.Age > 35 {
			break
		}
		fmt.Printf("  %s is aged %d\n", p.Name, p.Age)
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
