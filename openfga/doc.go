// Package openfga is an idiomatic Go client for the OpenFGA HTTP API.
//
// Construct a client with NewClient and one or more Option values:
//
//	c, err := openfga.NewClient("https://api.fga.example",
//		openfga.WithStoreID("01H..."),
//		openfga.WithAPIToken("secret"))
//
// The client owns only an *http.Client; authentication, retries, and custom
// headers are layered as composable http.RoundTripper transports.
package openfga
