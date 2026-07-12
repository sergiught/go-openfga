package integration

import (
	"context"
	"fmt"
	"strings"

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
	sc.Step(`^I bulk write viewer tuples for users "([^"]*)" on "([^"]*)" with chunk size (\d+)$`, st.bulkWriteViewers)
	sc.Step(`^I bulk delete viewer tuples for users "([^"]*)" on "([^"]*)" with chunk size (\d+)$`, st.bulkDeleteViewers)
	sc.Step(`^all bulk writes succeeded$`, st.allBulkWritesSucceeded)
	sc.Step(`^all bulk deletes succeeded$`, st.allBulkDeletesSucceeded)
}

func viewerKeys(usersCSV, object string) []openfga.TupleKey {
	users := strings.Split(usersCSV, ",")
	keys := make([]openfga.TupleKey, len(users))
	for i, u := range users {
		keys[i] = openfga.TupleKey{User: u, Relation: "viewer", Object: object}
	}
	return keys
}

func (st *suiteState) bulkWriteViewers(ctx context.Context, usersCSV, object string, chunk int) error {
	resp, err := st.client.Tuples.WriteTuples(ctx, viewerKeys(usersCSV, object),
		openfga.WithMaxPerChunk(chunk))
	st.lastErr = err
	st.writeResp = resp
	return nil
}

func (st *suiteState) bulkDeleteViewers(ctx context.Context, usersCSV, object string, chunk int) error {
	resp, err := st.client.Tuples.DeleteTuples(ctx, viewerKeys(usersCSV, object),
		openfga.WithMaxPerChunk(chunk))
	st.lastErr = err
	st.writeResp = resp
	return nil
}

func (st *suiteState) allBulkWritesSucceeded() error {
	return st.assertBulk(st.writeResultsFor("write"))
}

func (st *suiteState) allBulkDeletesSucceeded() error {
	return st.assertBulk(st.writeResultsFor("delete"))
}

func (st *suiteState) writeResultsFor(kind string) []openfga.TupleResult {
	if st.writeResp == nil {
		return nil
	}
	if kind == "delete" {
		return st.writeResp.Deletes
	}
	return st.writeResp.Writes
}

func (st *suiteState) assertBulk(results []openfga.TupleResult) error {
	if st.lastErr != nil {
		return fmt.Errorf("bulk operation failed: %w", st.lastErr)
	}
	if len(results) == 0 {
		return fmt.Errorf("expected per-tuple results, got none")
	}
	for _, r := range results {
		if r.Status != openfga.WriteStatusSuccess {
			return fmt.Errorf("tuple %+v: status %q err %v", r.TupleKey, r.Status, r.Err)
		}
	}
	return nil
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
	err := st.client.Tuples.Write(ctx, &openfga.WriteRequest{
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
