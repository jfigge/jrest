package models

import (
	"encoding/json"
	"fmt"
	"jrest/internal/models/enums/datatype"
	"log"
	"reflect"

	"github.com/hashicorp/go-memdb"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

var (
	lowerCase = cases.Lower(language.Und)
	titleCase = cases.Title(language.Und)
)

type Data map[string]interface{}
type Fields map[string]datatype.DataType
type Entities map[string]*Entity

type Index struct {
	Name   string   `json:"name" yaml:"name"`
	Fields []string `json:"fields,omitempty" yaml:"fields,omitempty"`
	Unique bool     `json:"unique,omitempty" yaml:"unique,omitempty"`
}
type Entity struct {
	Table   Table   `json:"fields" yaml:"fields"`
	Indexes []Index `json:"indexes" yaml:"indexes"`
}
type Table struct {
	structType reflect.Type
	fields     Fields
}
type Store struct {
	Entities Entities          `json:"entities" yaml:"entities"`
	Data     map[string][]Data `json:"data,omitempty" yaml:"data,omitempty"`
	schema   *memdb.DBSchema
}

func (s *Store) Schema() *memdb.DBSchema {
	if s.schema == nil {
		tables := make(map[string]*memdb.TableSchema)
		for entityName, definition := range s.Entities {
			table := memdb.TableSchema{
				Name: entityName,
			}
			indexes := make(map[string]*memdb.IndexSchema)
			for _, content := range definition.Indexes {
				index := memdb.IndexSchema{
					Name:   lowerCase.String(content.Name),
					Unique: content.Unique,
				}
				if len(content.Fields) > 0 {
					index.Indexer = &memdb.StringFieldIndex{Field: titleCase.String(content.Fields[0])}
				} else {
					index.Indexer = &memdb.StringFieldIndex{Field: titleCase.String(content.Name)}
				}
				indexes[content.Name] = &index
			}
			table.Indexes = indexes
			tables[entityName] = &table
		}
		s.schema = &memdb.DBSchema{
			Tables: tables,
		}
	}
	return s.schema
}

func (t *Table) UnmarshalYAML(value *yaml.Node) error {
	*t = Table{
		fields: make(Fields),
	}
	for index := 0; index < len(value.Content); index += 2 {
		name := value.Content[index].Value
		d := datatype.DataTypeOf(value.Content[index+1].Value)
		t.fields[name] = d
	}
	t.mapToStruct()
	return nil
}

func (t *Table) UnmarshalJSON(data []byte) error {
	tmp := make(map[string]string)
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	*t = Table{
		fields: make(map[string]datatype.DataType),
	}
	for name, value := range tmp {
		d := datatype.DataTypeOf(value)
		t.fields[name] = d
	}
	return nil
}

func (t *Table) mapToStruct() {
	var structFields []reflect.StructField

	for k, v := range t.fields {
		sf := reflect.StructField{
			Name: titleCase.String(k),
		}
		switch v {
		case datatype.String:
			sf.Type = reflect.TypeOf("")
		case datatype.Int:
			sf.Type = reflect.TypeOf(0)
		case datatype.Float:
			sf.Type = reflect.TypeOf(float64(0))
		}
		structFields = append(structFields, sf)
	}

	// Creates the struct type
	t.structType = reflect.StructOf(structFields)
}

func (t *Table) getInstance() reflect.Value {
	return reflect.New(t.structType)
}

func (t *Table) setValues(s reflect.Value, row Data) reflect.Value {
	for k, v := range row {
		fv := s.Elem().FieldByName(titleCase.String(k))
		switch x := v.(type) {
		case string:
			fv.SetString(x)
		case int:
			fv.SetInt(int64(x))
		case float64:
			fv.SetFloat(x)
		default:
			fmt.Printf("%s\n", v)
		}
	}
	return s
}

func (t *Table) toMap(s interface{}) map[string]interface{} {
	modelReflect := reflect.ValueOf(s)
	if modelReflect.Kind() == reflect.Ptr {
		modelReflect = modelReflect.Elem()
	}

	var fieldData interface{}

	m := make(map[string]interface{})
	for i := 0; i < t.structType.NumField(); i++ {
		field := modelReflect.Field(i)

		switch field.Kind() {
		case reflect.Struct:
			fallthrough
		case reflect.Ptr:
			log.Fatalf("Support for sub-structures has not been implemented: %v", field.String())
			//fieldData = t.toMap(field)
		default:
			fieldData = field.Interface()
		}

		m[t.structType.Field(i).Name] = fieldData
	}
	return m
}

func (t *Table) Fields() Fields {
	return t.fields
}
