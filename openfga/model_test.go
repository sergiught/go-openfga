package openfga

import (
	"encoding/json"
	"testing"
)

func TestUsersetBuilders_JSON(t *testing.T) {
	cases := []struct {
		name string
		in   Userset
		want string
	}{
		{"this", This(), `{"this":{}}`},
		{"computed", ComputedUserset("owner"), `{"computedUserset":{"relation":"owner"}}`},
		{"tupleTo", TupleTo("parent", "viewer"), `{"tupleToUserset":{"tupleset":{"relation":"parent"},"computedUserset":{"relation":"viewer"}}}`},
		{"union", Union(This(), ComputedUserset("editor")), `{"union":{"child":[{"this":{}},{"computedUserset":{"relation":"editor"}}]}}`},
		{"intersection", Intersection(This(), ComputedUserset("editor")), `{"intersection":{"child":[{"this":{}},{"computedUserset":{"relation":"editor"}}]}}`},
		{"exclusion", Exclusion(ComputedUserset("editor"), ComputedUserset("owner")), `{"difference":{"base":{"computedUserset":{"relation":"editor"}},"subtract":{"computedUserset":{"relation":"owner"}}}}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.in)
			if err != nil {
				t.Fatal(err)
			}
			if string(b) != tc.want {
				t.Errorf("json = %s, want %s", b, tc.want)
			}
		})
	}
}

func TestRelationReference_JSON(t *testing.T) {
	cases := []struct {
		name string
		in   RelationReference
		want string
	}{
		{"direct", DirectType("user"), `{"type":"user"}`},
		{"userset", RelationReference{Type: "group", Relation: "member"}, `{"type":"group","relation":"member"}`},
		{"wildcard", RelationReference{Type: "user", Wildcard: &Wildcard{}}, `{"type":"user","wildcard":{}}`},
		{"conditioned", RelationReference{Type: "user", Condition: "is_valid"}, `{"type":"user","condition":"is_valid"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.in)
			if err != nil {
				t.Fatal(err)
			}
			if string(b) != tc.want {
				t.Errorf("json = %s, want %s", b, tc.want)
			}
		})
	}
}
