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
				ResourceName:      "auth0_client_cimd.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
