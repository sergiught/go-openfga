package integration

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	"github.com/sergiught/go-openfga/openfga"
)

func registerAssertionsSteps(sc *godog.ScenarioContext, st *suiteState) {
	sc.Step(`^I write an assertion that "([^"]*)" "([^"]*)" "([^"]*)" is expected to be (true|false)$`, st.writeAssertion)
	sc.Step(`^I read the assertions back$`, st.readAssertions)
	sc.Step(`^the assertions include "([^"]*)" "([^"]*)" "([^"]*)" expected (true|false)$`, st.assertionsInclude)
}

func (st *suiteState) writeAssertion(ctx context.Context, user, relation, object, expect string) error {
	_, err := st.client.Assertions.Write(ctx, st.modelID, &openfga.WriteAssertionsRequest{
		Assertions: []openfga.Assertion{{
			TupleKey:    openfga.CheckRequestTupleKey{User: user, Relation: relation, Object: object},
			Expectation: expect == "true",
		}},
	})
	st.lastErr = err
	return nil
}

func (st *suiteState) readAssertions(ctx context.Context) error {
	if st.lastErr != nil {
		return fmt.Errorf("write assertions errored: %w", st.lastErr)
	}
	out, _, err := st.client.Assertions.Read(ctx, st.modelID)
	if err != nil {
		return err
	}
	st.assertions = out.Assertions
	return nil
}

func (st *suiteState) assertionsInclude(user, relation, object, expect string) error {
	want := expect == "true"
	for _, a := range st.assertions {
		if a.TupleKey.User == user && a.TupleKey.Relation == relation &&
			a.TupleKey.Object == object && a.Expectation == want {
			return nil
		}
	}
	return fmt.Errorf("assertions %v do not include %s/%s/%s expected %v", st.assertions, user, relation, object, want)
}
