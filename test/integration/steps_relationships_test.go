package integration

import (
	"context"

	"github.com/cucumber/godog"
	"github.com/sergiught/go-openfga/openfga"
)

func registerRelationshipsSteps(sc *godog.ScenarioContext, st *suiteState) {
	sc.Step(`^I check whether "([^"]*)" has "([^"]*)" on "([^"]*)"$`, st.check)
	sc.Step(`^I check whether "([^"]*)" has "([^"]*)" on "([^"]*)" with a contextual tuple "([^"]*)" "([^"]*)" "([^"]*)"$`, st.checkWithContextualTuple)
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
