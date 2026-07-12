package openfga

// AuthorizationModel represents an OpenFGA authorization model.
type AuthorizationModel struct {
	ID              string               `json:"id"`
	SchemaVersion   string               `json:"schema_version"`
	TypeDefinitions []TypeDefinition     `json:"type_definitions"`
	Conditions      map[string]Condition `json:"conditions,omitempty"`
}

// TypeDefinition describes a single type within an authorization model. Relations
// maps each relation name to its rewrite rule (Userset); Metadata carries the
// directly-assignable user types for those relations.
type TypeDefinition struct {
	Type      string             `json:"type"`
	Relations map[string]Userset `json:"relations,omitempty"`
	Metadata  *Metadata          `json:"metadata,omitempty"`
}

// WriteAuthorizationModelRequest is the body sent to the Write method.
type WriteAuthorizationModelRequest struct {
	SchemaVersion   string               `json:"schema_version"`
	TypeDefinitions []TypeDefinition     `json:"type_definitions"`
	Conditions      map[string]Condition `json:"conditions,omitempty"`
}

// WriteAuthorizationModelResponse is returned by the Write method.
type WriteAuthorizationModelResponse struct {
	AuthorizationModelID string `json:"authorization_model_id"`
}

// ListModelsOptions controls pagination for the List method.
type ListModelsOptions struct {
	PageSize          int
	ContinuationToken string
}

// ListAuthorizationModelsResponse is a page of authorization models returned by List.
type ListAuthorizationModelsResponse struct {
	AuthorizationModels []AuthorizationModel `json:"authorization_models"`
	ContinuationToken   string               `json:"continuation_token"`
}

func (r *ListAuthorizationModelsResponse) continuationToken() string { return r.ContinuationToken }

// readAuthorizationModelResponse wraps a single model returned by Get.
type readAuthorizationModelResponse struct {
	AuthorizationModel AuthorizationModel `json:"authorization_model"`
}
