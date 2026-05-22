package openfga

// AuthorizationModel represents an OpenFGA authorization model.
type AuthorizationModel struct {
	ID              string           `json:"id"`
	SchemaVersion   string           `json:"schema_version"`
	TypeDefinitions []TypeDefinition `json:"type_definitions"`
	Conditions      map[string]any   `json:"conditions,omitempty"`
}

// TypeDefinition describes a single type within an authorization model.
type TypeDefinition struct {
	Type      string         `json:"type"`
	Relations map[string]any `json:"relations,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// WriteAuthorizationModelRequest is the body sent to the Write method.
type WriteAuthorizationModelRequest struct {
	SchemaVersion   string           `json:"schema_version"`
	TypeDefinitions []TypeDefinition `json:"type_definitions"`
	Conditions      map[string]any   `json:"conditions,omitempty"`
}

// WriteAuthorizationModelResponse is returned by the Write method.
type WriteAuthorizationModelResponse struct {
	AuthorizationModelID string `json:"authorization_model_id"`
}

// ReadModelsOptions controls pagination for the List method.
type ReadModelsOptions struct {
	PageSize          int
	ContinuationToken string
}

// ReadAuthorizationModelsResponse is a page of authorization models returned by List.
type ReadAuthorizationModelsResponse struct {
	AuthorizationModels []AuthorizationModel `json:"authorization_models"`
	ContinuationToken   string               `json:"continuation_token"`
}

func (r *ReadAuthorizationModelsResponse) continuationToken() string { return r.ContinuationToken }

// ReadAuthorizationModelResponse wraps a single model returned by Get.
type ReadAuthorizationModelResponse struct {
	AuthorizationModel AuthorizationModel `json:"authorization_model"`
}
