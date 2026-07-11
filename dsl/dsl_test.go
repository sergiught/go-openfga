package dsl

import (
	"strings"
	"testing"

	"github.com/sergiught/go-openfga/openfga"
)

const sampleDSL = `model
  schema 1.1

type user

type document
  relations
    define viewer: [user]`

func TestToModel_ParsesDSL(t *testing.T) {
	req, err := ToModel(sampleDSL)
	if err != nil {
		t.Fatal(err)
	}
	if req.SchemaVersion != "1.1" {
		t.Fatalf("schema version = %q, want 1.1", req.SchemaVersion)
	}
	var types []string
	for _, td := range req.TypeDefinitions {
		types = append(types, td.Type)
	}
	if strings.Join(types, ",") != "user,document" {
		t.Fatalf("types = %v, want [user document]", types)
	}
}

func TestToDSL_RendersModel(t *testing.T) {
	req, err := ToModel(sampleDSL)
	if err != nil {
		t.Fatal(err)
	}
	model := &openfga.AuthorizationModel{
		SchemaVersion:   req.SchemaVersion,
		TypeDefinitions: req.TypeDefinitions,
		Conditions:      req.Conditions,
	}
	out, err := ToDSL(model)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "type document") || !strings.Contains(out, "define viewer") {
		t.Fatalf("rendered DSL missing content:\n%s", out)
	}
}

func TestToModel_InvalidDSLErrors(t *testing.T) {
	if _, err := ToModel("this is not valid dsl"); err == nil {
		t.Fatal("expected error for invalid DSL")
	}
}
