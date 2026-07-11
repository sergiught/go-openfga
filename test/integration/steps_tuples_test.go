package integration

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	"github.com/sergiught/go-openfga/openfga"
)

func registerTuplesSteps(sc *godog.ScenarioContext, st *suiteState) {
	sc.Step(`^I read tuples for object "([^"]*)"$`, st.readTuples)
	sc.Step(`^the tuples include "([^"]*)" "([^"]*)" "([^"]*)"$`, st.tuplesInclude)
	sc.Step(`^the tuples do not include "([^"]*)" "([^"]*)" "([^"]*)"$`, st.tuplesExclude)
	sc.Step(`^I delete the tuple "([^"]*)" "([^"]*)" "([^"]*)"$`, st.deleteTuple)
	sc.Step(`^I read all tuple changes$`, st.readChanges)
	sc.Step(`^the changes include a "([^"]*)" of "([^"]*)" "([^"]*)" "([^"]*)"$`, st.changesInclude)
	sc.Step(`^I write the tuple "([^"]*)" "([^"]*)" "([^"]*)"$`, st.writeTupleAction)
}

func (st *suiteState) readTuples(ctx context.Context, object string) error {
	st.tuples = nil
	for tp, err := range st.client.Tuples.ReadAll(ctx, &openfga.ReadRequest{
		TupleKey:    &openfga.ReadRequestTupleKey{Object: object},
		Consistency: openfga.ConsistencyHigherConsistency,
	}) {
		if err != nil {
			st.lastErr = err
			return nil
		}
		st.tuples = append(st.tuples, tp)
	}
	return nil
}

func (st *suiteState) tuplesInclude(user, relation, object string) error {
	if st.lastErr != nil {
		return fmt.Errorf("read errored: %w", st.lastErr)
	}
	if st.hasTuple(user, relation, object) {
		return nil
	}
	return fmt.Errorf("tuples do not include %s/%s/%s", user, relation, object)
}

func (st *suiteState) tuplesExclude(user, relation, object string) error {
	if st.lastErr != nil {
		return fmt.Errorf("read errored: %w", st.lastErr)
	}
	if st.hasTuple(user, relation, object) {
		return fmt.Errorf("tuples unexpectedly include %s/%s/%s", user, relation, object)
	}
	return nil
}

func (st *suiteState) hasTuple(user, relation, object string) bool {
	for _, tp := range st.tuples {
		if tp.Key.User == user && tp.Key.Relation == relation && tp.Key.Object == object {
			return true
		}
	}
	return false
}

func (st *suiteState) deleteTuple(ctx context.Context, user, relation, object string) error {
	_, err := st.client.Tuples.Write(ctx, &openfga.WriteRequest{
		AuthorizationModelID: st.modelID,
		Deletes: &openfga.WriteRequestTuples{TupleKeys: []openfga.TupleKey{
			{User: user, Relation: relation, Object: object},
		}},
	})
	st.lastErr = err
	return nil
}

func (st *suiteState) readChanges(ctx context.Context) error {
	st.changes = nil
	for ch, err := range st.client.Tuples.ChangesAll(ctx, &openfga.ReadChangesOptions{}) {
		if err != nil {
			st.lastErr = err
			return nil
		}
		st.changes = append(st.changes, ch)
	}
	return nil
}

func (st *suiteState) changesInclude(op, user, relation, object string) error {
	if st.lastErr != nil {
		return fmt.Errorf("read changes errored: %w", st.lastErr)
	}
	for _, ch := range st.changes {
		if ch.Operation == op && ch.TupleKey.User == user && ch.TupleKey.Relation == relation && ch.TupleKey.Object == object {
			return nil
		}
	}
	return fmt.Errorf("changes do not include %s of %s/%s/%s", op, user, relation, object)
}

// writeTupleAction is the error-capturing variant used by the validation scenario.
func (st *suiteState) writeTupleAction(ctx context.Context, user, relation, object string) error {
	st.lastErr = st.writeTuple(ctx, user, relation, object)
	return nil
}
