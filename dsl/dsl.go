// Package dsl converts between OpenFGA's DSL model syntax and the JSON model
// types used by the openfga package. It wraps github.com/openfga/language,
// which is why it lives in its own module: the heavy transitive dependency
// stays out of the core SDK's module graph.
package dsl

import (
	"encoding/json"

	language "github.com/openfga/language/pkg/go/transformer"
	"github.com/sergiught/go-openfga/openfga"
)

// ToModel parses OpenFGA DSL text into a WriteAuthorizationModelRequest ready
// to pass to AuthorizationModels.Write.
func ToModel(dslText string) (*openfga.WriteAuthorizationModelRequest, error) {
	jsonStr, err := language.TransformDSLToJSON(dslText)
	if err != nil {
		return nil, err
	}
	var req openfga.WriteAuthorizationModelRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		return nil, err
	}
	return &req, nil
}

// ToDSL renders an authorization model as OpenFGA DSL text.
func ToDSL(model *openfga.AuthorizationModel) (string, error) {
	b, err := json.Marshal(model)
	if err != nil {
		return "", err
	}
	dsl, err := language.TransformJSONStringToDSL(string(b))
	if err != nil {
		return "", err
	}
	return *dsl, nil
}
