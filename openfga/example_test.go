package openfga_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sergiught/go-openfga/openfga"
)

// ExampleNewClient shows how to construct a Client with a store ID and API token.
// No Output comment is present, so this is a compile-only example.
func ExampleNewClient() {
	client, err := openfga.NewClient(
		"https://api.fga.example",
		openfga.WithStoreID("01H5XGPQ5J6YBWBG4Z4BKRE7P"),
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
		openfga.WithStoreID("01H5XGPQ5J6YBWBG4Z4BKRE7P"),
		openfga.WithAPIToken("my-api-token"),
	)
	if err != nil {
		panic(err)
	}

	resp, _, err := client.Relationships.Check(context.Background(), &openfga.CheckRequest{
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
		openfga.WithStoreID("01H5XGPQ5J6YBWBG4Z4BKRE7P"),
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
