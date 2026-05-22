package openfga

// Assertion is a single test case for an authorization model: a tuple key, the
// expected Check outcome, and optional contextual tuples or condition context.
type Assertion struct {
	TupleKey         CheckRequestTupleKey `json:"tuple_key"`
	Expectation      bool                 `json:"expectation"`
	ContextualTuples []TupleKey           `json:"contextual_tuples,omitempty"`
	Context          map[string]any       `json:"context,omitempty"`
}

// WriteAssertionsRequest is the body for AssertionsService.Write.
type WriteAssertionsRequest struct {
	Assertions []Assertion `json:"assertions"`
}

// ReadAssertionsResponse is returned by AssertionsService.Read.
type ReadAssertionsResponse struct {
	AuthorizationModelID string      `json:"authorization_model_id"`
	Assertions           []Assertion `json:"assertions"`
}
