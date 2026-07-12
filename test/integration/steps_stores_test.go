package integration

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/sergiught/go-openfga/openfga"
)

func registerStoresSteps(sc *godog.ScenarioContext, st *suiteState) {
	sc.Step(`^I create a store named "([^"]*)"$`, st.createStore)
	sc.Step(`^I can read that store back by ID$`, st.readStoreBack)
	sc.Step(`^iterating all stores includes the background store and the created store$`, st.allStoresInclude)
	sc.Step(`^I delete that store$`, st.deleteThatStore)
	sc.Step(`^I get that store$`, st.getThatStore)
}

// createStore creates a store and records its ID. Action step.
func (st *suiteState) createStore(ctx context.Context, name string) error {
	store, err := st.client.Stores.Create(ctx, &openfga.CreateStoreRequest{Name: name})
	st.lastErr = err
	if store != nil {
		st.lastStoreID = store.ID
	}
	return nil
}

func (st *suiteState) readStoreBack(ctx context.Context) error {
	if st.lastErr != nil {
		return fmt.Errorf("create errored: %w", st.lastErr)
	}
	got, err := st.client.Stores.Get(ctx, st.lastStoreID)
	if err != nil {
		return err
	}
	if got.ID != st.lastStoreID {
		return fmt.Errorf("got store %q, want %q", got.ID, st.lastStoreID)
	}
	return nil
}

func (st *suiteState) allStoresInclude(ctx context.Context) error {
	seen := map[string]bool{}
	for store, err := range st.client.Stores.All(ctx, &openfga.ListStoresOptions{PageSize: 1}) {
		if err != nil {
			return err
		}
		seen[store.ID] = true
	}
	if !seen[st.storeID] || !seen[st.lastStoreID] {
		return fmt.Errorf("All did not include both %q and %q (saw %d stores)", st.storeID, st.lastStoreID, len(seen))
	}
	return nil
}

func (st *suiteState) deleteThatStore(ctx context.Context) error {
	err := st.client.Stores.Delete(ctx, st.lastStoreID)
	st.lastErr = err
	return nil
}

// getThatStore fetches the last-created store; captures the error for assertion.
func (st *suiteState) getThatStore(ctx context.Context) error {
	_, err := st.client.Stores.Get(ctx, st.lastStoreID)
	st.lastErr = err
	return nil
}
