// Package openfga is an idiomatic Go client for the OpenFGA HTTP API.
//
// # Construction
//
// Build a client with NewClient and one or more Option values:
//
//	c, err := openfga.NewClient("https://api.fga.example",
//		openfga.WithStoreID("01H..."),
//		openfga.WithAPIToken("secret"))
//
// NewClient never reads the environment. To configure from FGA_* variables,
// use NewClientFromEnv (or EnvOptions to merge env with explicit options).
//
// # Services
//
// API calls are grouped into service handles on the Client:
//
//	c.Stores               // create/list/get/delete stores
//	c.AuthorizationModels  // write/read authorization models
//	c.Tuples               // write/read relationship tuples and changes
//	c.Relationships        // check, batch-check, expand, list-objects, list-users
//	c.Assertions           // read/write test assertions
//
// Methods return (result, error), or just error for write-only calls. To reach
// the raw HTTP response (status, headers, request ID), pass the OnResponse
// request option, which receives the *Response after the body is decoded. For
// cross-cutting observation of every request, use WithRequestObserver; for full
// manual control, use NewRequest and Do.
//
// # Common calls
//
// Relationships.Allowed is the shortcut for the most common query:
//
//	ok, err := c.Relationships.Allowed(ctx, "user:anne", "reader", "document:budget")
//
// NewTupleKey and NewCheckRequest build the request structs with less ceremony.
//
// # Authentication
//
// Pass exactly one authentication option, or none for an unauthenticated
// client: WithAPIToken, WithClientCredentials, WithPrivateKeyJWT, or
// WithTokenSource (any oauth2.TokenSource, e.g. Vault or workload identity).
//
// # Pagination
//
// Range-over-func iterators page transparently; the second loop value is an
// error you must check:
//
//	for store, err := range c.Stores.All(ctx, nil) {
//		if err != nil { return err }
//		// use store
//	}
//
// Stores.All, AuthorizationModels.All, Tuples.ReadAll, and Tuples.ChangesAll
// follow this shape; the underlying List/Read methods expose manual cursors.
//
// # Errors
//
// Non-2xx responses become typed errors reachable with errors.As:
// *ValidationError (400), *AuthenticationError (401/403), *NotFoundError (404),
// *RateLimitError (429, with RetryAfter), and *InternalError (5xx). All embed
// *ErrorResponse, whose Code holds the OpenFGA error code (see the Code*
// constants) and whose RequestID reports the server correlation ID.
//
// # Options
//
// Client-wide options and per-call RequestOptions overlap by design; per-call
// wins. The pairs are WithStoreID/WithStore,
// WithAuthorizationModelID/WithAuthorizationModel, and
// WithDefaultConsistency/WithConsistency.
//
// # Transport and extensibility
//
// The client owns only an *http.Client; authentication, retries, and static
// headers are layered as composable http.RoundTripper transports. Add tracing,
// metrics, or a custom dialer beneath that chain with WithBaseTransport (it also
// carries out-of-band token fetches), observe each attempt with
// WithRequestObserver, or replace the whole stack with WithHTTPClient.
package openfga
