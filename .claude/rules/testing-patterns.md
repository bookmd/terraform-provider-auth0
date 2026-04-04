# Testing Patterns

## Framework & Recordings

Always use `acctest.Test()` instead of `resource.Test()` directly. Tests automatically run in parallel when HTTP recordings enabled.

HTTP recordings stored in `test/data/recordings/`. When `AUTH0_HTTP_RECORDINGS=on`, interactions replay without hitting Auth0 tenant.

**IMPORTANT:** To re-record a test, delete old cassette file first — new interactions won't overwrite existing recordings.

## Test Commands

```bash
make test-unit FILTER="TestName"          # Unit tests (no credentials)
make test-acc FILTER="TestName"           # Acceptance with recordings (no credentials)
make test-acc-record FILTER="TestName"    # Record new interactions (needs credentials)
```

## Test HCL Config

Use Go templates for unique naming: `{{.testName}}` extracted via `acctest.ParseTestName()` to avoid parallel test conflicts.

Include TestAccXxxDestroy checks to verify deletion.
