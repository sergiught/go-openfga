package integration

import (
	"github.com/cucumber/godog"
	"github.com/sergiught/go-openfga/openfga"
)

// suiteState is the per-scenario state shared across every step. A fresh value
// is assigned in the Before hook (see main_test.go).
type suiteState struct {
	client  *openfga.Client
	baseURL string

	modelID     string
	storeID     string
	lastStoreID string

	secondStoreID string
	secondModelID string

	allowed bool
	lastErr error

	objects    []string
	users      []map[string]any
	tuples     []openfga.Tuple
	changes    []openfga.TupleChange
	expand     *openfga.ExpandResponse
	batch      *openfga.BatchCheckResponse
	batchItems []openfga.BatchCheckItem
	assertions []openfga.Assertion
}

func registerSteps(sc *godog.ScenarioContext, st *suiteState) {
	registerCommonSteps(sc, st)
	registerRelationshipsSteps(sc, st)
	registerStoresSteps(sc, st)
	registerModelsSteps(sc, st)
	registerTuplesSteps(sc, st)
	registerAssertionsSteps(sc, st)
	registerOptionsSteps(sc, st)
}
