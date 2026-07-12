package openfga

import (
	"net/http"
	"testing"
	"time"
)

func TestResponse_QueryDuration(t *testing.T) {
	tests := []struct {
		name     string
		header   string // "" means header absent
		want     time.Duration
		wantOK   bool
		setEmpty bool // set the header to an empty string
	}{
		{name: "integer ms", header: "12", want: 12 * time.Millisecond, wantOK: true},
		{name: "fractional ms", header: "1.5", want: 1500 * time.Microsecond, wantOK: true},
		{name: "zero", header: "0", want: 0, wantOK: true},
		{name: "absent", header: "", want: 0, wantOK: false},
		{name: "empty value", setEmpty: true, want: 0, wantOK: false},
		{name: "non-numeric", header: "fast", want: 0, wantOK: false},
		{name: "negative", header: "-5", want: 0, wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := http.Header{}
			if tt.header != "" || tt.setEmpty {
				h.Set(queryDurationHeader, tt.header)
			}
			r := &Response{Response: &http.Response{Header: h}}

			got, ok := r.QueryDuration()
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if got != tt.want {
				t.Errorf("duration = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResponse_QueryDuration_NilSafe(t *testing.T) {
	var r *Response
	if _, ok := r.QueryDuration(); ok {
		t.Error("nil Response should report ok=false")
	}
	if _, ok := (&Response{}).QueryDuration(); ok {
		t.Error("Response with nil embedded *http.Response should report ok=false")
	}
}
