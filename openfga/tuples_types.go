package openfga

import "time"

// RelationshipCondition is an optional ABAC condition attached to a tuple.
type RelationshipCondition struct {
	Name    string         `json:"name"`
	Context map[string]any `json:"context,omitempty"`
}

// TupleKey identifies a relationship triple, with an optional condition.
type TupleKey struct {
	User      string                 `json:"user"`
	Relation  string                 `json:"relation"`
	Object    string                 `json:"object"`
	Condition *RelationshipCondition `json:"condition,omitempty"`
}

// Tuple is a stored relationship triple with a server-assigned timestamp.
type Tuple struct {
	Key       TupleKey  `json:"key"`
	Timestamp time.Time `json:"timestamp"`
}

// WriteRequest is the body for Tuples.Write.
type WriteRequest struct {
	Writes               *WriteRequestTuples `json:"writes,omitempty"`
	Deletes              *WriteRequestTuples `json:"deletes,omitempty"`
	AuthorizationModelID string              `json:"authorization_model_id,omitempty"`
}

// OnDuplicate controls how the server handles a write whose tuple already
// exists. Requires OpenFGA >= 1.10. Empty means the server default ("error").
type OnDuplicate string

const (
	OnDuplicateError  OnDuplicate = "error"
	OnDuplicateIgnore OnDuplicate = "ignore"
)

// OnMissing controls how the server handles a delete whose tuple does not
// exist. Requires OpenFGA >= 1.10. Empty means the server default ("error").
type OnMissing string

const (
	OnMissingError  OnMissing = "error"
	OnMissingIgnore OnMissing = "ignore"
)

// WriteRequestTuples carries a list of tuple keys for a write or delete
// operation. OnDuplicate is only meaningful on the Writes block; OnMissing is
// only meaningful on the Deletes block.
type WriteRequestTuples struct {
	TupleKeys   []TupleKey  `json:"tuple_keys"`
	OnDuplicate OnDuplicate `json:"on_duplicate,omitempty"`
	OnMissing   OnMissing   `json:"on_missing,omitempty"`
}

// ReadRequestTupleKey is a partial tuple key used as a filter in Read requests.
// All fields are optional; omit a field to match any value.
type ReadRequestTupleKey struct {
	User     string `json:"user,omitempty"`
	Relation string `json:"relation,omitempty"`
	Object   string `json:"object,omitempty"`
}

// ReadRequest is the body for Tuples.Read.
type ReadRequest struct {
	TupleKey          *ReadRequestTupleKey  `json:"tuple_key,omitempty"`
	PageSize          int                   `json:"page_size,omitempty"`
	ContinuationToken string                `json:"continuation_token,omitempty"`
	Consistency       ConsistencyPreference `json:"consistency,omitempty"`
}

// ReadResponse is the result of a Tuples.Read call.
type ReadResponse struct {
	Tuples            []Tuple `json:"tuples"`
	ContinuationToken string  `json:"continuation_token"`
}

func (r *ReadResponse) continuationToken() string { return r.ContinuationToken }

// ReadChangesOptions controls filtering and pagination for Tuples.ReadChanges.
type ReadChangesOptions struct {
	Type              string
	PageSize          int
	ContinuationToken string
	StartTime         string // RFC3339; optional
}

// TupleChange describes a single write or delete event in the changelog.
type TupleChange struct {
	TupleKey  TupleKey  `json:"tuple_key"`
	Operation string    `json:"operation"` // TUPLE_OPERATION_WRITE | TUPLE_OPERATION_DELETE
	Timestamp time.Time `json:"timestamp"`
}

// ReadChangesResponse is the result of a Tuples.ReadChanges call.
type ReadChangesResponse struct {
	Changes           []TupleChange `json:"changes"`
	ContinuationToken string        `json:"continuation_token"`
}

func (r *ReadChangesResponse) continuationToken() string { return r.ContinuationToken }

// WriteStatus reports the outcome of a single tuple in a bulk write/delete.
type WriteStatus string

const (
	WriteStatusSuccess WriteStatus = "success"
	WriteStatusFailure WriteStatus = "failure"
)

// TupleResult is the per-tuple outcome of Tuples.WriteTuples / DeleteTuples.
// Err is non-nil exactly when Status is WriteStatusFailure.
type TupleResult struct {
	TupleKey TupleKey
	Status   WriteStatus
	Err      error
}

// WriteTuplesResponse aggregates per-tuple outcomes. WriteTuples populates
// Writes; DeleteTuples populates Deletes.
type WriteTuplesResponse struct {
	Writes  []TupleResult
	Deletes []TupleResult
}
