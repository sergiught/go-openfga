package dsl_test

import (
	"fmt"

	"github.com/sergiught/go-openfga/dsl"
	"github.com/sergiught/go-openfga/openfga"
)

// ExampleToModel parses OpenFGA DSL text into a request ready for
// AuthorizationModels.Write.
func ExampleToModel() {
	const model = `model
  schema 1.1

type user

type document
  relations
    define viewer: [user]`

	req, err := dsl.ToModel(model)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("schema:", req.SchemaVersion)
	fmt.Println("types:", len(req.TypeDefinitions))
	// Output:
	// schema: 1.1
	// types: 2
}

// ExampleToDSL renders an authorization model back to OpenFGA DSL text.
func ExampleToDSL() {
	req, err := dsl.ToModel("model\n  schema 1.1\n\ntype user")
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	text, err := dsl.ToDSL(&openfga.AuthorizationModel{
		SchemaVersion:   req.SchemaVersion,
		TypeDefinitions: req.TypeDefinitions,
		Conditions:      req.Conditions,
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(text)
}
