package models

import (
	"encoding/json"
	"fmt"
	"jrest/internal/models/enums/datatype"
	"reflect"

	"github.com/hashicorp/go-memdb"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

var (
	titleCase = cases.Title(language.Und)
)

type Data map[string]interface{}
type Fields map[string]datatype.DataType
type Entities map[string]Entity
type Entity struct {
	structType reflect.Type
	fields     Fields
}

type Store struct {
	Entities Entities          `json:"entities" yaml:"entities"`
	Data     map[string][]Data `json:"data,omitempty" yaml:"data,omitempty"`
	Schema   Schema            `json:"schema" yaml:"schema"`
}

// Create the DB schema
var (
	schema = &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"person": &memdb.TableSchema{
				Name: "person",
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "Email"},
					},
					"age": &memdb.IndexSchema{
						Name:    "age",
						Unique:  false,
						Indexer: &memdb.IntFieldIndex{Field: "Age"},
					},
				},
			},
		},
	}
)

func (e *Entity) UnmarshalYAML(value *yaml.Node) error {
	*e = Entity{
		fields: make(Fields),
	}
	for index := 0; index < len(value.Content); index += 2 {
		name := value.Content[index].Value
		d := datatype.DataTypeOf(value.Content[index+1].Value)
		e.fields[name] = d
	}
	e.mapToStruct()
	return nil
}

func (e *Entity) UnmarshalJSON(data []byte) error {
	tmp := make(map[string]string)
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	*e = Entity{
		fields: make(map[string]datatype.DataType),
	}
	for name, value := range tmp {
		d := datatype.DataTypeOf(value)
		e.fields[name] = d
	}
	return nil
}

func (e *Entity) mapToStruct() {
	var structFields []reflect.StructField

	for k, v := range e.fields {
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
	e.structType = reflect.StructOf(structFields)
}

func (e *Entity) getInstance() reflect.Value {
	return reflect.New(e.structType)
}

func (e *Entity) setValues(s reflect.Value, row Data) reflect.Value {
	fmt.Println("\n---Setting struct fields...")
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

func (e *Entity) toMap(s interface{}) map[string]interface{} {
	modelReflect := reflect.ValueOf(s)
	if modelReflect.Kind() == reflect.Ptr {
		modelReflect = modelReflect.Elem()
	}

	var fieldData interface{}

	m := make(map[string]interface{})
	for i := 0; i < e.structType.NumField(); i++ {
		field := modelReflect.Field(i)

		switch field.Kind() {
		case reflect.Struct:
			fallthrough
		case reflect.Ptr:
			fieldData = e.toMap(field)
		default:
			fieldData = field.Interface()
		}

		m[e.structType.Field(i).Name] = fieldData
	}
	return m
}

func (e *Entity) Fields() Fields {
	return e.fields
}
