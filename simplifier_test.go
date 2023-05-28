package gosimplifier

import (
	"reflect"
	"testing"
)

type ExampleStruct struct {
	Test       int
	Debug      string
	Data       DataStruct
	EntityList []EntityStruct
}

type DataStruct struct {
	DataTest  string
	DataDebug int
}

type EntityStruct struct {
	SubProperties SubPropertyStruct
}

type SubPropertyStruct struct {
	ABC string
	DEF string
}

func TestNewSimplifier(t *testing.T) {
	rulesJson := `{
		"remove_properties": [ "Test", "Debug" ],
		"property_simplifiers": {
			"Data": {
				"remove_properties": [ "DataTest", "DataDebug" ]
			},
			"EntityList": {
				"property_simplifiers": {
					"SubProperties": {
						"remove_properties": [ "ABC", "DEF" ]
					}
				}
			}
		}
	}`

	simplifier, err := NewSimplifier(rulesJson)
	if err != nil {
		t.Error("Unexpected error", err)
	}
	if simplifier == nil {
		t.Error("Expected simplifier to be not nil")
	}
}

func TestNewSimplifierWithInvalidJson(t *testing.T) {
	rulesJson := `{ This is an invalid JSON string }`

	simplifier, err := NewSimplifier(rulesJson)
	if err == nil {
		t.Error("Expected error, but got none")
	}
	if simplifier != nil {
		t.Error("Expected simplifier to be nil")
	}
}

func TestSimplify(t *testing.T) {
	rulesJson := `{
		"remove_properties": [ "Test", "Debug" ],
		"property_simplifiers": {
			"Data": {
				"remove_properties": [ "DataTest", "DataDebug" ]
			},
			"EntityList": {
				"property_simplifiers": {
					"SubProperties": {
						"remove_properties": [ "ABC", "DEF" ]
					}
				}
			}
		}
	}`

	simplifier, _ := NewSimplifier(rulesJson)

	original := ExampleStruct{
		Test:  5,
		Debug: "debug",
		Data: DataStruct{
			DataTest:  "data_test",
			DataDebug: 123,
		},
		EntityList: []EntityStruct{
			{
				SubProperties: SubPropertyStruct{
					ABC: "abc",
					DEF: "def",
				},
			},
		},
	}

	simplified, err := simplifier.Simplify(original)
	if err != nil {
		t.Error("Unexpected error", err)
	}

	simplifiedStruct, ok := simplified.(ExampleStruct)
	if !ok {
		t.Error("Expected ExampleStruct, but got different type")
	}

	if simplifiedStruct.Test != 0 {
		t.Error("Expected Test to be zero")
	}
	if simplifiedStruct.Debug != "" {
		t.Error("Expected Debug to be zero value")
	}
	if simplifiedStruct.Data.DataTest != "" {
		t.Error("Expected Data.DataTest to be zero value")
	}
	if simplifiedStruct.Data.DataDebug != 0 {
		t.Error("Expected Data.DataDebug to be zero value")
	}
	for _, entity := range simplifiedStruct.EntityList {
		if entity.SubProperties.ABC != "" {
			t.Error("Expected EntityList.SubProperties.ABC to be zero value")
		}
		if entity.SubProperties.DEF != "" {
			t.Error("Expected EntityList.SubProperties.DEF to be zero value")
		}
	}
}

func TestNewSimplifierEmptyJson(t *testing.T) {
	rulesJson := `{}`

	simplifier, err := NewSimplifier(rulesJson)
	if err != nil {
		t.Error("Unexpected error", err)
	}
	if simplifier == nil {
		t.Error("Expected simplifier to be not nil")
	}
}

func TestSimplifyInvalidType(t *testing.T) {
	rulesJson := `{
		"remove_properties": [ "Test", "Debug" ]
	}`

	simplifier, _ := NewSimplifier(rulesJson)

	// Using an int instead of a struct
	original := 10

	simplified, err := simplifier.Simplify(original)
	if err != nil {
		t.Error("Unexpected error", err)
	}

	simplifiedInt, ok := simplified.(int)
	if !ok {
		t.Error("Expected int, but got different type")
	}
	if simplifiedInt != 10 {
		t.Error("Expected value to remain unchanged")
	}
}

func TestSimplifyNoRules(t *testing.T) {
	rulesJson := `{}`

	simplifier, _ := NewSimplifier(rulesJson)

	original := ExampleStruct{
		Test:  5,
		Debug: "debug",
		Data: DataStruct{
			DataTest:  "data_test",
			DataDebug: 123,
		},
		EntityList: []EntityStruct{
			{
				SubProperties: SubPropertyStruct{
					ABC: "abc",
					DEF: "def",
				},
			},
		},
	}

	simplified, err := simplifier.Simplify(original)
	if err != nil {
		t.Error("Unexpected error", err)
	}

	simplifiedStruct, ok := simplified.(ExampleStruct)
	if !ok {
		t.Error("Expected ExampleStruct, but got different type")
	}
	if !reflect.DeepEqual(simplifiedStruct, original) {
		t.Error("Expected original and simplified to be identical")
	}
}

func TestNewSimplifierInvalidPropertyName(t *testing.T) {
	rulesJson := `{
		"remove_properties": [ "NonexistentProperty" ]
	}`

	simplifier, err := NewSimplifier(rulesJson)
	if err != nil {
		t.Error("Unexpected error", err)
	}
	if simplifier == nil {
		t.Error("Expected simplifier to be not nil")
	}
}

func TestSimplifyInvalidPropertyName(t *testing.T) {
	rulesJson := `{
		"remove_properties": [ "NonexistentProperty" ]
	}`

	simplifier, _ := NewSimplifier(rulesJson)

	original := ExampleStruct{
		Test:  5,
		Debug: "debug",
		Data: DataStruct{
			DataTest:  "data_test",
			DataDebug: 123,
		},
		EntityList: []EntityStruct{
			{
				SubProperties: SubPropertyStruct{
					ABC: "abc",
					DEF: "def",
				},
			},
		},
	}

	simplified, err := simplifier.Simplify(original)
	if err != nil {
		t.Error("Unexpected error", err)
	}

	simplifiedStruct, ok := simplified.(ExampleStruct)
	if !ok {
		t.Error("Expected ExampleStruct, but got different type")
	}
	if !reflect.DeepEqual(simplifiedStruct, original) {
		t.Error("Expected original and simplified to be identical when removing non-existent properties")
	}
}

func TestNewSimplifierEmptyPropertyName(t *testing.T) {
	rulesJson := `{
		"remove_properties": [ "" ]
	}`

	simplifier, err := NewSimplifier(rulesJson)
	if err != nil {
		t.Error("Unexpected error", err)
	}
	if simplifier == nil {
		t.Error("Expected simplifier to be not nil")
	}
}

func TestSimplifyEmptyPropertyName(t *testing.T) {
	rulesJson := `{
		"remove_properties": [ "" ]
	}`

	simplifier, _ := NewSimplifier(rulesJson)

	original := ExampleStruct{
		Test:  5,
		Debug: "debug",
		Data: DataStruct{
			DataTest:  "data_test",
			DataDebug: 123,
		},
		EntityList: []EntityStruct{
			{
				SubProperties: SubPropertyStruct{
					ABC: "abc",
					DEF: "def",
				},
			},
		},
	}

	simplified, err := simplifier.Simplify(original)
	if err != nil {
		t.Error("Unexpected error", err)
	}

	simplifiedStruct, ok := simplified.(ExampleStruct)
	if !ok {
		t.Error("Expected ExampleStruct, but got different type")
	}
	if !reflect.DeepEqual(simplifiedStruct, original) {
		t.Error("Expected original and simplified to be identical when removing empty property name")
	}
}
