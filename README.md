# Go Simplifier

This Go package simplifies your data structures according to the rules you define. It's perfect for reducing complex structures to just the data you need, providing a cleaner, simplified view of your data. The Go Simplifier is especially useful when dealing with large data structures where only a subset of the fields are required.

## Installation

Install Go Simplifier with:

```
go get github.com/xhinliang/gosimplifier
```

Make sure to replace `xhinliang` with your GitHub username.

## Usage

Here is a simple example of how to use Go Simplifier:

```go
package main

import (
	"fmt"
	"github.com/xhinliang/gosimplifier"
)

type MyStruct struct {
	Field1 int
	Field2 string
}

func main() {
	rulesJson := `{
		"remove_properties": ["Field1"]
	}`

	simplifier, err := gosimplifier.NewSimplifier(rulesJson)
	if err != nil {
		// handle error
	}

	original := MyStruct{
		Field1: 1,
		Field2: "data",
	}

	simplified, err := simplifier.Simplify(original)
	if err != nil {
		// handle error
	}

	fmt.Printf("%v\n", simplified)
}
```

This will output:

```sh
{  data}
```

You can also use `NewSimplifierByRule` to create a simplifier from a `Rule` struct:

```go
rule := &gosimplifier.Rule{
	RemoveProperties:    []string{"Field1"},
	PropertySimplifiers: map[string]*gosimplifier.Rule{
        
    },
}

simplifier, err := gosimplifier.NewSimplifierByRule(rule)
// ...
```

## License

This project is licensed under the terms of the Apache 2.0 license. For more information, please see the [LICENSE](LICENSE) file.

## Contributing

Contributions to Go Simplifier are very welcome! Please submit a pull request or create an issue on the GitHub repository.

## Testing

Go Simplifier includes a suite of unit tests. To run them, use `go test`:

```sh
go test ./...
```

## Contact

If you encounter any issues or have questions about Go Simplifier, feel free to reach out via GitHub issues.
