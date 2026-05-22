package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// startOpenFGA launches an ephemeral OpenFGA container and returns its base URL.
func startOpenFGA(ctx context.Context, t *testing.T) (string, func()) {
	t.Helper()
	image := "openfga/openfga:latest"
	if v := os.Getenv("OPENFGA_IMAGE"); v != "" {
		image = v
	}
	req := testcontainers.ContainerRequest{
		Image:        image,
		Cmd:          []string{"run"},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForHTTP("/healthz").WithPort("8080/tcp").WithStartupTimeout(60 * time.Second),
	}
	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Skipf("cannot start OpenFGA container (Docker unavailable?): %v", err)
	}
	host, _ := ctr.Host(ctx)
	port, _ := ctr.MappedPort(ctx, "8080/tcp")
	cleanup := func() { _ = ctr.Terminate(ctx) }
	return "http://" + host + ":" + port.Port(), cleanup
}
