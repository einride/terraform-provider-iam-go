package iamgo

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestIAMGoMemberResource(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: providers(newMockClient()),
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
