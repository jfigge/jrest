package models

import (
	"fmt"
	"jrest/internal/security"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/hashicorp/go-memdb"
)

type Paths map[string]*Path
type Methods map[string]*Response
type Schema map[string]interface{}

type Response struct {
	Authentication *Authentication   `json:"auth,omitempty" yaml:"auth,omitempty"`
	Status         int               `json:"status_code,omitempty" yaml:"status_code,omitempty"`
	Content        string            `json:"content" yaml:"content"`
	Headers        map[string]string `json:"headers,omitempty" yaml:"headers"`
}

type Authentication struct {
	Bearer      security.Claims `json:"bearer,omitempty" yaml:"bearer"`
	Credentials security.Claims `json:"credentials,omitempty" yaml:"credentials"`
}

type Path struct {
	Authentication *Authentication `json:"auth,omitempty" yaml:"auth,omitempty"`
	Methods        Methods         `json:"methods" yaml:"methods"`
}

type Source struct {
	Host           string          `json:"host" yaml:"host" default:"127.0.0.1"`
	Base           string          `json:"base" yaml:"base" default:"/"`
	Port           int             `json:"port" yaml:"port" default:"8080"`
	Timeout        int             `json:"timeout" yaml:"timeout" default:"30"`
	Authentication *Authentication `json:"auth,omitempty" yaml:"auth,omitempty"`
	Paths          Paths           `json:"paths" yaml:"paths"`
	Storage        *Store          `json:"storage,omitempty" yaml:"storage,omitempty"`
	DB             *memdb.MemDB
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
	for path, body := range s.Paths {
		for method := range body.Methods {
			log.Printf("  %-6s %s\n", fmt.Sprintf("%s:", method), path)
		}
	}
}

func (s *Source) ConfigureMemDB() {

	if s.Storage == nil || len(s.Storage.Entities) == 0 {
		s.DB = nil
		return
	}

	// Create a new database
	var err error
	s.DB, err = memdb.NewMemDB(s.Storage.buildSchema())
	if err != nil {
		log.Fatalf("Unable to start database: %v", err)
		return
	}

	// Load test data
	if s.Storage.Data != nil {
		// Create a writeable transaction
		txn := s.DB.Txn(true)
		defer txn.Abort()

		for entityName, rows := range s.Storage.Data {
			for _, row := range rows {
				entity, ok := s.Storage.Entities[entityName]
				if !ok {
					log.Fatalf("unknown data entity: %s", entityName)
					return
				}
				table := entity.Table
				instance := table.getInstance()
				entry := table.setValues(instance, row).Interface()
				if err = txn.Insert(entityName, entry); err != nil {
					panic(err)
				}
			}
		}

		// Commit the transaction
		txn.Commit()

		// Create read-only transaction
		txn = s.DB.Txn(false)
		defer txn.Abort()

		// Lookup by email
		var raw interface{}
		raw, err = txn.First("person", "id", "jason.figge@gmail.com")
		if err != nil {
			panic(err)
		}

		// Say hi!
		table := s.Storage.Entities["person"].Table
		fmt.Printf("Hello %v!\n", table.toMap(raw)["Name"])

		// List all the people
		var it memdb.ResultIterator
		it, err = txn.Get("person", "id")
		if err != nil {
			panic(err)
		}

		fmt.Println("All the people:")
		for obj := it.Next(); obj != nil; obj = it.Next() {
			p := table.toMap(obj)
			fmt.Printf("  %-8s%v\n", fmt.Sprintf("%s:", p["Name"]), p["Age"])
		}

		// Range scan over people with ages between 25 and 35 inclusive
		it, err = txn.LowerBound("person", "age", 49)
		if err != nil {
			panic(err)
		}
		fmt.Println("All the people over 50:")
		for obj := it.Next(); obj != nil; obj = it.Next() {
			p := table.toMap(obj)
			fmt.Printf("  %-8s%v\n", fmt.Sprintf("%s:", p["Name"]), p["Age"])
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
