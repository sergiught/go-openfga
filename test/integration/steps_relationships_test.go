package integration

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	"github.com/sergiught/go-openfga/openfga"
)

func registerRelationshipsSteps(sc *godog.ScenarioContext, st *suiteState) {
	sc.Step(`^I check whether "([^"]*)" has "([^"]*)" on "([^"]*)"$`, st.check)
	sc.Step(`^I check whether "([^"]*)" has "([^"]*)" on "([^"]*)" with a contextual tuple "([^"]*)" "([^"]*)" "([^"]*)"$`, st.checkWithContextualTuple)
	sc.Step(`^a batch item "([^"]*)" checking "([^"]*)" has "([^"]*)" on "([^"]*)"$`, st.addBatchItem)
	sc.Step(`^I run the batch check$`, st.runBatchCheck)
	sc.Step(`^batch item "([^"]*)" is allowed$`, st.batchAllowed)
	sc.Step(`^batch item "([^"]*)" is denied$`, st.batchDenied)
	sc.Step(`^I expand "([^"]*)" on "([^"]*)"$`, st.expandRelation)
	sc.Step(`^the expansion tree is not empty$`, st.expandNotEmpty)
	sc.Step(`^I list "([^"]*)" objects "([^"]*)" can "([^"]*)"$`, st.listObjects)
	sc.Step(`^the objects include "([^"]*)"$`, st.objectsInclude)
	sc.Step(`^the objects do not include "([^"]*)"$`, st.objectsExclude)
	sc.Step(`^I stream "([^"]*)" objects "([^"]*)" can "([^"]*)"$`, st.streamObjects)
	sc.Step(`^the streamed objects include "([^"]*)"$`, st.objectsInclude)
}

// check performs a Check. Action step: captures error into st.lastErr.
func (st *suiteState) check(ctx context.Context, user, relation, object string) error {
	out, _, err := st.client.Relationships.Check(ctx, &openfga.CheckRequest{
		AuthorizationModelID: st.modelID,
		TupleKey:             openfga.CheckRequestTupleKey{User: user, Relation: relation, Object: object},
	})
	st.lastErr = err
	st.allowed = out.Allowed
	return nil
}

func (st *suiteState) addBatchItem(correlation, user, relation, object string) error {
	st.batchItems = append(st.batchItems, openfga.BatchCheckItem{
		CorrelationID: correlation,
		TupleKey:      openfga.CheckRequestTupleKey{User: user, Relation: relation, Object: object},
	})
	return nil
}

func (st *suiteState) runBatchCheck(ctx context.Context) error {
	out, _, err := st.client.Relationships.BatchCheck(ctx, &openfga.BatchCheckRequest{
		AuthorizationModelID: st.modelID,
		Checks:               st.batchItems,
	})
	st.lastErr = err
	st.batch = out
	return nil
}

func (st *suiteState) batchAllowed(correlation string) error {
	return st.assertBatch(correlation, true)
}

func (st *suiteState) batchDenied(correlation string) error {
	return st.assertBatch(correlation, false)
}

func (st *suiteState) assertBatch(correlation string, want bool) error {
	if st.lastErr != nil {
		return fmt.Errorf("batch check errored: %w", st.lastErr)
	}
	res, ok := st.batch.Result[correlation]
	if !ok {
		return fmt.Errorf("no result for correlation %q", correlation)
	}
	if res.Allowed != want {
		return fmt.Errorf("correlation %q: want allowed=%v, got %v", correlation, want, res.Allowed)
	}
	return nil
}

func (st *suiteState) expandRelation(ctx context.Context, relation, object string) error {
	out, _, err := st.client.Relationships.Expand(ctx, &openfga.ExpandRequest{
		AuthorizationModelID: st.modelID,
		TupleKey:             openfga.CheckRequestTupleKey{Relation: relation, Object: object},
	})
	st.lastErr = err
	st.expand = out
	return nil
}

func (st *suiteState) expandNotEmpty() error {
	if st.lastErr != nil {
		return fmt.Errorf("expand errored: %w", st.lastErr)
	}
	if st.expand == nil || len(st.expand.Tree) == 0 {
		return fmt.Errorf("expected a non-empty expansion tree")
	}
	return nil
}

func (st *suiteState) listObjects(ctx context.Context, typ, user, relation string) error {
	out, _, err := st.client.Relationships.ListObjects(ctx, &openfga.ListObjectsRequest{
		AuthorizationModelID: st.modelID,
		Type:                 typ,
		Relation:             relation,
		User:                 user,
	})
	st.lastErr = err
	if out != nil {
		st.objects = out.Objects
	}
	return nil
}

func (st *suiteState) streamObjects(ctx context.Context, typ, user, relation string) error {
	st.objects = nil
	for obj, err := range st.client.Relationships.StreamedListObjects(ctx, &openfga.ListObjectsRequest{
		AuthorizationModelID: st.modelID,
		Type:                 typ,
		Relation:             relation,
		User:                 user,
	}) {
		if err != nil {
			st.lastErr = err
			return nil
		}
		st.objects = append(st.objects, obj.Object)
	}
	return nil
}

func (st *suiteState) objectsInclude(object string) error {
	if st.lastErr != nil {
		return fmt.Errorf("list errored: %w", st.lastErr)
	}
	for _, o := range st.objects {
		if o == object {
			return nil
		}
	}
	return fmt.Errorf("objects %v do not include %q", st.objects, object)
}

func (st *suiteState) objectsExclude(object string) error {
	for _, o := range st.objects {
		if o == object {
			return fmt.Errorf("objects %v unexpectedly include %q", st.objects, object)
		}
	}
	return nil
}

func (st *suiteState) checkWithContextualTuple(ctx context.Context, user, relation, object, cu, cr, co string) error {
	out, _, err := st.client.Relationships.Check(ctx, &openfga.CheckRequest{
		AuthorizationModelID: st.modelID,
		TupleKey:             openfga.CheckRequestTupleKey{User: user, Relation: relation, Object: object},
		ContextualTuples: &openfga.ContextualTupleKeys{TupleKeys: []openfga.TupleKey{
			{User: cu, Relation: cr, Object: co},
		}},
	})
	st.lastErr = err
	st.allowed = out.Allowed
	return nil
}
