package gosimplifier

import (
	"reflect"
	"testing"
)

func TestSimplify(t *testing.T) {
	// Define an example struct
	type SubProperties struct {
		ABC string
		DEF string
	}

	type Entity struct {
		SubProperties []SubProperties
		EntityTest    string
	}

	type ExampleStruct struct {
		Test       string
		Debug      bool
		Data       map[string]interface{}
		EntityList []Entity
		OtherField string
	}

	// Define the rules JSON
	rulesJSON := `
{
  "remove_properties": [
    "Test",
    "Debug"
  ],
  "property_simplifiers": {
    "Data": {
      "remove_properties": [
        "data_test",
        "data_debug"
      ]
    },
    "EntityList": {
      "remove_properties": [
        "EntityTest"
      ],
      "property_simplifiers": {
        "SubProperties": {
          "remove_properties": [
            "ABC",
            "DEF"
          ]
        }
      }
    }
  }
}
`

	// Create a new Simplifier instance
	simplifier, err := NewSimplifier(rulesJSON)
	if err != nil {
		t.Fatalf("Failed to create Simplifier: %v", err)
	}

	// Create an example struct
	original := &ExampleStruct{
		Test:  "test value",
		Debug: true,
		Data: map[string]interface{}{
			"data_test":  "data test value",
			"data_debug": false,
		},
		EntityList: []Entity{
			{
				SubProperties: []SubProperties{
					{
						ABC: "abc value",
						DEF: "def value",
					},
				},
				EntityTest: "entity test value",
			},
		},
		OtherField: "other value",
	}
	// Simplify the original struct
	simplified, err := simplifier.Simplify(original)
	if err != nil {
		t.Fatalf("Failed to simplify struct: %v", err)
	}

	// Validate the simplified struct
	expected := &ExampleStruct{
		Test:       "",
		Debug:      false,
		Data:       map[string]interface{}{},
		EntityList: []Entity{{SubProperties: []SubProperties{{ABC: "", DEF: ""}}}},
		OtherField: "other value",
	}

	if !reflect.DeepEqual(simplified, expected) {
		t.Errorf("Simplified struct does not match expected value.\nExpected: %+v\nActual: %+v", expected, simplified)
	}
}
