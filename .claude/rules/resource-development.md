# Resource Development Patterns

## Resource Structure

Each Auth0 resource must include:
- **resource.go** — `NewResource()` returns `*schema.Resource` with CreateContext, ReadContext, UpdateContext, DeleteContext handlers
- **expand.go** — Convert Terraform state → Auth0 API structs using `value.*()` helpers from `internal/value/`
- **flatten.go** — Convert Auth0 API response → `[]interface{}` or `map[string]interface{}`
- **data_source.go** — Read-only variant with same schema + flatten logic

## Registration & Error Handling

Register in `internal/provider/provider.go` ResourcesMap (and DataSourcesMap for data sources). Use `internalError` from `internal/error/` for 404 detection to auto-remove deleted resources from state.

## API Versions

- **v1:** `meta.(*config.Config).GetAPI()` → `*management.Management`
- **v2:** `meta.(*config.Config).GetAPIV2()` → `*managementv2.Management`

Prefer v2 for new resources where available.
