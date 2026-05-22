package integration

import (
	"context"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/sergiu/go-openfga/openfga"
)

type suiteState struct {
	client  *openfga.Client
	storeID string
	modelID string
	allowed bool
	lastErr error
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
	_ = os.Getenv // reserved for OPENFGA_IMAGE override hook
}
