package client

import (
	"context"
	"fmt"

	managementv2 "github.com/auth0/go-auth0/v2/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// NewCIMDResource will return a new auth0_client_cimd resource.
func NewCIMDResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createCIMDClient,
		ReadContext:   readClient,
		UpdateContext: updateClient,
		DeleteContext: deleteCIMDClient,
		Importer: &schema.ResourceImporter{
			StateContext: importCIMDClient,
		},
		Description: "With this resource, you can register an Auth0 client from a " +
			"Client ID Metadata Document (CIMD) URL. CIMD enables tenant admins to " +
			"onboard MCP agent clients by providing a URL to an externally-hosted " +
			"metadata document instead of using Dynamic Client Registration.\n\n" +
			"Requires the `client_id_metadata_document_supported` tenant setting to be enabled.",
		Schema: cimdClientSchema(),
	}
}

// cimdClientSchema derives the CIMD schema from auth0_client.
// CIMD clients are Strict 3P (third_party_security_mode: "strict") with an
// additional blocklist. PATCHable = (Strict 3P Allowlist) − (CIMD_PATCH_BLOCKED_FIELDS).
// See CIMD_DESIGN_DECISIONS.md for sources and API test evidence.
func cimdClientSchema() map[string]*schema.Schema {
	base := NewResource().Schema

	// PATCHable top-level fields. Sub-fields of jwt_configuration and
	// refresh_token are restricted separately below via sub-field allowlists.
	// Note: redirection_policy is PATCHable but not yet in the go-auth0 SDK.
	cimdEditable := map[string]bool{
		"allowed_origins":                true,
		"app_type":                       true,
		"description":                    true,
		"oidc_conformant":                true,
		"organization_usage":             true,
		"organization_require_behavior":  true,
		"organization_discovery_methods": true,
		"web_origins":                    true,
		"grant_types":                    true,
		"client_metadata":                true,
		"default_organization":           true,
		"require_proof_of_possession":    true,
		"token_quota":                    true,
		"jwt_configuration":              true,
		"refresh_token":                  true,
		"skip_non_verifiable_callback_uri_confirmation_prompt": true,
	}

	for key, s := range base {
		if cimdEditable[key] {
			s.Required = false
			s.Optional = true
			s.Computed = true
			s.ForceNew = false
		} else {
			makeSchemaReadOnly(s)
		}
	}

	// Computed in auth0_client → Required+ForceNew for CIMD registration URL.
	base["external_client_id"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
		Description: "The HTTPS URL of the Client ID Metadata Document. " +
			"Must include a path component (e.g. `https://app.example.com/client.json`). " +
			"This value is immutable after creation.",
	}

	// CIMD only supports native, regular_web, spa (API rejects others).
	base["app_type"].ValidateFunc = validation.StringInSlice([]string{"native", "regular_web", "spa"}, false)
	base["app_type"].Description = "Type of application the client represents. " +
		"CIMD clients only support `native`, `spa`, and `regular_web`."

	// jwt_configuration: only lifetime_in_seconds and alg are PATCHable.
	jwtEditable := map[string]bool{
		"lifetime_in_seconds": true,
		"alg":                 true,
	}
	jwtSub := base["jwt_configuration"].Elem.(*schema.Resource).Schema
	for key, s := range jwtSub {
		if !jwtEditable[key] {
			makeSchemaReadOnly(s)
		}
	}
	// CIMD restricts alg to asymmetric algorithms (no HS256).
	jwtSub["alg"].Computed = true
	jwtSub["alg"].ValidateFunc = validation.StringInSlice([]string{"RS256", "RS512", "PS256"}, false)
	jwtSub["alg"].Description = "Algorithm used to sign JWTs. " +
		"CIMD clients support `RS256`, `RS512`, and `PS256` (asymmetric only)."

	// refresh_token: only these 5 sub-fields are PATCHable.
	rtEditable := map[string]bool{
		"rotation_type":           true,
		"leeway":                  true,
		"token_lifetime":          true,
		"infinite_token_lifetime": true,
		"idle_token_lifetime":     true,
	}
	rtSub := base["refresh_token"].Elem.(*schema.Resource).Schema
	for key, s := range rtSub {
		if !rtEditable[key] {
			makeSchemaReadOnly(s)
		}
	}
	// rotation_type is Required in auth0_client but Optional for CIMD
	rtSub["rotation_type"].Required = false
	rtSub["rotation_type"].Optional = true
	rtSub["rotation_type"].Computed = true

	return base
}

// makeSchemaReadOnly recursively sets a field and nested sub-fields to Computed-only.
func makeSchemaReadOnly(s *schema.Schema) {
	s.Required = false
	s.Optional = false
	s.Computed = true
	s.ForceNew = false
	s.Default = nil
	s.ValidateFunc = nil
	s.ValidateDiagFunc = nil
	s.DiffSuppressFunc = nil
	s.AtLeastOneOf = nil
	s.RequiredWith = nil
	s.ConflictsWith = nil
	s.ExactlyOneOf = nil
	s.MaxItems = 0
	s.MinItems = 0

	if r, ok := s.Elem.(*schema.Resource); ok {
		for _, sub := range r.Schema {
			makeSchemaReadOnly(sub)
		}
	}
}

func createCIMDClient(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()

	// Register via CIMD endpoint (v2 SDK).
	externalClientID := data.Get("external_client_id").(string)

	req := &managementv2.RegisterCimdClientRequestContent{}
	req.SetExternalClientID(externalClientID)

	result, err := apiv2.Clients.RegisterCimdClient(ctx, req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("CIMD registration failed: %w", err))
	}

	clientID := result.GetClientID()
	if clientID == "" {
		return diag.Errorf("CIMD registration response missing client_id")
	}

	data.SetId(clientID)

	// PATCH editable fields via v1 SDK (reuses expandClient).
	api := meta.(*config.Config).GetAPI()

	client, err := expandClient(data)
	if err != nil {
		return diag.FromErr(err)
	}

	// expandClient sets TokenEndpointAuthMethod when IsNewResource() is true
	// TokenEndpointAuthMethod is a CIMD-blocked field. We clear it here
	// rather than modifying expandClient to avoid coupling expand.go to
	// CIMD-specific logic.
	client.TokenEndpointAuthMethod = nil

	// Skip PATCH if expandClient produced an empty struct (user only set
	// external_client_id with no editable fields).
	if clientHasChange(client) {
		if err := api.Client.Update(ctx, data.Id(), client); err != nil {
			return diag.FromErr(internalError.HandleAPIError(data, err))
		}
	}

	return readClient(ctx, data, meta)
}

func deleteCIMDClient(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()

	if err := apiv2.Clients.Delete(ctx, data.Id()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func importCIMDClient(ctx context.Context, data *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	api := meta.(*config.Config).GetAPI()

	client, err := api.Client.Read(ctx, data.Id())
	if err != nil {
		return nil, err
	}

	if client.GetExternalMetadataType() != "cimd" {
		return nil, fmt.Errorf(
			"client %q is not a CIMD client (external_metadata_type=%q). "+
				"Use the auth0_client resource to manage regular clients",
			data.Id(),
			client.GetExternalMetadataType(),
		)
	}

	data.SetId(client.GetClientID())

	return []*schema.ResourceData{data}, nil
}
