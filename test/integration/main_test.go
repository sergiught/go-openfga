package integration

import (
	"context"
	"testing"

	"github.com/cucumber/godog"
	"github.com/sergiu/go-openfga/openfga"
)

type suiteState struct {
	client  *openfga.Client
	modelID string
	allowed bool
}

func TestFeatures(t *testing.T) {
	ctx := context.Background()
	baseURL, cleanup := startOpenFGA(ctx, t)
	defer cleanup()

	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			st := &suiteState{}
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				c, err := openfga.NewClient(baseURL)
				st.client = c
				return ctx, err
			})
			registerSteps(sc, st)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("non-zero godog status")
	}
}
