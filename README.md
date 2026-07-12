<p align="center">
  <img src="docs/assets/banner.svg" alt="go-openfga — fine-grained authorization for Go, a hand-crafted OpenFGA client" width="100%">
</p>

<h1 align="center">go-openfga</h1>

<div align="center">

[![ci](https://github.com/sergiught/go-openfga/actions/workflows/ci.yml/badge.svg)](https://github.com/sergiught/go-openfga/actions/workflows/ci.yml)
[![codeql](https://github.com/sergiught/go-openfga/actions/workflows/codeql.yml/badge.svg)](https://github.com/sergiught/go-openfga/actions/workflows/codeql.yml)
[![codecov](https://codecov.io/gh/sergiught/go-openfga/branch/main/graph/badge.svg)](https://codecov.io/gh/sergiught/go-openfga)
[![Go Reference](https://pkg.go.dev/badge/github.com/sergiught/go-openfga/openfga.svg)](https://pkg.go.dev/github.com/sergiught/go-openfga/openfga)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/sergiught/go-openfga/badge)](https://securityscorecards.dev/viewer/?uri=github.com/sergiught/go-openfga)
[![Release](https://img.shields.io/github/v/release/sergiught/go-openfga?sort=semver)](https://github.com/sergiught/go-openfga/releases)
[![Go version](https://img.shields.io/github/go-mod/go-version/sergiught/go-openfga)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-fa6673.svg)](https://www.conventionalcommits.org)
[![PRs welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

[Quickstart](#-quickstart) · [Authentication](#-authentication) · [Configuration](#-configuration) · [Error handling](#-error-handling) · [API reference](https://pkg.go.dev/github.com/sergiught/go-openfga/openfga)

</div>

A hand-crafted, idiomatic Go client for the [OpenFGA](https://openfga.dev) HTTP API.

The client is auth-agnostic at its core: it owns only an `*http.Client`, and
authentication, retries, and custom headers are layered as composable
`http.RoundTripper` transports. Its consumer-facing dependency footprint is just
`golang.org/x/oauth2` and `github.com/golang-jwt/jwt/v5`.

---

## 📑 Table of contents

- [✨ Features](#-features)
- [📋 Requirements](#-requirements)
- [📦 Installation](#-installation)
- [🚀 Quickstart](#-quickstart)
- [🔐 Authentication](#-authentication)
- [🌱 Configuration from the environment](#-configuration-from-the-environment)
- [📄 Pagination](#-pagination)
- [📝 Writing tuples](#-writing-tuples)
- [🔎 Batch checking](#-batch-checking)
- [🧭 Listing objects and users](#-listing-objects-and-users)
- [🧩 Contextual tuples and conditions](#-contextual-tuples-and-conditions)
- [🧬 DSL models](#-dsl-models)
- [🔧 Configuration](#-configuration)
- [🚨 Error handling](#-error-handling)
- [🔌 Extensibility and observability](#-extensibility-and-observability)
- [🧪 Testing against a fake](#-testing-against-a-fake)
- [📌 Stability](#-stability)
- [📖 Documentation](#-documentation)
- [🤝 Contributing](#-contributing)
- [📜 License](#-license)

---

## ✨ Features

- **Full v1 API coverage** — stores, authorization models, relationship tuples, all
  relationship queries (check, batch-check, expand, list-objects, list-users), and
  assertions.
- **Five authentication modes** — no-auth, pre-shared API token, OAuth2
  client-credentials, private-key JWT (RFC 7523 client assertion), and any
  `oauth2.TokenSource`.
- **Auto-paginating iterators** — Go 1.23 range-over-func for stores, models, tuple
  reads and changes, plus manual cursor control when you need it.
- **Bulk & parallel helpers** — `WriteTuples`/`DeleteTuples` chunk large slices into
  parallel non-transactional writes with per-tuple results; `BatchCheckAll` fans a
  check list across parallel batch-check requests; `ListRelations` reports which
  relations a user has on an object.
- **DSL transformer** — the optional `dsl` module converts models between DSL and JSON.
- **Streaming** — `StreamedListObjects` yields results from the NDJSON endpoint as they
  arrive.
- **Configurable retries** — exponential backoff with equal jitter, on by default for
  HTTP 429 and transient network errors, honoring `Retry-After`; 5xx is opt-in.
- **Typed errors** — `*ValidationError`, `*AuthenticationError`, `*NotFoundError`,
  `*RateLimitError`, `*InternalError`, all reachable via `errors.As`.
- **Composable transport** — layer tracing/metrics/logging under the auth+retry
  chain with `WithBaseTransport`, or observe every attempt with
  `WithRequestObserver`.
- **Escape hatch** — `NewRequest`/`Do` let you call any endpoint while reusing the
  configured auth and transport stack.

<p align="right"><sub><a href="#-table-of-contents">↑ Back to table of contents</a></sub></p>

## 📋 Requirements

- Go 1.25 or newer.
- An OpenFGA server to talk to — see the [OpenFGA docs](https://openfga.dev/docs) to run one.

<p align="right"><sub><a href="#-table-of-contents">↑ Back to table of contents</a></sub></p>

## 📦 Installation

```bash
go get github.com/sergiught/go-openfga/openfga
```

<p align="right"><sub><a href="#-table-of-contents">↑ Back to table of contents</a></sub></p>

## 🚀 Quickstart

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

	allowed, err := client.Relationships.Allowed(
		context.Background(), "user:anne", "reader", "document:budget")
	var notFound *openfga.NotFoundError
	switch {
	case errors.As(err, &notFound):
		fmt.Println("store or model not found")
	case err != nil:
		panic(err)
	default:
		fmt.Println("allowed:", allowed)
	}
}
```

`Allowed` is the shortcut for the common check. For contextual tuples, ABAC
context, or a per-call model, build a `CheckRequest` (optionally with
`openfga.NewCheckRequest`) and call `Check`.

Every typed method returns `(result, error)` (or just `error` for writes). To
reach the raw HTTP response — status, headers, or the server request ID — pass
the `openfga.OnResponse` option, which hands your callback the `*Response` after
the body is decoded:

```go
allowed, err := client.Relationships.Allowed(ctx, "user:anne", "reader", "document:budget",
	openfga.OnResponse(func(r *openfga.Response) {
		log.Println("request id:", r.RequestID(), "status:", r.StatusCode)
	}))
```

`OnResponse` fires even on API errors, so you can read headers off a failure. For
cross-cutting observation of every request use `WithRequestObserver`; for full
manual control use `NewRequest` + `Do`. The fan-out helpers (`WriteTuples`,
`DeleteTuples`, `BatchCheckAll`, `ListRelations`) issue several requests and do
not invoke `OnResponse`.

### Bootstrapping a store and model

Starting from nothing? Create a store, write a model, and point the client at
both — after which every call uses them by default:

```go
client, _ := openfga.NewClient("https://api.fga.example", openfga.WithAPIToken("my-api-token"))

store, err := client.Stores.Create(ctx, &openfga.CreateStoreRequest{Name: "acme"})
if err != nil {
	return err
}
client.SetStoreID(store.ID)

// modelReq is a *WriteAuthorizationModelRequest — see "DSL models" below for
// how to build one from DSL text or the typed schema.
model, err := client.AuthorizationModels.Write(ctx, modelReq)
if err != nil {
	return err
}
client.SetAuthorizationModelID(model.AuthorizationModelID)
```

<p align="right"><sub><a href="#-table-of-contents">↑ Back to table of contents</a></sub></p>

## 🔐 Authentication

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

// Any oauth2.TokenSource — for credential sources beyond the built-in modes
// (Vault, workload identity, an existing token source, ...).
openfga.WithTokenSource(mySource)
```

Token fetches for the OAuth2 modes run through the configured base transport
(see [Extensibility](#-extensibility-and-observability)) with a bounded timeout,
so a slow issuer cannot wedge your requests indefinitely.

<p align="right"><sub><a href="#-table-of-contents">↑ Back to table of contents</a></sub></p>

## 🌱 Configuration from the environment

`NewClient` never reads the environment. Opt in with `NewClientFromEnv`, which
resolves `FGA_*` variables; explicit options override them.

```go
client, err := openfga.NewClientFromEnv(openfga.WithUserAgent("my-app/1.0"))
```

| Variable | Maps to |
| --- | --- |
| `FGA_API_URL` | Base URL |
| `FGA_STORE_ID` | Default store ID |
| `FGA_MODEL_ID` | Default authorization model ID |
| `FGA_API_TOKEN` | Pre-shared API token auth |
| `FGA_CLIENT_ID` / `FGA_CLIENT_SECRET` | OAuth2 client-credentials auth |
| `FGA_API_TOKEN_ISSUER` | OAuth2 token endpoint |
| `FGA_API_AUDIENCE` | OAuth2 audience |
| `FGA_API_SCOPES` | OAuth2 scopes (comma-separated) |

Use `openfga.EnvOptions()` to merge env-derived options with your own in a
custom order.

<p align="right"><sub><a href="#-table-of-contents">↑ Back to table of contents</a></sub></p>

## 📄 Pagination

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

<p align="right"><sub><a href="#-table-of-contents">↑ Back to table of contents</a></sub></p>

## 📝 Writing tuples

A single transactional write (all-or-nothing, capped by the server at ~100 tuples):

```go
_, err := client.Tuples.Write(ctx, &openfga.WriteRequest{
	Writes: &openfga.WriteRequestTuples{
		TupleKeys: []openfga.TupleKey{
			{User: "user:anne", Relation: "reader", Object: "document:budget"},
		},
	},
})
```

### Bulk writes and deletes

`WriteTuples` and `DeleteTuples` accept arbitrarily large slices. By default they
split the input into non-transactional chunks issued in parallel, so one chunk
failing doesn't roll back the rest. The response reports a per-tuple outcome
(order matches the input). Each chunk is one server-side atomic write, so a failed
chunk marks all of its tuples failed with the same error — use `WithMaxPerChunk(1)`
for exact per-tuple attribution:

```go
resp, err := client.Tuples.WriteTuples(ctx, keys,
	openfga.WithMaxPerChunk(50),  // tuples per request (default 50)
	openfga.WithMaxParallel(10),  // concurrent requests (default 10)
)
if err != nil {
	return err // only set when no request could be issued at all
}
if err := resp.FirstError(); err != nil {
	return err // a partial failure: some chunk(s) failed
}
for _, r := range resp.Writes {
	if r.Status == openfga.WriteStatusFailure {
		fmt.Println("failed:", r.TupleKey, r.Err)
	}
}
```

Because per-tuple failures are reported in the response rather than the returned
error, call `resp.FirstError()` (or `resp.Failed()` for the full list) to detect a
partial failure with one check. Pass `openfga.WithTransaction()` to send
everything as one transactional request instead of chunking.

### Write-conflict handling

On OpenFGA ≥ 1.10 you can tell the server to ignore a write whose tuple already
exists, or a delete whose tuple is missing, instead of erroring. Set the fields on
the request block, or use the options on the bulk helpers:

```go
// On the raw Write request:
&openfga.WriteRequestTuples{TupleKeys: keys, OnDuplicate: openfga.OnDuplicateIgnore}

// On the bulk helpers:
client.Tuples.WriteTuples(ctx, keys, openfga.WithOnDuplicate(openfga.OnDuplicateIgnore))
client.Tuples.DeleteTuples(ctx, keys, openfga.WithOnMissing(openfga.OnMissingIgnore))
```

<p align="right"><sub><a href="#-table-of-contents">↑ Back to table of contents</a></sub></p>

## 🔎 Batch checking

The native `Relationships.BatchCheck` sends up to the server's per-request limit in
one call. `BatchCheckAll` accepts any number of checks, splits them across parallel
`/batch-check` requests, and merges the results into one map keyed by correlation
ID. Items without a `CorrelationID` get one generated automatically:

```go
resp, err := client.Relationships.BatchCheckAll(ctx, &openfga.BatchCheckRequest{
	Checks: checks, // any length
}, openfga.WithMaxChecksPerBatch(50), openfga.WithMaxParallel(10))
if err != nil {
	return err
}
for id, result := range resp.Result {
	fmt.Println(id, result.Allowed)
}
```

`ListRelations` answers "which of these relations does the user have on this
object?" — useful for deciding which actions to enable in a UI. It runs the
candidate relations through `BatchCheckAll` and returns the allowed ones, in the
order supplied:

```go
allowed, err := client.Relationships.ListRelations(ctx, &openfga.ListRelationsRequest{
	User:      "user:anne",
	Object:    "document:budget",
	Relations: []string{"can_view", "can_edit", "can_delete"},
})
// allowed == []string{"can_view", "can_edit"}
```

<p align="right"><sub><a href="#-table-of-contents">↑ Back to table of contents</a></sub></p>

## 🧭 Listing objects and users

`ListObjects` answers the inverse of a check — "which objects of a type can this
user access?" — and is the query behind most "list the things I can see" views:

```go
resp, err := client.Relationships.ListObjects(ctx, &openfga.ListObjectsRequest{
	Type:     "document",
	Relation: "reader",
	User:     "user:anne",
})
// resp.Objects == []string{"document:budget", "document:roadmap"}
```

`ListUsers` is the mirror image (which users have a relation on an object) and
`Expand` returns the full userset tree for a single relation. When the result set
is large, `StreamedListObjects` yields objects from the NDJSON endpoint as they
arrive instead of buffering the whole page:

```go
for obj, err := range client.Relationships.StreamedListObjects(ctx, req) {
	if err != nil {
		return err
	}
	fmt.Println(obj.Object)
}
```

<p align="right"><sub><a href="#-table-of-contents">↑ Back to table of contents</a></sub></p>

## 🧩 Contextual tuples and conditions

Pass tuples that exist only for the duration of one query via `ContextualTuples`,
and supply the `Context` that OpenFGA's conditioned (ABAC) relations evaluate
against. Both hang off a `CheckRequest` — `NewCheckRequest` fills in the common
fields so you only set what you need:

```go
req := openfga.NewCheckRequest("user:anne", "reader", "document:budget")
req.ContextualTuples = &openfga.ContextualTupleKeys{
	TupleKeys: []openfga.TupleKey{
		{User: "user:anne", Relation: "member", Object: "team:finance"},
	},
}
req.Context = map[string]any{"current_time": "2026-01-01T09:00:00Z"}

resp, err := client.Relationships.Check(ctx, req)
```

Contextual tuples and `Context` are accepted by the query methods
(`Check`, `BatchCheck`, `ListObjects`, `ListUsers`) the same way.

<p align="right"><sub><a href="#-table-of-contents">↑ Back to table of contents</a></sub></p>

## 🧬 DSL models

The `dsl` module converts between OpenFGA's DSL syntax and the JSON model types. It
lives in a separate module so its transformer dependency stays out of the core SDK's
graph — install it only if you need it:

```bash
go get github.com/sergiught/go-openfga/dsl
```

```go
import "github.com/sergiught/go-openfga/dsl"

// DSL text -> a model you can pass to AuthorizationModels.Write.
req, err := dsl.ToModel(dslText)

// A model -> DSL text.
out, err := dsl.ToDSL(model)
```

### Authoring models in Go

For building a model programmatically without the DSL, the core package exposes a
strongly-typed schema and small builder helpers, so relation rewrites read close
to the DSL and the compiler checks your work:

```go
req := &openfga.WriteAuthorizationModelRequest{
	SchemaVersion: "1.1",
	TypeDefinitions: []openfga.TypeDefinition{
		{Type: "user"},
		{
			Type: "document",
			Relations: map[string]openfga.Userset{
				"owner":  openfga.This(),
				"editor": openfga.Union(openfga.This(), openfga.ComputedUserset("owner")),
				"viewer": openfga.TupleTo("parent", "viewer"), // "viewer from parent"
			},
			Metadata: &openfga.Metadata{
				Relations: map[string]openfga.RelationMetadata{
					"owner":  {DirectlyRelatedUserTypes: []openfga.RelationReference{openfga.DirectType("user")}},
					"editor": {DirectlyRelatedUserTypes: []openfga.RelationReference{openfga.DirectType("user")}},
				},
			},
		},
	},
}
```

The builders map to the DSL operators: `This` (`[...]`), `ComputedUserset` (a bare
relation), `TupleTo` (`X from Y`), `Union` (`or`), `Intersection` (`and`), and
`Exclusion` (`but not`). The typed schema round-trips losslessly with the `dsl`
module, so you can mix the two.

<p align="right"><sub><a href="#-table-of-contents">↑ Back to table of contents</a></sub></p>

## 🔧 Configuration

Client-wide options are passed to `NewClient`:

| Option | Purpose |
| --- | --- |
| `WithStoreID` / `WithAuthorizationModelID` | Defaults applied to every request. |
| `WithDefaultConsistency` | Default read consistency for queries and reads. |
| `WithAPIToken` / `WithClientCredentials` / `WithPrivateKeyJWT` / `WithTokenSource` | Authentication. |
| `WithRetry(openfga.RetryConfig{...})` / `WithoutRetry()` | Tune or disable retries. |
| `WithHeaders(http.Header{...})` | Static headers on every request. |
| `WithUserAgent` / `WithBaseURL` | Override the User-Agent or base URL. |
| `WithBaseTransport` / `WithRequestObserver` | Add tracing/metrics/logging beneath the chain, or observe each attempt. |
| `WithHTTPClient` | Supply your own `*http.Client` (disables the built-in transport chain). |

Per-call options override client defaults for a single request:
`WithStore`, `WithAuthorizationModel`, `WithConsistency`, and `WithRequestHeader`.
(Each client-wide default has a matching per-call override: `WithStoreID`/`WithStore`,
`WithAuthorizationModelID`/`WithAuthorizationModel`,
`WithDefaultConsistency`/`WithConsistency`.)

### Retries

Retries are on by default: 3 attempts, exponential backoff with equal jitter
between 1s and 30s, honoring the server's `Retry-After`. By default only
**HTTP 429** and transient network failures (connection resets, refused dials,
timeouts) are retried.

**5xx responses are not retried by default.** This is deliberate — a 5xx on a
non-idempotent write is ambiguous, so blindly retrying it risks duplicating the
effect. (Note this differs from the official `openfga/go-sdk`, which retries 5xx
out of the box; if you are migrating and want that behavior, opt in explicitly.)
Add the statuses you want to `RetryableStatus`:

```go
openfga.WithRetry(openfga.RetryConfig{
	RetryableStatus: []int{429, 500, 502, 503, 504},
})
```

`WithRetry` merges with the defaults, so setting one field (e.g. `MaxAttempts`)
keeps the rest. Use `WithoutRetry()` to disable retries entirely. The whole call
— across every retry — is bounded by the context deadline you pass in.

<p align="right"><sub><a href="#-table-of-contents">↑ Back to table of contents</a></sub></p>

## 🚨 Error handling

Non-2xx responses become typed errors, all embedding `*ErrorResponse` and
reachable with `errors.As`:

| Type | HTTP status |
| --- | --- |
| `*ValidationError` | 400 |
| `*AuthenticationError` | 401, 403 |
| `*NotFoundError` | 404 |
| `*RateLimitError` (carries `RetryAfter`) | 429 |
| `*InternalError` | 5xx |

```go
allowed, err := client.Relationships.Allowed(ctx, "user:anne", "reader", "document:budget")

var rl *openfga.RateLimitError
switch {
case errors.As(err, &rl):
	// rl.RetryAfter, rl.RequestID()
case err != nil:
	// inspect (*openfga.ErrorResponse).Code — see the openfga.Code* constants
}
```

`ErrorResponse.Code` holds the OpenFGA error code (match it against the
`openfga.Code*` constants), and `ErrorResponse.RequestID()` returns the server
correlation ID for support tickets.

<p align="right"><sub><a href="#-table-of-contents">↑ Back to table of contents</a></sub></p>

## 🔌 Extensibility and observability

The client owns only an `*http.Client`; auth, retries, and headers are layered
as `http.RoundTripper` transports. To add tracing, metrics, or a custom dialer
while keeping the SDK's auth and retries, set the innermost transport:

```go
openfga.WithBaseTransport(otelhttp.NewTransport(nil))
```

For lightweight logging or metrics without writing a transport, observe each
attempt:

```go
openfga.WithRequestObserver(func(req *http.Request, resp *http.Response, err error, took time.Duration) {
	log.Printf("%s %s -> %v (%s)", req.Method, req.URL.Path, statusOf(resp, err), took)
})
```

`client.Transport()` returns the assembled chain if you want to reuse it
elsewhere. `WithHTTPClient` remains the full escape hatch, but it replaces the
entire chain (auth, retries, and headers included).

<p align="right"><sub><a href="#-table-of-contents">↑ Back to table of contents</a></sub></p>

## 🧪 Testing against a fake

The client talks to any base URL, so point it at an `httptest.Server` in unit
tests — no live OpenFGA required:

```go
srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"allowed": true}`))
}))
defer srv.Close()

client, _ := openfga.NewClient(srv.URL, openfga.WithStoreID("01ARZ3NDEKTSV4RRFFQ69G5FAV"))
```

Give each request a deadline with `context.WithTimeout`; the deadline bounds the
whole call including retries.

<p align="right"><sub><a href="#-table-of-contents">↑ Back to table of contents</a></sub></p>

## 📌 Stability

Pre-1.0: the public API may change between minor versions. Pin a version and
review the [changelog](CHANGELOG.md) before upgrading. Once tagged `v1.0.0`, the
package follows semantic versioning.

<p align="right"><sub><a href="#-table-of-contents">↑ Back to table of contents</a></sub></p>

## 📖 Documentation

Full API documentation, with runnable examples for the major entry points, lives on
[pkg.go.dev](https://pkg.go.dev/github.com/sergiught/go-openfga/openfga).

<p align="right"><sub><a href="#-table-of-contents">↑ Back to table of contents</a></sub></p>

## 🤝 Contributing

Contributions are welcome — see [CONTRIBUTING.md](CONTRIBUTING.md) for the development
workflow, and the [Code of Conduct](CODE_OF_CONDUCT.md). To report a security issue,
follow the [security policy](SECURITY.md).

<p align="right"><sub><a href="#-table-of-contents">↑ Back to table of contents</a></sub></p>

## 📜 License

[MIT](LICENSE) © 2024-2026 Sergiu Ghitea.

This project is an independent client and is not affiliated with or endorsed by the
OpenFGA project or the CNCF.

<p align="right"><sub><a href="#-table-of-contents">↑ Back to table of contents</a></sub></p>
