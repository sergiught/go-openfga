package integration

import (
	"context"
	"errors"
	"fmt"

	"github.com/cucumber/godog"
	"github.com/sergiught/go-openfga/openfga"
)

func registerCommonSteps(sc *godog.ScenarioContext, st *suiteState) {
	sc.Step(`^a fresh store with the shared model$`, st.freshStore)
	sc.Step(`^the tuple "([^"]*)" "([^"]*)" "([^"]*)" is written$`, st.writeTuple)
	sc.Step(`^the result is allowed$`, st.resultAllowed)
	sc.Step(`^the result is denied$`, st.resultDenied)
	sc.Step(`^the call fails with a validation error$`, st.failsValidation)
	sc.Step(`^the call fails with a not found error$`, st.failsNotFound)
}

// freshStore creates a new store, binds the client to it, and writes the shared
// authorization model. Setup step: returns its error to abort on failure.
func (st *suiteState) freshStore(ctx context.Context) error {
	store, _, err := st.client.Stores.Create(ctx, &openfga.CreateStoreRequest{Name: "bdd"})
	if err != nil {
		return err
	}
	st.storeID = store.ID
	st.client = mustWithStore(st.baseURL, store.ID)

	wm, _, err := st.client.AuthorizationModels.Write(ctx, sharedModel())
	if err != nil {
		return err
	}
	st.modelID = wm.AuthorizationModelID
	return nil
}

// writeTuple grants a relationship. Setup step: returns its error.
func (st *suiteState) writeTuple(ctx context.Context, user, relation, object string) error {
	_, err := st.client.Tuples.Write(ctx, &openfga.WriteRequest{
		AuthorizationModelID: st.modelID,
		Writes: &openfga.WriteRequestTuples{TupleKeys: []openfga.TupleKey{
			{User: user, Relation: relation, Object: object},
		}},
	})
	return err
}

func (st *suiteState) resultAllowed() error {
	if st.lastErr != nil {
		return fmt.Errorf("expected allowed, got error: %w", st.lastErr)
	}
	if !st.allowed {
		return fmt.Errorf("expected allowed, got denied")
	}
	return nil
}

func (st *suiteState) resultDenied() error {
	if st.lastErr != nil {
		return fmt.Errorf("expected denied, got error: %w", st.lastErr)
	}
	if st.allowed {
		return fmt.Errorf("expected denied, got allowed")
	}
	return nil
}

func (st *suiteState) failsValidation() error {
	var target *openfga.ValidationError
	if !errors.As(st.lastErr, &target) {
		return fmt.Errorf("expected a validation error, got: %v", st.lastErr)
	}
	return nil
}

func (st *suiteState) failsNotFound() error {
	var target *openfga.NotFoundError
	if !errors.As(st.lastErr, &target) {
		return fmt.Errorf("expected a not found error, got: %v", st.lastErr)
	}
	return nil
}

// mustWithStore returns a client bound to storeID with strong read consistency.
func mustWithStore(baseURL, storeID string) *openfga.Client {
	c, err := openfga.NewClient(baseURL,
		openfga.WithStoreID(storeID),
		openfga.WithDefaultConsistency(openfga.ConsistencyHigherConsistency),
	)
	if err != nil {
		panic(err)
	}
	return c
}

// sharedModel builds the document/group/user model every scenario relies on.
func sharedModel() *openfga.WriteAuthorizationModelRequest {
	return &openfga.WriteAuthorizationModelRequest{
		SchemaVersion: "1.1",
		TypeDefinitions: []openfga.TypeDefinition{
			{Type: "user"},
			{Type: "group", Relations: map[string]any{
				"member": map[string]any{"this": map[string]any{}},
			}, Metadata: map[string]any{
				"relations": map[string]any{
					"member": map[string]any{
						"directly_related_user_types": []map[string]any{{"type": "user"}},
					},
				},
			}},
			{Type: "document", Relations: map[string]any{
				"owner":  map[string]any{"this": map[string]any{}},
				"editor": map[string]any{"this": map[string]any{}},
				"viewer": map[string]any{
					"union": map[string]any{
						"child": []map[string]any{
							{"this": map[string]any{}},
							{"computedUserset": map[string]any{"relation": "editor"}},
							{"computedUserset": map[string]any{"relation": "owner"}},
						},
					},
				},
			}, Metadata: map[string]any{
				"relations": map[string]any{
					"owner": map[string]any{
						"directly_related_user_types": []map[string]any{{"type": "user"}},
					},
					"editor": map[string]any{
						"directly_related_user_types": []map[string]any{
							{"type": "user"},
							{"type": "group", "relation": "member"},
						},
					},
					"viewer": map[string]any{
						"directly_related_user_types": []map[string]any{
							{"type": "user"},
							{"type": "group", "relation": "member"},
						},
					},
				},
			}},
		},
	}
}
