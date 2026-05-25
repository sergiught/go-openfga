package openfga

import (
	"context"
	"net/http"
)

// StoresService groups the store lifecycle endpoints (create/list/get/delete).
type StoresService service

// AuthorizationModelsService groups the authorization-model endpoints
// (write/list/get/read-latest).
type AuthorizationModelsService service

// TuplesService groups the relationship-tuple endpoints (write/read/changes).
type TuplesService service

// RelationshipsService groups the query endpoints
// (check/batch-check/expand/list-objects/list-users).
type RelationshipsService service

// AssertionsService groups the assertion endpoints (write/read).
type AssertionsService service

// doStorePost is the common path for store-scoped POST query endpoints.
func (c *Client) doStorePost(ctx context.Context, suffix string, body any, out any, opts []RequestOption) (*Response, error) {
	rc := newRequestConfig()
	applyOptions(rc, opts)
	store, err := c.storeFor(rc)
	if err != nil {
		return nil, err
	}
	req, err := c.newRequest(ctx, http.MethodPost, "/stores/"+store+suffix, body, rc.header)
	if err != nil {
		return nil, err
	}
	return c.Do(req, out)
}
