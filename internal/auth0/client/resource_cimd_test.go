package client_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccClientCIMDCreate = `
resource "auth0_client_cimd" "test" {
    external_client_id = "https://tinywiki.xyz/client.json"
}
`

const testAccClientCIMDWithEditableFields = `
resource "auth0_client_cimd" "test" {
    external_client_id = "https://tinywiki.xyz/client.json"
    description        = "CIMD test client"
    allowed_origins    = ["https://example.com"]
    web_origins        = ["https://example.com"]
}
`

const testAccClientCIMDUpdate = `
resource "auth0_client_cimd" "test" {
    external_client_id = "https://tinywiki.xyz/client.json"
    description        = "Updated CIMD test client"
    allowed_origins    = ["https://example.com", "https://other.com"]
    web_origins        = ["https://example.com"]
}
`

func TestAccClientCIMD(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccClientCIMDCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("auth0_client_cimd.test", "client_id"),
					resource.TestCheckResourceAttrSet("auth0_client_cimd.test", "name"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "external_client_id", "https://tinywiki.xyz/client.json"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "external_metadata_type", "cimd"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "external_metadata_created_by", "admin"),
				),
			},
			{
				Config: testAccClientCIMDWithEditableFields,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "description", "CIMD test client"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "allowed_origins.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "allowed_origins.0", "https://example.com"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "web_origins.#", "1"),
					resource.TestCheckResourceAttrSet("auth0_client_cimd.test", "name"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "external_metadata_type", "cimd"),
				),
			},
			{
				Config: testAccClientCIMDUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "description", "Updated CIMD test client"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "allowed_origins.#", "2"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "external_metadata_type", "cimd"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "external_metadata_created_by", "admin"),
				),
			},
			{
				ResourceName:      "auth0_client_cimd.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
