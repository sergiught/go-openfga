package openfga

import (
	"encoding/json"
	"strings"
)

// ContextualTupleKeys is a set of tuple keys provided as context for a query.
// These tuples are treated as if they were already written to the store for
// the duration of the request.
type ContextualTupleKeys struct {
	TupleKeys []TupleKey `json:"tuple_keys"`
}

// CheckRequestTupleKey identifies the relationship to check. Unlike TupleKey,
// it carries no condition field. It is also reused by AssertionsService.
type CheckRequestTupleKey struct {
	User     string `json:"user"`
	Relation string `json:"relation"`
	Object   string `json:"object"`
}

// CheckRequest is the body for Relationships.Check.
type CheckRequest struct {
	TupleKey             CheckRequestTupleKey  `json:"tuple_key"`
	ContextualTuples     *ContextualTupleKeys  `json:"contextual_tuples,omitempty"`
	AuthorizationModelID string                `json:"authorization_model_id,omitempty"`
	Context              map[string]any        `json:"context,omitempty"`
	Consistency          ConsistencyPreference `json:"consistency,omitempty"`
}

// CheckResponse is returned by Relationships.Check.
type CheckResponse struct {
	Allowed    bool   `json:"allowed"`
	Resolution string `json:"resolution,omitempty"`
}

// BatchCheckItem is a single check within a BatchCheckRequest.
type BatchCheckItem struct {
	TupleKey         CheckRequestTupleKey `json:"tuple_key"`
	ContextualTuples *ContextualTupleKeys `json:"contextual_tuples,omitempty"`
	Context          map[string]any       `json:"context,omitempty"`
	CorrelationID    string               `json:"correlation_id"`
}

// BatchCheckRequest is the body for Relationships.BatchCheck.
type BatchCheckRequest struct {
	Checks               []BatchCheckItem      `json:"checks"`
	AuthorizationModelID string                `json:"authorization_model_id,omitempty"`
	Consistency          ConsistencyPreference `json:"consistency,omitempty"`
}

// BatchCheckSingleResult holds the outcome of one check within a batch.
type BatchCheckSingleResult struct {
	Allowed bool           `json:"allowed"`
	Error   map[string]any `json:"error,omitempty"`
}

// BatchCheckResponse is returned by Relationships.BatchCheck. Results are keyed
// by the CorrelationID supplied in each BatchCheckItem.
type BatchCheckResponse struct {
	Result map[string]BatchCheckSingleResult `json:"result"`
}

// ExpandRequest is the body for Relationships.Expand.
type ExpandRequest struct {
	TupleKey             CheckRequestTupleKey  `json:"tuple_key"`
	AuthorizationModelID string                `json:"authorization_model_id,omitempty"`
	Consistency          ConsistencyPreference `json:"consistency,omitempty"`
}

// ExpandResponse is returned by Relationships.Expand. The tree is returned as
// an untyped map to accommodate the recursive, schema-version-dependent shape.
type ExpandResponse struct {
	Tree map[string]any `json:"tree"`
}

// ListObjectsRequest is the body for Relationships.ListObjects.
type ListObjectsRequest struct {
	Type                 string                `json:"type"`
	Relation             string                `json:"relation"`
	User                 string                `json:"user"`
	ContextualTuples     *ContextualTupleKeys  `json:"contextual_tuples,omitempty"`
	AuthorizationModelID string                `json:"authorization_model_id,omitempty"`
	Context              map[string]any        `json:"context,omitempty"`
	Consistency          ConsistencyPreference `json:"consistency,omitempty"`
}

// ListObjectsResponse is returned by Relationships.ListObjects.
type ListObjectsResponse struct {
	Objects []string `json:"objects"`
}

// FGAObjectRelation identifies an object and an optional relation, used as the
// target in ListUsersRequest. The Object is given in the convenient "type:id"
// string form; it is serialized to OpenFGA's structured {type, id} object on
// the wire (see MarshalJSON).
type FGAObjectRelation struct {
	Object   string `json:"object,omitempty"`
	Relation string `json:"relation,omitempty"`
}

// MarshalJSON encodes the object in the structure OpenFGA's ListUsers endpoint
// expects: a nested {"type": ..., "id": ...} object split from the "type:id"
// string form. The optional relation is included only when set.
func (o FGAObjectRelation) MarshalJSON() ([]byte, error) {
	typ, id := o.Object, ""
	if i := strings.IndexByte(o.Object, ':'); i >= 0 {
		typ, id = o.Object[:i], o.Object[i+1:]
	}
	obj := struct {
		Type string `json:"type,omitempty"`
		ID   string `json:"id,omitempty"`
	}{Type: typ, ID: id}

	if o.Relation == "" {
		return json.Marshal(obj)
	}
	return json.Marshal(struct {
		Type     string `json:"type,omitempty"`
		ID       string `json:"id,omitempty"`
		Relation string `json:"relation,omitempty"`
	}{Type: typ, ID: id, Relation: o.Relation})
}

// UnmarshalJSON is the inverse of MarshalJSON: it accepts the structured
// {type, id, relation} object and rebuilds the "type:id" string form. A bare
// JSON string is also accepted for backward compatibility.
func (o *FGAObjectRelation) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		o.Object = s
		return nil
	}
	var obj struct {
		Type     string `json:"type"`
		ID       string `json:"id"`
		Relation string `json:"relation"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	if obj.ID != "" {
		o.Object = obj.Type + ":" + obj.ID
	} else {
		o.Object = obj.Type
	}
	o.Relation = obj.Relation
	return nil
}

// UserTypeFilter limits ListUsers results to a specific object type and
// optional relation.
type UserTypeFilter struct {
	Type     string `json:"type"`
	Relation string `json:"relation,omitempty"`
}

// ListUsersRequest is the body for Relationships.ListUsers.
type ListUsersRequest struct {
	Object               FGAObjectRelation     `json:"object"`
	Relation             string                `json:"relation"`
	UserFilters          []UserTypeFilter      `json:"user_filters"`
	ContextualTuples     *ContextualTupleKeys  `json:"contextual_tuples,omitempty"`
	AuthorizationModelID string                `json:"authorization_model_id,omitempty"`
	Context              map[string]any        `json:"context,omitempty"`
	Consistency          ConsistencyPreference `json:"consistency,omitempty"`
}

// ListUsersResponse is returned by Relationships.ListUsers.
type ListUsersResponse struct {
	Users []map[string]any `json:"users"`
}

// StreamedListObjectsResponse is one NDJSON line of the streaming response.
// The server wraps each item under "result".
type StreamedListObjectsResponse struct {
	Object string `json:"object"`
}

type streamedEnvelope struct {
	Result StreamedListObjectsResponse `json:"result"`
}
