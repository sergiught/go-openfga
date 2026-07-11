# go-openfga

[![ci](https://github.com/sergiught/go-openfga/actions/workflows/ci.yml/badge.svg)](https://github.com/sergiught/go-openfga/actions/workflows/ci.yml)
[![codeql](https://github.com/sergiught/go-openfga/actions/workflows/codeql.yml/badge.svg)](https://github.com/sergiught/go-openfga/actions/workflows/codeql.yml)
[![codecov](https://codecov.io/gh/sergiught/go-openfga/branch/main/graph/badge.svg)](https://codecov.io/gh/sergiught/go-openfga)
[![Go Reference](https://pkg.go.dev/badge/github.com/sergiught/go-openfga/openfga.svg)](https://pkg.go.dev/github.com/sergiught/go-openfga/openfga)
[![Go Report Card](https://goreportcard.com/badge/github.com/sergiught/go-openfga)](https://goreportcard.com/report/github.com/sergiught/go-openfga)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/sergiught/go-openfga/badge)](https://securityscorecards.dev/viewer/?uri=github.com/sergiught/go-openfga)
[![Release](https://img.shields.io/github/v/release/sergiught/go-openfga?sort=semver)](https://github.com/sergiught/go-openfga/releases)
[![Go version](https://img.shields.io/github/go-mod/go-version/sergiught/go-openfga)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-fa6673.svg)](https://www.conventionalcommits.org)
[![PRs welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

A hand-crafted, idiomatic Go client for the [OpenFGA](https://openfga.dev) HTTP API,
modeled on the design quality of [`google/go-github`](https://github.com/google/go-github).

The client is auth-agnostic at its core: it owns only an `*http.Client`, and
authentication, retries, and custom headers are layered as composable
`http.RoundTripper` transports. Its consumer-facing dependency footprint is just
`golang.org/x/oauth2` and `github.com/golang-jwt/jwt/v5`.

## Features

- **Full v1 API coverage** — stores, authorization models, relationship tuples, all
  relationship queries (check, batch-check, expand, list-objects, list-users), and
  assertions.
- **Four authentication modes** — no-auth, pre-shared API token, OAuth2
  client-credentials, and private-key JWT (RFC 7523 client assertion).
- **Auto-paginating iterators** — Go 1.23 range-over-func for stores, models, tuple
  reads and changes, plus manual cursor control when you need it.
- **Streaming** — `StreamedListObjects` yields results from the NDJSON endpoint as they
  arrive.
- **Configurable retries** — exponential backoff with full jitter, on by default for
  HTTP 429, honoring `Retry-After`; 5xx is opt-in.
- **Typed errors** — `*ValidationError`, `*AuthenticationError`, `*NotFoundError`,
  `*RateLimitError`, `*InternalError`, all reachable via `errors.As`.
- **Escape hatch** — `NewRequest`/`Do` let you call any endpoint while reusing the
  configured auth and transport stack.

## Requirements

- Go 1.25 or newer.
- An OpenFGA server to talk to — see the [OpenFGA docs](https://openfga.dev/docs) to run one.

## Installation

```bash
go get github.com/sergiught/go-openfga/openfga
```

## Quickstart

```go
package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/sergiught/go-openfga/openfga"
)

func main() {
	client, err := openfga.NewClient(
		"https://api.fga.example",
		openfga.WithStoreID("01H5XGPQ5J6YBWBG4Z4BKRE7P"),
		openfga.WithAPIToken("my-api-token"),
	)
	if err != nil {
		panic(err)
	}

	resp, _, err := client.Relationships.Check(context.Background(), &openfga.CheckRequest{
		TupleKey: openfga.CheckRequestTupleKey{
			User:     "user:anne",
			Relation: "reader",
			Object:   "document:budget",
		},
	})
	var notFound *openfga.NotFoundError
	switch {
	case errors.As(err, &notFound):
		fmt.Println("store or model not found")
	case err != nil:
		panic(err)
	default:
		fmt.Println("allowed:", resp.Allowed)
	}
}
```

Every typed method returns `(result, *Response, error)` (or `(*Response, error)` for
writes), where `*Response` wraps the underlying `*http.Response` so you can inspect
status codes and headers.

## Authentication

Pass exactly one authentication option to `NewClient`. Omit them all for an
unauthenticated client.

```go
// Pre-shared API token.
openfga.WithAPIToken("my-api-token")

// OAuth2 client-credentials grant.
openfga.WithClientCredentials(openfga.ClientCredentialsConfig{
	TokenURL:     "https://issuer.example/oauth/token",
	ClientID:     "client-id",
	ClientSecret: "client-secret",
	Audience:     "https://api.fga.example",
})

// Private-key JWT (RFC 7523 client assertion).
openfga.WithPrivateKeyJWT(openfga.PrivateKeyJWTConfig{
	TokenURL:      "https://issuer.example/oauth/token",
	ClientID:      "client-id",
	Audience:      "https://issuer.example/",
	APIAudience:   "https://api.fga.example",
	SigningKey:    privateKey, // *rsa.PrivateKey or *ecdsa.PrivateKey
	SigningMethod: jwt.SigningMethodRS256,
})
```

## Pagination

Range-over-func iterators page transparently and lazily; the second loop value is an
error you must check:

```go
for store, err := range client.Stores.All(ctx, nil) {
	if err != nil {
		return err
	}
	fmt.Println(store.ID, store.Name)
}
```

For manual control, call the `List`/`Read` methods and follow `ContinuationToken`
yourself.

## Writing tuples

```go
_, err := client.Tuples.Write(ctx, &openfga.WriteRequest{
	Writes: &openfga.WriteRequestTuples{
		TupleKeys: []openfga.TupleKey{
			{User: "user:anne", Relation: "reader", Object: "document:budget"},
		},
	},
})
```

## Configuration

Client-wide options are passed to `NewClient`:

| Option | Purpose |
| --- | --- |
| `WithStoreID` / `WithAuthorizationModelID` | Defaults applied to every request. |
| `WithAPIToken` / `WithClientCredentials` / `WithPrivateKeyJWT` | Authentication. |
| `WithRetry(openfga.RetryConfig{...})` | Override retry attempts, backoff bounds, retryable statuses, jitter. |
| `WithHeaders(http.Header{...})` | Static headers on every request. |
| `WithUserAgent` / `WithBaseURL` | Override the User-Agent or base URL. |
| `WithHTTPClient` | Supply your own `*http.Client` (disables the built-in transport chain). |

Per-call options override client defaults for a single request:
`WithStore`, `WithAuthorizationModel`, `WithConsistency`, and `WithRequestHeader`.

## Documentation

Full API documentation, with runnable examples for the major entry points, lives on
[pkg.go.dev](https://pkg.go.dev/github.com/sergiught/go-openfga/openfga).

## Contributing

Contributions are welcome — see [CONTRIBUTING.md](CONTRIBUTING.md) for the development
workflow, and the [Code of Conduct](CODE_OF_CONDUCT.md). To report a security issue,
follow the [security policy](SECURITY.md).

## License

[MIT](LICENSE) © 2024-2026 Sergiu Ghitea.

This project is an independent client and is not affiliated with or endorsed by the
OpenFGA project or the CNCF.
