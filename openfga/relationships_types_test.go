package openfga

import (
	"encoding/json"
	"testing"
)

func TestFGAObjectRelation_UnmarshalJSON(t *testing.T) {
	cases := []struct {
		name         string
		in           string
		wantObject   string
		wantRelation string
		wantErr      bool
	}{
		{"bare string", `"document:1"`, "document:1", "", false},
		{"structured with id", `{"type":"document","id":"1","relation":"viewer"}`, "document:1", "viewer", false},
		{"structured type only", `{"type":"document"}`, "document", "", false},
		{"malformed", `{"type":`, "", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var o FGAObjectRelation
			err := json.Unmarshal([]byte(tc.in), &o)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q", tc.in)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if o.Object != tc.wantObject || o.Relation != tc.wantRelation {
				t.Errorf("got {Object:%q Relation:%q}, want {%q %q}", o.Object, o.Relation, tc.wantObject, tc.wantRelation)
			}
		})
	}
}
