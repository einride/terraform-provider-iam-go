package iamgo

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"google.golang.org/genproto/googleapis/iam/v1"
)

func TestIAMGoMemberResource(t *testing.T) {
	// We initiate the provider here in order to pass it to the CheckDestroy case
	testProvider, err := testIAMGoProvider(newMockClient())()
	factories := map[string]func() (*schema.Provider, error){"iam-go": func() (*schema.Provider, error) {
		return testProvider, err
	}}
	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: factories,
		CheckDestroy:      verifyResourceDestroy(testProvider),
		Steps: []resource.TestStep{
			{
				Config: testIAMGoMemberResource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"iam-go_member.foo",
						"member",
						"serviceAccount:admin@foo.iam.gserviceaccount.com",
					),
					resource.TestCheckResourceAttr(
						"iam-go_member.foo",
						"role",
						"roles/account.Admin",
					),
					resource.TestCheckResourceAttr(
						"iam-go_member.foo",
						"resource",
						"/",
					),
					resource.TestCheckResourceAttr(
						"iam-go_member.foo2",
						"member",
						"serviceAccount:admin2@foo.iam.gserviceaccount.com",
					),
					resource.TestCheckResourceAttr(
						"iam-go_member.foo2",
						"role",
						"roles/account.Admin2",
					),
					resource.TestCheckResourceAttr(
						"iam-go_member.foo2",
						"resource",
						"/",
					),
				),
			},
		},
	})
}

func verifyResourceDestroy(provider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.Meta().(*policyUpdate).client
		// loop through the resources in state, verifying each resource is destroyed
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "iam-go_member" {
				continue
			}
			id := strings.Fields(rs.Primary.ID)
			mem := iamMember{
				resource: id[0],
				role:     id[1],
				member:   id[2],
			}
			req := &iam.GetIamPolicyRequest{Resource: mem.resource, Options: nil}
			bb, _ := client.GetIamPolicy(context.Background(), req)
			if contains(&mem, bb) {
				return fmt.Errorf("resource (%s) still exists in remote api", rs.Primary.ID)
			}
		}
		return nil
	}
}

const testIAMGoMemberResource = `
resource "iam-go_member" "foo" {
  role     = "roles/account.Admin"
  member   = "serviceAccount:admin@foo.iam.gserviceaccount.com"
  resource = "/"
}

resource "iam-go_member" "foo2" {
  role     = "roles/account.Admin2"
  member   = "serviceAccount:admin2@foo.iam.gserviceaccount.com"
  resource = "/"
}
`
