package integration

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	"github.com/sergiu/go-openfga/openfga"
)

func registerSteps(sc *godog.ScenarioContext, st *suiteState) {
	sc.Step(`^a fresh store with a document authorization model$`, st.freshStoreWithModel)
	sc.Step(`^the tuple "([^"]*)" "([^"]*)" "([^"]*)" is written$`, st.writeTuple)
	sc.Step(`^I check whether "([^"]*)" has "([^"]*)" on "([^"]*)"$`, st.check)
	sc.Step(`^the result is allowed$`, st.resultAllowed)
	sc.Step(`^the result is denied$`, st.resultDenied)
}

func (st *suiteState) freshStoreWithModel(ctx context.Context) error {
	store, _, err := st.client.Stores.Create(ctx, &openfga.CreateStoreRequest{Name: "bdd"})
	if err != nil {
		return err
	}
	st.client = mustWithStore(st.client, store.ID)

	model := &openfga.WriteAuthorizationModelRequest{
		SchemaVersion: "1.1",
		TypeDefinitions: []openfga.TypeDefinition{
			{Type: "user"},
			{Type: "document", Relations: map[string]any{
				"reader": map[string]any{"this": map[string]any{}},
			}, Metadata: map[string]any{
				"relations": map[string]any{
					"reader": map[string]any{
						"directly_related_user_types": []map[string]any{{"type": "user"}},
					},
				},
			}},
		},
	}
	wm, _, err := st.client.AuthorizationModels.Write(ctx, model)
	if err != nil {
		return err
	}
	st.modelID = wm.AuthorizationModelID
	return nil
}

func (st *suiteState) writeTuple(ctx context.Context, user, relation, object string) error {
	_, err := st.client.Tuples.Write(ctx, &openfga.WriteRequest{
		AuthorizationModelID: st.modelID,
		Writes: &openfga.WriteRequestTuples{TupleKeys: []openfga.TupleKey{
			{User: user, Relation: relation, Object: object},
		}},
	})
	return err
}

func (st *suiteState) check(ctx context.Context, user, relation, object string) error {
	out, _, err := st.client.Relationships.Check(ctx, &openfga.CheckRequest{
		AuthorizationModelID: st.modelID,
		TupleKey:             openfga.CheckRequestTupleKey{User: user, Relation: relation, Object: object},
	})
	if err != nil {
		return err
	}
	st.allowed = out.Allowed
	return nil
}

func (st *suiteState) resultAllowed() error {
	if !st.allowed {
		return fmt.Errorf("expected allowed, got denied")
	}
	return nil
}

func (st *suiteState) resultDenied() error {
	if st.allowed {
		return fmt.Errorf("expected denied, got allowed")
	}
	return nil
}

// mustWithStore returns a client bound to storeID (constructed against the same base URL).
func mustWithStore(c *openfga.Client, storeID string) *openfga.Client {
	nc, err := openfga.NewClient(c.BaseURL(), openfga.WithStoreID(storeID))
	if err != nil {
		panic(err)
	}
	return nc
}
