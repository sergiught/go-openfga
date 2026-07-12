package openfga

import "time"

// Store is an OpenFGA store.
type Store struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// CreateStoreRequest is the body for Create.
type CreateStoreRequest struct {
	Name string `json:"name"`
}

// ListStoresOptions controls List pagination and filtering.
type ListStoresOptions struct {
	PageSize          int
	ContinuationToken string
	Name              string // filter to stores with this exact name; optional
}

// ListStoresResponse is the List result page.
type ListStoresResponse struct {
	Stores            []Store `json:"stores"`
	ContinuationToken string  `json:"continuation_token"`
}

func (r *ListStoresResponse) continuationToken() string { return r.ContinuationToken }
