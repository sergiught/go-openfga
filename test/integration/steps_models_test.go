package integration

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	"github.com/sergiught/go-openfga/openfga"
)

func registerModelsSteps(sc *godog.ScenarioContext, st *suiteState) {
	sc.Step(`^I can read the model back by ID$`, st.readModelBack)
	sc.Step(`^reading the latest model returns the shared model$`, st.readLatestModel)
	sc.Step(`^listing all models includes the written model$`, st.listAllModels)
	sc.Step(`^I write an authorization model with an undefined relation reference$`, st.writeInvalidModel)
}

func (st *suiteState) readModelBack(ctx context.Context) error {
	m, _, err := st.client.AuthorizationModels.Get(ctx, st.modelID)
	if err != nil {
		return err
	}
	if m.ID != st.modelID {
		return fmt.Errorf("got model %q, want %q", m.ID, st.modelID)
	}
	return nil
}

func (st *suiteState) readLatestModel(ctx context.Context) error {
	m, _, err := st.client.AuthorizationModels.ReadLatest(ctx)
	if err != nil {
		return err
	}
	if m.ID != st.modelID {
		return fmt.Errorf("latest model %q, want %q", m.ID, st.modelID)
	}
	return nil
}

func (st *suiteState) listAllModels(ctx context.Context) error {
	for m, err := range st.client.AuthorizationModels.All(ctx, &openfga.ListModelsOptions{PageSize: 1}) {
		if err != nil {
			return err
		}
		if m.ID == st.modelID {
			return nil
		}
	}
	return fmt.Errorf("All models did not include %q", st.modelID)
}

// writeInvalidModel writes a model whose viewer references a relation that does
// not exist, which the server rejects with 400. Action step.
func (st *suiteState) writeInvalidModel(ctx context.Context) error {
	_, _, err := st.client.AuthorizationModels.Write(ctx, &openfga.WriteAuthorizationModelRequest{
		SchemaVersion: "1.1",
		TypeDefinitions: []openfga.TypeDefinition{
			{Type: "user"},
			{Type: "document", Relations: map[string]any{
				"viewer": map[string]any{
					"computedUserset": map[string]any{"relation": "missing"},
				},
			}},
		},
	})
	st.lastErr = err
	return nil
}
