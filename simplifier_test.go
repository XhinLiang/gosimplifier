package simplifier

import (
	"reflect"
	"testing"
)

// Example structs for testing
type ExampleStruct struct {
	Test       string
	Debug      string
	Data       []Data
	EntityList []Entity
}

type Data struct {
	DataTest  string
	DataDebug string
}

type Entity struct {
	SubProperties []SubProperties
	EntityTest    string
}

type SubProperties struct {
	ABC string
	DEF string
}

func TestSimplify(t *testing.T) {
	// Define the original struct for testing
	original := ExampleStruct{
		Test:  "test value",
		Debug: "debug value",
		Data: []Data{
			{
				DataTest:  "data test value",
				DataDebug: "data debug value",
			},
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
	}

	// Define the expected simplified struct based on the provided rules
	expected := ExampleStruct{
		Data: []Data{
			{
				DataDebug: "data debug value",
			},
		},
		EntityList: []Entity{
			{
				SubProperties: []SubProperties{
					{},
				},
			},
		},
	}

	// Define the rules JSON
	rulesJSON := `
{
  "remove_properties": [
    "test",
    "debug"
  ],
  "property_rules": {
    "data": {
      "remove_properties": [
        "data_test",
        "data_debug"
      ],
      "property_rules": {}
    },
    "entity_list": {
      "remove_properties": [
        "entity_test"
      ],
      "property_rules": {
        "sub_properties": {
          "remove_properties": [
            "abc",
            "def"
          ]
        }
      }
    }
  }
}
`

	// Create a new simplifier with the rules
	simplifier, err := NewSimplifier(nil, rulesJSON)
	if err != nil {
		t.Fatalf("Error creating simplifier: %v", err)
	}

	// Simplify the original struct
	simplified, err := simplifier.Simplify(original)
	if err != nil {
		t.Fatalf("Error simplifying struct: %v", err)
	}

	// Compare the simplified struct with the expected struct
	if !reflect.DeepEqual(simplified, expected) {
		t.Errorf("Simplified struct does not match expected struct.\nExpected: %+v\nActual: %+v", expected, simplified)
	}
}

func TestNewSimplifierWithOrigin(t *testing.T) {
	// Define the original rules for the origin Simplifier
	originRulesJSON := `
{
  "property_rules": {
    "data": {
      "remove_properties": [
        "data_test"
      ],
      "property_rules": {}
    }
  }
}
`

	// Create the origin Simplifier
	origin, err := NewSimplifier(nil, originRulesJSON)
	if err != nil {
		t.Fatalf("Error creating origin Simplifier: %v", err)
	}

	// Define the additional rules to be added to the origin Simplifier
	additionalRulesJSON := `
{
  "remove_properties": [
    "debug"
  ],
  "property_rules": {
    "data": {
      "remove_properties": [
        "data_debug"
      ],
      "property_rules": {}
    },
    "entity_list": {
      "remove_properties": [
        "entity_test"
      ],
      "property_rules": {
        "sub_properties": {
          "remove_properties": [
            "abc",
            "def"
          ]
        }
      }
    }
  }
}
`

	// Create a new simplifier based on the origin Simplifier and additional rules
	simplifier, err := NewSimplifier(origin, additionalRulesJSON)
	if err != nil {
		t.Fatalf("Error creating new Simplifier with origin: %v", err)
	}

	// Simplify an example struct using the new simplifier
	original := ExampleStruct{
		Test:       "test value",
		Debug:      "debug value",
		Data:       []Data{{DataTest: "data test value", DataDebug: "data debug value"}},
		EntityList: []Entity{{EntityTest: "entity test value"}},
	}

	simplified, err := simplifier.Simplify(original)
	if err != nil {
		t.Fatalf("Error simplifying struct: %v", err)
	}

	// Define the expected simplified struct based on the origin rules and additional rules
	expected := ExampleStruct{
		Data:       []Data{{DataTest: "data test value"}},
		EntityList: []Entity{{}},
	}

	// Compare the simplified struct with the expected struct
	if !reflect.DeepEqual(simplified, expected) {
		t.Errorf("Simplified struct does not match expected struct.\nExpected: %+v\nActual: %+v", expected, simplified)
	}
}

func TestNewSimplifierWithoutOrigin(t *testing.T) {
	// Define the rules for the new Simplifier
	rulesJSON := `
{
  "remove_properties": [
    "test",
    "debug"
  ],
  "property_rules": {
    "data": {
      "remove_properties": [
        "data_test",
        "data_debug"
      ],
      "property_rules": {}
    },
    "entity_list": {
      "remove_properties": [
        "entity_test"
      ],
      "property_rules": {
        "sub_properties": {
          "remove_properties": [
            "abc",
            "def"
          ]
        }
      }
    }
  }
}
`

	// Create a new simplifier without an origin
	simplifier, err := NewSimplifier(nil, rulesJSON)
	if err != nil {
		t.Fatalf("Error creating new Simplifier without origin: %v", err)
	}

	// Simplify an example struct using the new simplifier
	original := ExampleStruct{
		Test:       "test value",
		Debug:      "debug value",
		Data:       []Data{{DataTest: "data test value", DataDebug: "data debug value"}},
		EntityList: []Entity{{EntityTest: "entity test value"}},
	}

	simplified, err := simplifier.Simplify(original)
	if err != nil {
		t.Fatalf("Error simplifying struct: %v", err)
	}

	// Define the expected simplified struct based on the rules
	expected := ExampleStruct{
		Data:       []Data{{DataDebug: "data debug value"}},
		EntityList: []Entity{{SubProperties: []SubProperties{{}}}},
	}

	// Compare the simplified struct with the expected struct
	if !reflect.DeepEqual(simplified, expected) {
		t.Errorf("Simplified struct does not match expected struct.\nExpected: %+v\nActual: %+v", expected, simplified)
	}
}
