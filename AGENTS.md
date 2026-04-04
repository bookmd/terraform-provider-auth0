# AGENTS.md

## Overview

Official Terraform provider for Auth0, built with HashiCorp's Terraform Plugin SDK v2 and Go 1.24. Manages Auth0 tenant configuration as infrastructure-as-code, providing 72 resources and 40 data sources across 31 domain packages.

## Architecture

### Resource Layout

Each Auth0 resource lives in `internal/auth0/<domain>/` following a uniform pattern:

- **resource.go** — CRUD operations (CreateContext, ReadContext, UpdateContext, DeleteContext) + resource schema definition
- **expand.go** — Converts Terraform state data → Auth0 API structs (uses `value.*()` helpers from `internal/value`)
- **flatten.go** — Converts Auth0 API response → Terraform-compatible format (`[]interface{}` or `map[string]interface{}`)
- **data_source.go** — Read-only data source using the same schema transformation
- **tests** — Unit and acceptance tests with HTTP recording support

Resources are registered in `internal/provider/provider.go` within ResourcesMap. Data sources are registered in DataSourcesMap.

### Key Packages

| Package | Purpose |
|---------|---------|
| `internal/config/` | API client setup, authentication, connection management |
| `internal/acctest/` | Test framework wrapper, HTTP recorder (go-vcr) integration |
| `internal/schema/` | Resource-to-datasource schema transformation utilities |
| `internal/error/` | 404 detection; auto-removes deleted resources from Terraform state |
| `internal/value/` | Type conversion helpers for `cty.Value` → Go types |
| `internal/validation/` | Custom validators for resource schemas |
| `internal/mutex/` | Distributed locking for concurrent resource operations |

### Auth0 SDK: Dual Version Strategy

The provider uses **both** `go-auth0` v1 and v2 concurrently during incremental migration:

- **v1:** `meta.(*config.Config).GetAPI()` returns `*management.Management` (import: `github.com/auth0/go-auth0/management`)
- **v2:** `meta.(*config.Config).GetAPIV2()` returns `*managementv2.Management` (import: `github.com/auth0/go-auth0/v2/management/client`)

**Guidance:** New resources should prefer v2 where available. Both can coexist during the migration.

## Development Commands

All commands use the Makefile:

```bash
# Build & Install
make build VERSION=0.2.0          # Build provider binary
make install VERSION=0.2.0        # Install as local Terraform plugin

# Code Quality
make lint                          # Run golangci-lint with auto-fix
make check-vuln                    # Run govulncheck
make docs                          # Generate docs (go generate)
make check-docs                    # Verify generated docs are up-to-date

# Testing (append FILTER="TestName" to run specific tests)
make test-unit                     # Unit tests (no Auth0 credentials needed)
make test-acc                      # Acceptance tests with HTTP recordings
make test-acc-record               # Record new HTTP interactions (needs credentials)
make test-acc-e2e                  # Run against real Auth0 tenant (needs credentials)
```

## Testing System

### HTTP Recordings

Tests use **HTTP recordings** (go-vcr) stored in `test/data/recordings/`. When `AUTH0_HTTP_RECORDINGS=on`, recorded interactions are replayed instead of hitting a real Auth0 tenant, enabling credential-free CI runs.

- `acctest.Test()` wraps `resource.Test()` / `resource.ParallelTest()` and automatically runs tests in parallel when recordings are enabled
- Always use `acctest.Test()` instead of calling `resource.Test()` directly
- Test HCL configs use Go templates: `{{.testName}}` for unique naming (via `acctest.ParseTestName()`)
- **IMPORTANT:** To re-record a test, delete the old cassette file in `test/data/recordings/` first — new interactions won't overwrite existing recordings

### Unit vs Acceptance Tests

- **Unit tests:** `make test-unit FILTER="TestName"` — No credentials needed, tests internal logic
- **Acceptance tests:** `make test-acc FILTER="TestName"` — Uses HTTP recordings, no credentials needed for CI

## Important Notes

- **Docs:** Files in `docs/` are auto-generated from resource schemas and `examples/`. Never edit `docs/` directly — run `make docs` to regenerate.
- **Error Handling:** Use `internalError` (from `internal/error/`) for 404 detection to auto-remove deleted resources from state.
- **Environment Setup:** See `CONTRIBUTING.md` for prerequisites and local setup instructions.
- **Resource Registration:** After creating a new resource, register it in `internal/provider/provider.go` ResourcesMap (and DataSourcesMap if offering a data source).

