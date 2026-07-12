package integration

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/sergiught/go-openfga/openfga"
)

func registerOptionsSteps(sc *godog.ScenarioContext, st *suiteState) {
	sc.Step(`^a second store with the shared model granting "([^"]*)" "([^"]*)" on "([^"]*)"$`, st.secondStore)
	sc.Step(`^I check "([^"]*)" "([^"]*)" "([^"]*)" using the second store and model overrides$`, st.checkWithOverrides)
	sc.Step(`^I check "([^"]*)" "([^"]*)" "([^"]*)" with a higher-consistency override$`, st.checkWithConsistency)
	sc.Step(`^I check "([^"]*)" "([^"]*)" "([^"]*)" with a custom header$`, st.checkWithHeader)
	sc.Step(`^I read tuples with a client that has no store ID$`, st.readTuplesUnbound)
	sc.Step(`^the call fails because no store ID is set$`, st.failsNoStore)
}

// secondStore provisions an independent store+model and seeds one grant. Setup step.
func (st *suiteState) secondStore(ctx context.Context, user, relation, object string) error {
	store, err := st.client.Stores.Create(ctx, &openfga.CreateStoreRequest{Name: "second"})
	if err != nil {
		return err
	}
	st.secondStoreID = store.ID
	sc := mustWithStore(st.baseURL, store.ID)
	wm, err := sc.AuthorizationModels.Write(ctx, sharedModel())
	if err != nil {
		return err
	}
	st.secondModelID = wm.AuthorizationModelID
	err = sc.Tuples.Write(ctx, &openfga.WriteRequest{
		AuthorizationModelID: st.secondModelID,
		Writes: &openfga.WriteRequestTuples{TupleKeys: []openfga.TupleKey{
			{User: user, Relation: relation, Object: object},
		}},
	})
	return err
}

// checkWithOverrides uses the Background client (bound to the first store) but
// targets the second store and model via per-request options.
func (st *suiteState) checkWithOverrides(ctx context.Context, user, relation, object string) error {
	out, err := st.client.Relationships.Check(ctx, &openfga.CheckRequest{
		TupleKey: openfga.CheckRequestTupleKey{User: user, Relation: relation, Object: object},
	}, openfga.WithStore(st.secondStoreID), openfga.WithAuthorizationModel(st.secondModelID))
	st.lastErr = err
	st.allowed = out.Allowed
	return nil
}

func (st *suiteState) checkWithConsistency(ctx context.Context, user, relation, object string) error {
	out, err := st.client.Relationships.Check(ctx, &openfga.CheckRequest{
		AuthorizationModelID: st.modelID,
		TupleKey:             openfga.CheckRequestTupleKey{User: user, Relation: relation, Object: object},
	}, openfga.WithConsistency(openfga.ConsistencyHigherConsistency))
	st.lastErr = err
	st.allowed = out.Allowed
	return nil
}

func (st *suiteState) checkWithHeader(ctx context.Context, user, relation, object string) error {
	out, err := st.client.Relationships.Check(ctx, &openfga.CheckRequest{
		AuthorizationModelID: st.modelID,
		TupleKey:             openfga.CheckRequestTupleKey{User: user, Relation: relation, Object: object},
	}, openfga.WithRequestHeader("X-Custom-Test", "integration"))
	st.lastErr = err
	st.allowed = out.Allowed
	return nil
}

// readTuplesUnbound calls a store-scoped op on a client with no store ID. The SDK
// returns an error before making any HTTP request.
func (st *suiteState) readTuplesUnbound(ctx context.Context) error {
	c, err := openfga.NewClient(st.baseURL)
	if err != nil {
		return err
	}
	_, st.lastErr = c.Tuples.Read(ctx, &openfga.ReadRequest{})
	return nil
}

func (st *suiteState) failsNoStore() error {
	if st.lastErr == nil {
		return fmt.Errorf("expected a no-store-ID error, got nil")
	}
	if !strings.Contains(st.lastErr.Error(), "no store ID") {
		return fmt.Errorf("expected a no-store-ID error, got: %w", st.lastErr)
	}
	return nil
}
