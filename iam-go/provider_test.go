package iamgo

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/genproto/googleapis/iam/v1"
	"google.golang.org/grpc"
)

type (
	providerFactories map[string]func() (*schema.Provider, error)
	provider          func() (*schema.Provider, error)
)

// providers are used to instantiate a provider during acceptance testing.
// The factory function will be invoked for every Terraform CLI command executed
// to create a provider server to which the CLI can reattach.
func providers(client *mockIamService) providerFactories {
	p := providerFactories{}
	p["iam-go"] = testIAMGoProvider(client)
	return p
}

func testIAMGoProvider(client *mockIamService) provider {
	return func() (*schema.Provider, error) {
		return &schema.Provider{
			Schema: map[string]*schema.Schema{},
			ResourcesMap: map[string]*schema.Resource{
				"iam-go_member": resourceMember(),
			},
			ConfigureContextFunc: testProviderConfigure(client),
		}, nil
	}
}

func testProviderConfigure(client *mockIamService) schema.ConfigureContextFunc {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		var diags diag.Diagnostics
		return newPolicyUpdate(client), diags
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func newMockClient() *mockIamService {
	return &mockIamService{
		make(map[string]*iam.Policy),
	}
}

var _ iam.IAMPolicyClient = &mockIamService{}

type mockIamService struct {
	policies map[string]*iam.Policy
}

func (m mockIamService) SetIamPolicy(
	ctx context.Context,
	req *iam.SetIamPolicyRequest,
	opts ...grpc.CallOption,
) (*iam.Policy, error) {
	m.policies[req.GetResource()] = req.Policy
	return req.Policy, nil
}

func (m mockIamService) GetIamPolicy(
	ctx context.Context,
	req *iam.GetIamPolicyRequest,
	opts ...grpc.CallOption,
) (*iam.Policy, error) {
	if val, ok := m.policies[req.GetResource()]; ok {
		return val, nil
	}
	return &iam.Policy{}, nil
}

func (m mockIamService) TestIamPermissions(
	ctx context.Context,
	req *iam.TestIamPermissionsRequest,
	opts ...grpc.CallOption,
) (*iam.TestIamPermissionsResponse, error) {
	panic("implement me")
}
