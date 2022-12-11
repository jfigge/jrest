package models

import (
	"context"
	"encoding/json"
	"fmt"
	"jrest/internal/handlers"
	"jrest/internal/security"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/hashicorp/go-memdb"
	"gopkg.in/yaml.v3"
)

type Source struct {
	Host           string          `json:"host" yaml:"host" default:"127.0.0.1"`
	Base           string          `json:"base" yaml:"base" default:"/"`
	Port           int             `json:"port" yaml:"port" default:"8080"`
	Timeout        int             `json:"timeout" yaml:"timeout" default:"30"`
	TLS            *Tls            `json:"tls,omitempty" yaml:"tls,omitempty"`
	Authentication *Authentication `json:"auth,omitempty" yaml:"auth,omitempty"`
	Paths          Paths           `json:"paths" yaml:"paths"`
	Storage        *Store          `json:"storage,omitempty" yaml:"storage,omitempty"`
}
type Tls struct {
	CertFile string `json:"certFile" yaml:"certFile"`
	KeyFile  string `json:"keyFile" yaml:"keyFile"`
}
type Authentication struct {
	Bearer      security.Claims `json:"bearer,omitempty" yaml:"bearer"`
	Credentials security.Claims `json:"credentials,omitempty" yaml:"credentials"`
}
type Paths struct {
	audit   []string
	dynamic []*PathMeta
	static  map[string]*Path
}
type PathMeta struct {
	parts     []string
	arguments map[int]string
	path      *Path
}
type Path struct {
	Authentication *Authentication `json:"auth,omitempty" yaml:"auth,omitempty"`
	Methods        Methods         `json:"methods" yaml:"methods"`
}
type Methods map[string]*Response
type Response struct {
	Authentication *Authentication   `json:"auth,omitempty" yaml:"auth,omitempty"`
	Status         int               `json:"status_code,omitempty" yaml:"status_code,omitempty"`
	Content        *string           `json:"content" yaml:"content"`
	Headers        map[string]string `json:"headers,omitempty" yaml:"headers"`
	Select         *Query            `json:"select" yaml:"select"`
	Insert         *Query            `json:"insert" yaml:"insert"`
	Delete         *Query            `json:"delete" yaml:"delete"`
}
type Query struct {
	Action   string  `json:"action" yaml:"action"`
	Entity   string  `json:"entity" yaml:"entity"`
	Filter   *Filter `json:"filter,omitempty" yaml:"filter,omitempty"`
	Page     *int    `json:"page,omitempty" yaml:"page,omitempty"`
	PageSize *int    `json:"page_size,omitempty" yaml:"page_size,omitempty"`
}
type Filter struct {
	Index  *string  `json:"index,omitempty" yaml:"index,omitempty"`
	Fields []string `json:"fields,omitempty" yaml:"fields,omitempty"`
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
}

func (s *Source) LogPaths() {
	for _, api := range s.Paths.audit {
		log.Printf("%s\n", api)
	}
}

func (s *Source) ConfigureMemDB() {
	if s.Storage == nil || len(s.Storage.Entities) == 0 {
		s.Storage.DB = nil
		return
	}

	// Create a new database
	var err error
	s.Storage.DB, err = memdb.NewMemDB(s.Storage.buildSchema())
	if err != nil {
		log.Fatalf("Unable to start database: %v", err)
		return
	}

	// Load test data
	if s.Storage.Data != nil {
		// Create a writeable transaction
		txn := s.Storage.DB.Txn(true)
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

		//// Create read-only transaction
		//txn = s.DB.Txn(false)
		//defer txn.Abort()
		//
		//// Lookup by email
		//var raw interface{}
		//raw, err = txn.First("person", "id", "jason.figge@gmail.com")
		//if err != nil {
		//	panic(err)
		//}
		//
		//// Say hi!
		//table := s.Storage.Entities["person"].Table
		//fmt.Printf("Hello %v!\n", table.toMap(raw)["Name"])
		//
		//// List all the people
		//var it memdb.ResultIterator
		//it, err = txn.Get("person", "id")
		//if err != nil {
		//	panic(err)
		//}
		//
		//fmt.Println("All the people:")
		//for obj := it.Next(); obj != nil; obj = it.Next() {
		//	p := table.toMap(obj)
		//	fmt.Printf("  %-8s%v\n", fmt.Sprintf("%s:", p["Name"]), p["Age"])
		//}
		//
		//// Range scan over people with ages between 25 and 35 inclusive
		//it, err = txn.LowerBound("person", "age", 49)
		//if err != nil {
		//	panic(err)
		//}
		//fmt.Println("All the people over 50:")
		//for obj := it.Next(); obj != nil; obj = it.Next() {
		//	p := table.toMap(obj)
		//	fmt.Printf("  %-8s%v\n", fmt.Sprintf("%s:", p["Name"]), p["Age"])
		//}
	}
}

func (ps *Paths) MatchPath(ctx context.Context, path string) (*Path, bool) {
	attr := ctx.Value(handlers.Attributes).(map[string]interface{})
	p, ok := ps.static[path]
	if ok {
		attr[handlers.AttrPathArgs] = make(map[string]string)
		return p, true
	}

	parts := strings.Split(path, "/")
next:
	for _, pathMeta := range ps.dynamic {
		if len(pathMeta.parts) != len(parts) {
			continue
		}
		arguments := make(map[string]string)
		for index, part := range pathMeta.parts {
			name, found := pathMeta.arguments[index]
			if found {
				arguments[fmt.Sprintf("{%s}", name)] = parts[index]
				continue
			}
			if !strings.EqualFold(part, parts[index]) {
				continue next
			}
		}
		// matched
		attr[handlers.AttrPathArgs] = arguments
		return pathMeta.path, true
	}
	// check dynamic content
	return nil, false
}

func (ps *Paths) UnmarshalYAML(value *yaml.Node) error {
	ps.audit = []string{}
	ps.static = make(map[string]*Path)
	ps.dynamic = make([]*PathMeta, 0)
	for index := 0; index < len(value.Content); index += 2 {
		name := lower.String(value.Content[index].Value)
		if strings.HasPrefix(name, "/") {
			name = name[1:]
		}
		path := &Path{}
		err := value.Content[index+1].Decode(path)
		if err != nil {
			return err
		}
		ps.processPath(name, path)
	}
	return nil
}

func (ps *Paths) UnmarshalJSON(data []byte) error {
	ps.audit = []string{}
	ps.static = make(map[string]*Path)
	ps.dynamic = make([]*PathMeta, 0)
	tmp := make(map[string]*Path)
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	for name, path := range tmp {
		name = lower.String(name)
		if strings.HasPrefix(name, "/") {
			name = name[1:]
		}
		ps.processPath(name, path)
	}
	return nil
}

func (ps *Paths) processPath(name string, path *Path) {
	if strings.Contains(name, "{") {
		ps.dynamic = append(ps.dynamic, newPathMeta(name, path))
	} else {
		ps.static[name] = path
	}
	for method := range path.Methods {
		ps.audit = append(ps.audit, fmt.Sprintf("  %-6s %s", fmt.Sprintf("%s:", method), name))
	}
}

func newPathMeta(name string, path *Path) *PathMeta {
	meta := &PathMeta{
		parts:     strings.Split(name, "/"),
		arguments: make(map[int]string),
		path:      path,
	}
	for index, part := range meta.parts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			meta.arguments[index] = part[1 : len(part)-1]
		}
	}
	return meta
}

func atoi(val string) int {
	i, err := strconv.Atoi(val)
	if err != nil {
		fmt.Printf("Debug: Invalid default: %s: %v\n", val, err)
		return 0
	}
	return i
}
