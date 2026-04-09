package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	managementv2 "github.com/auth0/go-auth0/v2/management"
	"github.com/auth0/go-auth0/v2/management/core"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
)

// NewCIMDResource will return a new auth0_client_cimd resource.
func NewCIMDResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createCIMDClient,
		ReadContext:   readCIMDClient,
		DeleteContext: deleteCIMDClient,
		Importer: &schema.ResourceImporter{
			StateContext: importCIMDClient,
		},
		Description: "With this resource, you can register an Auth0 client from a " +
			"Client ID Metadata Document (CIMD) URL. CIMD enables tenant admins to " +
			"onboard MCP agent clients by providing a URL to an externally-hosted " +
			"metadata document instead of using Dynamic Client Registration.\n\n" +
			"Requires the `client_id_metadata_document_supported` tenant setting to be enabled.",
		Schema: map[string]*schema.Schema{
			"external_client_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				Description: "The HTTPS URL of the Client ID Metadata Document. " +
					"Must include a path component (e.g. `https://app.example.com/client.json`). " +
					"Root-path URLs like `https://example.com/` are rejected per CIMD spec. " +
					"This value is immutable after creation.",
			},
			"client_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Auth0 client ID generated for this CIMD client (typically `tpc_` prefixed).",
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "The client name derived from the CIMD document. " +
					"Managed by the CIMD document and cannot be modified via Terraform.",
			},
			"external_metadata_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of external metadata (always `cimd` for this resource).",
			},
			"external_metadata_created_by": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Who created the client: `admin` (Management API) or `client` (self-registered).",
			},
			"jwks_uri": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "URL for the JSON Web Key Set (JWKS) containing the public keys used for " +
					"`private_key_jwt` authentication. Only present for CIMD clients using `private_key_jwt` authentication.",
			},
		},
	}
}

func createCIMDClient(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()
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

	return readCIMDClient(ctx, data, meta)
}

func readCIMDClient(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()

	client, err := apiv2.Clients.Get(ctx, data.Id(), nil)
	if err != nil {
		var apiErr *core.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	return diag.FromErr(flattenCIMDClient(data, client))
}

func deleteCIMDClient(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()

	if err := apiv2.Clients.Delete(ctx, data.Id()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func flattenCIMDClient(data *schema.ResourceData, client *managementv2.GetClientResponseContent) error {
	result := multierror.Append(
		data.Set("client_id", client.GetClientID()),
		data.Set("name", client.GetName()),
		data.Set("external_client_id", client.GetExternalClientID()),
		data.Set("external_metadata_type", string(client.GetExternalMetadataType())),
		data.Set("external_metadata_created_by", string(client.GetExternalMetadataCreatedBy())),
		data.Set("jwks_uri", client.GetJwksURI()),
	)

	return result.ErrorOrNil()
}

// importCIMDClient validates the client is actually a CIMD client before
// allowing import. Prevents users from accidentally importing a regular
// client into auth0_client_cimd (which would cause destroy+recreate).
func importCIMDClient(ctx context.Context, data *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	apiv2 := meta.(*config.Config).GetAPIV2()

	client, err := apiv2.Clients.Get(ctx, data.Id(), nil)
	if err != nil {
		return nil, err
	}

	if string(client.GetExternalMetadataType()) != "cimd" {
		return nil, fmt.Errorf(
			"client %q is not a CIMD client (external_metadata_type=%q). "+
				"Use the auth0_client resource to manage regular clients",
			data.Id(),
			string(client.GetExternalMetadataType()),
		)
	}

	data.SetId(client.GetClientID())

	return []*schema.ResourceData{data}, nil
}
