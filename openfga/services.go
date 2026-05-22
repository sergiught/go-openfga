package openfga

import (
	"context"
	"net/http"
)

type StoresService service
type AuthorizationModelsService service
type TuplesService service
type RelationshipsService service
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
