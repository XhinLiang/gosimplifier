package gosimplifier

import (
	"encoding/json"
	"reflect"
	"testing"
)

type ExampleStruct0 struct {
	Test       int
	Debug      string
	Data       DataStruct
	EntityList []EntityStruct
}

type ExampleStruct struct {
	Test       int
	Debug      string
	Data       DataStruct
	EntityList []EntityStruct
	Nest       ExampleStruct0
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
		"remove_properties": [ "Debug" ],
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
		Nest: ExampleStruct0{
			Debug: "debug",
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
	deepCheck(t, simplifiedStruct, ExampleStruct{
		Test: 5,
		EntityList: []EntityStruct{
			{
				SubProperties: SubPropertyStruct{
					ABC: "",
					DEF: "",
				},
			},
		},
	})
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
	deepCheck(t, simplifiedStruct, original)
}

func deepCheck(t *testing.T, simplified ExampleStruct, expect ExampleStruct) {
	if simplified.Test != expect.Test {
		t.Error("Expected Test to be unchanged")
	}
	if simplified.Debug != expect.Debug {
		t.Error("Expected Debug to be unchanged")
	}
	if simplified.Data.DataTest != expect.Data.DataTest {
		t.Error("Expected Data.DataTest to be unchanged")
	}
	if simplified.Data.DataDebug != expect.Data.DataDebug {
		t.Error("Expected Data.DataDebug to be unchanged")
	}
	if len(simplified.EntityList) != len(expect.EntityList) {
		t.Fatal("Expected EntityList to be unchanged")
	}
	for i, entity := range simplified.EntityList {
		if entity.SubProperties.ABC != expect.EntityList[i].SubProperties.ABC {
			t.Error("Expected EntityList.SubProperties.ABC to be unchanged")
		}
		if entity.SubProperties.DEF != expect.EntityList[i].SubProperties.DEF {
			t.Error("Expected EntityList.SubProperties.DEF to be unchanged")
		}
	}
	if simplified.Nest.Debug != expect.Nest.Debug {
		t.Error("Expected Nest.Debug to be unchanged")
	}
	if simplified.Nest.Data.DataTest != expect.Nest.Data.DataTest {
		t.Error("Expected Nest.Data.DataTest to be unchanged")
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
	deepCheck(t, simplifiedStruct, original)
}

func TestSimplifyInvalidPropertyName0(t *testing.T) {
	rulesJson := `{
		"remove_properties": [ "field1" ]
	}`

	simplifier, _ := NewSimplifier(rulesJson)

	original := map[string]interface{}{
		"field1": 5,
		"field2": "debug",
	}

	simplified, err := simplifier.Simplify(original)
	if err != nil {
		t.Error("Unexpected error", err)
	}

	simplifiedMap, ok := simplified.(map[string]interface{})
	if !ok {
		t.Error("Expected ExampleStruct, but got different type")
	}
	if simplifiedMap["field1"] != nil {
		t.Error("Expected field1 to be nil")
	}
	if simplifiedMap["field2"] != "debug" {
		t.Error("Expected field2 to be unchanged")
	}
	simplifiedJson, jsonErr := json.Marshal(simplifiedMap)
	if jsonErr != nil {
		t.Error("Unexpected error", jsonErr)
	}
	t.Log("simplifiedMap", string(simplifiedJson))
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
	deepCheck(t, simplifiedStruct, original)
}

func TestExtendSimplifier0(t *testing.T) {
	rulesJsonBase := `{
		"remove_properties": [ "Test", "Debug" ],
		"property_simplifiers": {
			"Data": {
				"remove_properties": [ "DataTest" ]
			}
		}
	}`

	rulesJsonExtend := `{
		"remove_properties": [ "Debug" ],
		"property_simplifiers": {
			"Data": {
				"remove_properties": [ "DataDebug" ]
			}
		}
	}`

	baseSimplifier, _ := NewSimplifier(rulesJsonBase)
	extendedSimplifier, err := ExtendSimplifier(baseSimplifier, rulesJsonExtend)
	if err != nil {
		t.Error("Unexpected error", err)
	}

	original := ExampleStruct{
		Test:  5,
		Debug: "debug",
		Data: DataStruct{
			DataTest:  "data_test",
			DataDebug: 123,
		},
	}

	simplified, err := extendedSimplifier.Simplify(original)
	if err != nil {
		t.Error("Unexpected error", err)
	}

	simplifiedStruct, ok := simplified.(ExampleStruct)
	if !ok {
		t.Error("Expected ExampleStruct, but got different type")
	}
	if simplifiedStruct.Test != 0 {
		t.Error("Expected Test to be removed by the base simplifier rules")
	}
	if simplifiedStruct.Debug != "" {
		t.Error("Expected Debug to be removed by the base and extended simplifier rules")
	}
	if simplifiedStruct.Data.DataTest != "" {
		t.Error("Expected Data.DataTest to be removed by the base simplifier rules")
	}
	if simplifiedStruct.Data.DataDebug != 0 {
		t.Error("Expected Data.DataDebug to be removed by the extended simplifier rules")
	}
}

type ExampleStruct2 struct {
	Name     string
	Age      int
	Data     string
	Info     *SubStruct
	NewField *AnotherStruct
}

type SubStruct struct {
	Test  string
	Debug string
}

type AnotherStruct struct {
	SubTest string
}

func TestExtendSimplifier(t *testing.T) {
	baseRulesJson := `{
		"remove_properties": ["Name", "Data"],
		"property_simplifiers": {
			"Info": {
				"remove_properties": ["Debug"]
			}
		}
	}`

	extendRulesJson := `{
		"remove_properties": ["Age"],
		"property_simplifiers": {
			"Info": {
				"remove_properties": ["Test"]
			},
			"NewField": {
				"remove_properties": ["SubTest"]
			}
		}
	}`

	original := &ExampleStruct2{
		Name: "John Doe",
		Age:  30,
		Data: "Data",
		Info: &SubStruct{
			Test:  "Test Value",
			Debug: "Debug Value",
		},
		NewField: &AnotherStruct{
			SubTest: "SubTest Value",
		},
	}

	expected := &ExampleStruct2{
		Name:     "",
		Age:      0,
		Data:     "",
		Info:     nil,
		NewField: nil,
	}

	baseSimplifier, err := NewSimplifier(baseRulesJson)
	if err != nil {
		t.Fatal(err)
	}

	extendSimplifier, err := ExtendSimplifier(baseSimplifier, extendRulesJson)
	if err != nil {
		t.Fatal(err)
	}

	result, err := extendSimplifier.Simplify(original)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestExtendSimplifierEmptyBase(t *testing.T) {
	baseRulesJson := `{}`
	extendRulesJson := `{
		"remove_properties": ["Name", "Age"]
	}`

	original := &ExampleStruct2{
		Name: "John Doe",
		Age:  30,
	}

	expected := &ExampleStruct2{}

	baseSimplifier, err := NewSimplifier(baseRulesJson)
	if err != nil {
		t.Fatal(err)
	}

	extendSimplifier, err := ExtendSimplifier(baseSimplifier, extendRulesJson)
	if err != nil {
		t.Fatal(err)
	}

	result, err := extendSimplifier.Simplify(original)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestExtendSimplifierEmptyExtend(t *testing.T) {
	baseRulesJson := `{
		"remove_properties": ["Name", "Age"]
	}`
	extendRulesJson := `{}`

	original := &ExampleStruct2{
		Name: "John Doe",
		Age:  30,
	}

	expected := &ExampleStruct2{}

	baseSimplifier, err := NewSimplifier(baseRulesJson)
	if err != nil {
		t.Fatal(err)
	}

	extendSimplifier, err := ExtendSimplifier(baseSimplifier, extendRulesJson)
	if err != nil {
		t.Fatal(err)
	}

	result, err := extendSimplifier.Simplify(original)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestExtendSimplifierWithConflictingRules(t *testing.T) {
	baseRulesJson := `{
		"remove_properties": ["Name"],
		"property_simplifiers": {
			"Info": {
				"remove_properties": ["Test"]
			}
		}
	}`

	extendRulesJson := `{
		"remove_properties": ["Age"],
		"property_simplifiers": {
			"Info": {
				"remove_properties": ["Debug"]
			}
		}
	}`

	original := &ExampleStruct2{
		Name: "John Doe",
		Age:  30,
		Info: &SubStruct{
			Test:  "Test Value",
			Debug: "Debug Value",
		},
	}

	expected := &ExampleStruct2{
		Name: "",
		Age:  0,
		Info: nil,
	}

	baseSimplifier, err := NewSimplifier(baseRulesJson)
	if err != nil {
		t.Fatal(err)
	}

	extendSimplifier, err := ExtendSimplifier(baseSimplifier, extendRulesJson)
	if err != nil {
		t.Fatal(err)
	}

	result, err := extendSimplifier.Simplify(original)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
