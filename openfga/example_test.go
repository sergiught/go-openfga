package openfga_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/sergiught/go-openfga/openfga"
)

// ExampleNewClient shows how to construct a Client with a store ID and API token.
// No Output comment is present, so this is a compile-only example.
func ExampleNewClient() {
	client, err := openfga.NewClient(
		"https://api.fga.example",
		openfga.WithStoreID("01ARZ3NDEKTSV4RRFFQ69G5FAV"),
		openfga.WithAPIToken("my-api-token"),
	)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(client.BaseURL())
}

// ExampleRelationshipsService_Check shows how to call the Check API.
// No Output comment is present, so this is a compile-only example.
func ExampleRelationshipsService_Check() {
	client, err := openfga.NewClient(
		"https://api.fga.example",
		openfga.WithStoreID("01ARZ3NDEKTSV4RRFFQ69G5FAV"),
		openfga.WithAPIToken("my-api-token"),
	)
	if err != nil {
		panic(err)
	}

	resp, err := client.Relationships.Check(context.Background(), &openfga.CheckRequest{
		TupleKey: openfga.CheckRequestTupleKey{
			User:     "user:anne",
			Relation: "reader",
			Object:   "document:budget",
		},
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("allowed:", resp.Allowed)
}

// ExampleClient_NewRequest demonstrates the arbitrary-call escape hatch for
// endpoints not yet covered by the typed service methods.
// No Output comment is present, so this is a compile-only example.
func ExampleClient_NewRequest() {
	client, err := openfga.NewClient(
		"https://api.fga.example",
		openfga.WithStoreID("01ARZ3NDEKTSV4RRFFQ69G5FAV"),
		openfga.WithAPIToken("my-api-token"),
	)
	if err != nil {
		panic(err)
	}

	req, err := client.NewRequest(context.Background(), http.MethodGet, "/stores", nil)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("method:", req.Method)
}

// ExampleRelationshipsService_Allowed shows the shortcut for the most common
// query: does a user have a relation on an object?
func ExampleRelationshipsService_Allowed() {
	client, err := openfga.NewClient(
		"https://api.fga.example",
		openfga.WithStoreID("01ARZ3NDEKTSV4RRFFQ69G5FAV"),
	)
	if err != nil {
		panic(err)
	}

	ok, err := client.Relationships.Allowed(
		context.Background(), "user:anne", "reader", "document:budget")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("allowed:", ok)
}

// ExampleStoresService_All ranges over every store, paging transparently. The
// second loop value is an error that must be checked before using the store.
func ExampleStoresService_All() {
	client, err := openfga.NewClient("https://api.fga.example")
	if err != nil {
		panic(err)
	}

	for store, err := range client.Stores.All(context.Background(), nil) {
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		fmt.Println(store.Name)
	}
}

// ExampleTuplesService_WriteTuples writes an arbitrarily large slice of tuples
// as parallel non-transactional chunks, then inspects the per-tuple outcome.
func ExampleTuplesService_WriteTuples() {
	client, err := openfga.NewClient(
		"https://api.fga.example",
		openfga.WithStoreID("01ARZ3NDEKTSV4RRFFQ69G5FAV"),
	)
	if err != nil {
		panic(err)
	}

	keys := []openfga.TupleKey{
		openfga.NewTupleKey("user:anne", "reader", "document:budget"),
		openfga.NewTupleKey("user:bob", "editor", "document:roadmap"),
	}

	resp, err := client.Tuples.WriteTuples(context.Background(), keys)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	for _, r := range resp.Writes {
		if r.Err != nil {
			fmt.Printf("failed %s: %v\n", r.TupleKey.Object, r.Err)
		}
	}
}

// Example_errorHandling matches the typed errors the client returns. All embed
// *ErrorResponse, so errors.As reaches both the specific type and the base.
func Example_errorHandling() {
	client, err := openfga.NewClient(
		"https://api.fga.example",
		openfga.WithStoreID("01ARZ3NDEKTSV4RRFFQ69G5FAV"),
	)
	if err != nil {
		panic(err)
	}

	_, err = client.Relationships.Allowed(
		context.Background(), "user:anne", "reader", "document:budget")

	var rl *openfga.RateLimitError
	var notFound *openfga.NotFoundError
	var apiErr *openfga.ErrorResponse
	switch {
	case errors.As(err, &rl):
		fmt.Println("rate limited; retry after", rl.RetryAfter)
	case errors.As(err, &notFound):
		fmt.Println("not found:", notFound.RequestID())
	case errors.As(err, &apiErr):
		fmt.Println("api error code:", apiErr.Code)
	case err != nil:
		fmt.Println("transport error:", err)
	}
}

// ExampleAuthorizationModelsService_Write authors a model in Go with the typed
// builder helpers, writes it, and adopts the returned model ID as the client
// default. The builders map to the DSL: This is `[...]`, ComputedUserset is a
// bare relation, and Union is `or`.
func ExampleAuthorizationModelsService_Write() {
	client, err := openfga.NewClient(
		"https://api.fga.example",
		openfga.WithStoreID("01ARZ3NDEKTSV4RRFFQ69G5FAV"),
	)
	if err != nil {
		panic(err)
	}

	req := &openfga.WriteAuthorizationModelRequest{
		SchemaVersion: "1.1",
		TypeDefinitions: []openfga.TypeDefinition{
			{Type: "user"},
			{
				Type: "document",
				Relations: map[string]openfga.Userset{
					"owner":  openfga.This(),
					"editor": openfga.Union(openfga.This(), openfga.ComputedUserset("owner")),
					"viewer": openfga.Union(openfga.This(), openfga.ComputedUserset("editor")),
				},
				Metadata: &openfga.Metadata{
					Relations: map[string]openfga.RelationMetadata{
						"owner":  {DirectlyRelatedUserTypes: []openfga.RelationReference{openfga.DirectType("user")}},
						"editor": {DirectlyRelatedUserTypes: []openfga.RelationReference{openfga.DirectType("user")}},
						"viewer": {DirectlyRelatedUserTypes: []openfga.RelationReference{openfga.DirectType("user")}},
					},
				},
			},
		},
	}

	resp, err := client.AuthorizationModels.Write(context.Background(), req)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	_ = client.SetAuthorizationModelID(resp.AuthorizationModelID)
	fmt.Println("wrote model:", resp.AuthorizationModelID)
}

// ExampleAssertionsService_Write records assertions that pin the expected Check
// outcome for a model, so a model change that breaks them is caught early.
func ExampleAssertionsService_Write() {
	client, err := openfga.NewClient(
		"https://api.fga.example",
		openfga.WithStoreID("01ARZ3NDEKTSV4RRFFQ69G5FAV"),
	)
	if err != nil {
		panic(err)
	}

	modelID := "01HXAAAAAAAAAAAAAAAAAAAAAA"
	err = client.Assertions.Write(context.Background(), modelID, &openfga.WriteAssertionsRequest{
		Assertions: []openfga.Assertion{
			{
				TupleKey:    openfga.CheckRequestTupleKey{User: "user:anne", Relation: "viewer", Object: "document:budget"},
				Expectation: true,
			},
			{
				TupleKey:    openfga.CheckRequestTupleKey{User: "user:bob", Relation: "owner", Object: "document:budget"},
				Expectation: false,
			},
		},
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("assertions written")
}

// ExampleRelationshipsService_StreamedListObjects ranges over the NDJSON
// streaming endpoint, receiving objects as they arrive instead of buffering a
// whole page. The second loop value is an error that must be checked.
func ExampleRelationshipsService_StreamedListObjects() {
	client, err := openfga.NewClient(
		"https://api.fga.example",
		openfga.WithStoreID("01ARZ3NDEKTSV4RRFFQ69G5FAV"),
	)
	if err != nil {
		panic(err)
	}

	req := &openfga.ListObjectsRequest{
		Type:     "document",
		Relation: "viewer",
		User:     "user:anne",
	}
	for obj, err := range client.Relationships.StreamedListObjects(context.Background(), req) {
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		fmt.Println(obj.Object)
	}
}

// ExampleNewClientFromEnv builds a client from FGA_* environment variables,
// with explicit options overriding the environment.
func ExampleNewClientFromEnv() {
	client, err := openfga.NewClientFromEnv(
		openfga.WithUserAgent("my-app/1.0"),
	)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(client.BaseURL())
}
