# Dual API Version Coordination

## Overview

Provider incrementally migrates from go-auth0 v1 to v2. Both versions coexist during migration.

## API Access

- **v1 (Legacy):** `meta.(*config.Config).GetAPI()` → `*management.Management` (import: `github.com/auth0/go-auth0/management`)
- **v2 (New):** `meta.(*config.Config).GetAPIV2()` → `*managementv2.Management` (import: `github.com/auth0/go-auth0/v2/management/client`)

## Migration Guidance

**New Resources:** Use v2 where available.

**Existing Resources:** Only migrate if v2 endpoint exists and behaves identically. Avoid partial migrations.

**Expand/Flatten:** Conversion logic remains identical across v1/v2 structs — only the API call differs.

See `internal/config/` for API client initialization.
