package models

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-memdb"
	"gopkg.in/yaml.v3"
	"log"
	"reflect"
	"strconv"
	"strings"
)

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

func (s *Source) ConfigureMemDB() {

	// Create a new database
	var err error
	s.DB, err = memdb.NewMemDB(schema)
	if err != nil {
		log.Fatalf("Unable to start database: %v", err)
		return
	}

	// Create a writeable transaction
	txn := s.DB.Txn(true)
	defer txn.Abort()

	// Insert some people
	people := []*Person{
		{"joe@aol.com", "Joe", 30},
		{"lucy@aol.com", "Lucy", 35},
		{"tariq@aol.com", "Tariq", 21},
		{"dorothy@aol.com", "Dorothy", 53},
	}
	for _, p := range people {
		if err = txn.Insert("person", p); err != nil {
			panic(err)
		}
	}

	// Commit the transaction
	txn.Commit()

	// Create read-only transaction
	txn = s.DB.Txn(false)
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

func (e *Entity) UnmarshalYAML(value *yaml.Node) error {
	*e = make(Entity)
	for index := 0; index < len(value.Content); index += 2 {
		name := value.Content[index].Value
		d := DataTypeOf(value.Content[index+1].Value)
		(*e)[name] = d
	}
	return nil
}

func (e *Entity) UnmarshalJSON(data []byte) error {
	tmp := make(map[string]string)
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	*e = make(Entity)
	for name, value := range tmp {
		d := DataTypeOf(value)
		(*e)[name] = d
	}
	return nil
}
