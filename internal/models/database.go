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
	lower = cases.Lower(language.Und)
	title = cases.Title(language.Und)
)

type Data map[string]interface{}
type Fields map[string]datatype.DataType

type Entities map[string]*Entity

type Index struct {
	Name   string   `json:"name" yaml:"name"`
	Field  string   `json:"field,omitempty" yaml:"field,omitempty"`
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
}

func (s *Store) buildSchema() *memdb.DBSchema {
	tables := make(map[string]*memdb.TableSchema)
	for entityName, definition := range s.Entities {
		table := memdb.TableSchema{
			Name: entityName,
		}
		indexes := make(map[string]*memdb.IndexSchema)
		for _, index := range definition.Indexes {
			indexes[lower.String(index.Name)] = &memdb.IndexSchema{
				Name:    lower.String(index.Name),
				Unique:  index.Unique,
				Indexer: definition.Table.fields.Indexer(index),
			}
		}
		table.Indexes = indexes
		tables[entityName] = &table
	}
	return &memdb.DBSchema{
		Tables: tables,
	}
}

func (t *Table) UnmarshalYAML(value *yaml.Node) error {
	*t = Table{
		fields: make(Fields),
	}
	for index := 0; index < len(value.Content); index += 2 {
		name := lower.String(value.Content[index].Value)
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
		t.fields[lower.String(name)] = d
	}
	t.mapToStruct()
	return nil
}

func (t *Table) mapToStruct() {
	var structFields []reflect.StructField

	for k, v := range t.fields {
		sf := reflect.StructField{
			Name: title.String(k),
		}
		switch v {
		case datatype.String:
			sf.Type = reflect.TypeOf("")
		case datatype.Int:
			sf.Type = reflect.TypeOf(int64(0))
		case datatype.Bool:
			sf.Type = reflect.TypeOf(false)
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
		d := t.fields[lower.String(k)]
		fv := s.Elem().FieldByName(title.String(k))
		switch d {
		case datatype.String:
			fv.SetString(v.(string))
		case datatype.Int:
			switch x := v.(type) {
			case float64: // json unmarshalling of int into a interface{}
				fv.SetInt(int64(x))
			case int:
				fv.SetInt(int64(x))
			}
		case datatype.Bool:
			fv.SetBool(v.(bool))
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

func (f Fields) Indexer(index Index) memdb.Indexer {
	fieldName := index.Name
	if index.Field != "" {
		fieldName = index.Field
	} else if len(index.Fields) > 0 {
		fieldName = index.Fields[0]
	}
	fieldName = title.String(fieldName)
	d, ok := f[lower.String(fieldName)]
	if !ok {
		log.Fatalf("unknown index field: %s", index.Name)
		return nil
	}
	switch d {
	case datatype.String:
		return &memdb.StringFieldIndex{Field: fieldName}
	case datatype.Int:
		return &memdb.IntFieldIndex{Field: fieldName}
	case datatype.Bool:
		return &memdb.BoolFieldIndex{Field: fieldName}
	}
	return nil
}
