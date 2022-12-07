package models

import (
	"fmt"
	"strings"
)

// DataType - Basic data types for an in memory database
type DataType int

// Declare related constants for each weekday starting with index 1
const (
	String    DataType = iota + 1 // EnumIndex = 1
	Int                           // EnumIndex = 2
	Float                         // EnumIndex = 3
	Timestamp                     // EnumIndex = 4
)

// String - Creating common behavior - give the type a String function
func (d DataType) String() string {
	return [...]string{"String", "Int", "Float", "Timestamp"}[d-1]
}

// EnumIndex - Creating common behavior - give the type a EnumIndex function
func (d DataType) EnumIndex() int {
	return int(d)
}

func DataTypeOf(value string) DataType {
	switch strings.ToLower(value) {
	case "string":
		return String
	case "int":
		return Int
	case "float":
		return Float
	case "timestamp":
		return Timestamp
	}
	panic(fmt.Sprintf("unknown value: %s", value))
}
