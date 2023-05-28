package gosimplifier

import (
	"testing"
)

func BenchmarkSimplify(b *testing.B) {
	// Define an example struct
	type ExampleStruct struct {
		Data map[string]interface{}
	}

	// Define the rules JSON
	rulesJSON := `
{
  "remove_properties": [
    "Data"
  ]
}
`

	// Create a new Simplifier instance
	simplifier, err := NewSimplifier(rulesJSON)
	if err != nil {
		b.Fatalf("Failed to create Simplifier: %v", err)
	}

	// Create an example struct with nested map
	original := &ExampleStruct{
		Data: map[string]interface{}{
			"nested": map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
	}

	// Run the benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := simplifier.Simplify(original)
		if err != nil {
			b.Fatalf("Failed to simplify struct: %v", err)
		}
	}
}
