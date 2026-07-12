package dsl

import (
	"encoding/json"
	"reflect"
	"testing"

	language "github.com/openfga/language/pkg/go/transformer"
)

// TestTypedModelRoundTripFidelity proves the openfga package's strongly-typed
// model schema loses nothing versus the canonical transformer JSON: DSL -> JSON
// (transformer) -> typed structs -> JSON must be semantically identical to the
// transformer's own output. It exercises every rewrite form and a condition.
func TestTypedModelRoundTripFidelity(t *testing.T) {
	dslText := `model
  schema 1.1

type user

type group
  relations
    define member: [user]

type document
  relations
    define owner: [user]
    define editor: [user, group#member] or owner
    define viewer: [user:*, user]
    define both: viewer and editor
    define restricted: editor but not owner
    define parent: [document]
    define can_read: viewer from parent

type resource
  relations
    define accessor: [user with is_valid]

condition is_valid(x: int) {
  x > 0
}
`
	canonical, err := language.TransformDSLToJSON(dslText)
	if err != nil {
		t.Fatal(err)
	}

	// Transformer JSON -> typed structs -> JSON.
	req, err := ToModel(dslText)
	if err != nil {
		t.Fatal(err)
	}
	roundTripped, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}

	var want, got any
	if err := json.Unmarshal([]byte(canonical), &want); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(roundTripped, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("typed round-trip diverged from transformer JSON\ncanonical:   %s\nroundtripped: %s", canonical, roundTripped)
	}
}
