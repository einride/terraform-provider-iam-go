package iam_go

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"go.einride.tech/iam/iampolicy"
	"google.golang.org/genproto/googleapis/iam/v1"
)

func resourceMember() *schema.Resource {
	noWhiteSpaceValidation := validation.StringDoesNotMatch(regexp.MustCompile(`.*\s.*`), "contains whitespace")
	return &schema.Resource{
		CreateContext: resourceMemberCreate,
		ReadContext:   resourceMemberRead,
		DeleteContext: resourceMemberDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"resource": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: noWhiteSpaceValidation,
			},
			"role": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: noWhiteSpaceValidation,
			},
			"member": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: noWhiteSpaceValidation,
			},
		},
	}
}

type iamMember struct {
	resource string
	role     string
	member   string
}

func resourceMemberCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	IAMMember := getResourceIAMMember(data)

	ePolicy, ok := meta.(*policyUpdate)
	if !ok {
		return diag.Errorf("meta interface did not provide policyUpdate")
	}
	unlock := ePolicy.lock(resourceName(IAMMember.resource))
	defer unlock()

	policy, err := getIamPolicy(ctx, ePolicy.client, IAMMember.resource)
	if err != nil {
		return diag.FromErr(err)
	}

	iampolicy.AddBinding(policy, IAMMember.role, IAMMember.member)

	_, err = setIamPolicy(ctx, ePolicy.client, IAMMember.resource, policy)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(fmt.Sprintf("%s %s %s", IAMMember.resource, IAMMember.role, IAMMember.member))

	resourceMemberRead(ctx, data, meta)

	return diags
}

func resourceMemberRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	IAMMember := getResourceIAMMember(data)

	ePolicy, ok := meta.(*policyUpdate)
	if !ok {
		return diag.Errorf("meta interface did not provide policyUpdate")
	}

	policy, err := getIamPolicy(ctx, ePolicy.client, IAMMember.resource)
	if err != nil {
		return diag.FromErr(err)
	}

	if !contains(IAMMember, policy) {
		data.SetId("")
		return diags
	}

	if err := data.Set("member", IAMMember.member); err != nil {
		return diag.Errorf("error setting member: %s", err)
	}
	if err := data.Set("role", IAMMember.role); err != nil {
		return diag.Errorf("error setting role: %s", err)
	}
	if err := data.Set("resource", IAMMember.resource); err != nil {
		return diag.Errorf("error setting resource: %s", err)
	}
	return diags
}

func resourceMemberDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	IAMMember := getResourceIAMMember(data)

	ePolicy, ok := meta.(*policyUpdate)
	if !ok {
		return diag.Errorf("meta interface did not provide policyUpdate")
	}
	unlock := ePolicy.lock(resourceName(IAMMember.resource))
	defer unlock()

	policy, err := getIamPolicy(ctx, ePolicy.client, IAMMember.resource)
	if err != nil {
		return diag.FromErr(err)
	}

	iampolicy.RemoveBinding(policy, IAMMember.role, IAMMember.member)

	_, err = setIamPolicy(ctx, ePolicy.client, IAMMember.resource, policy)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId("")

	return diags
}

func getResourceIAMMember(data *schema.ResourceData) *iamMember {
	if data.Id() != "" {
		fields := strings.Fields(data.Id())
		return &iamMember{
			resource: fields[0],
			role:     fields[1],
			member:   fields[2],
		}
	}

	return &iamMember{
		resource: data.Get("resource").(string),
		role:     data.Get("role").(string),
		member:   data.Get("member").(string),
	}
}

func setIamPolicy(
	ctx context.Context,
	client iam.IAMPolicyClient,
	resource string,
	policy *iam.Policy,
) (*iam.Policy, error) {
	return client.SetIamPolicy(
		ctx,
		&iam.SetIamPolicyRequest{
			Resource: resource,
			Policy:   policy,
		},
	)
}

func getIamPolicy(ctx context.Context, client iam.IAMPolicyClient, resource string) (*iam.Policy, error) {
	return client.GetIamPolicy(
		ctx,
		&iam.GetIamPolicyRequest{
			Resource: resource,
			Options:  nil,
		},
	)
}

func contains(iamMember *iamMember, policy *iam.Policy) bool {
	var binding *iam.Binding
	for _, b := range policy.Bindings {
		if b.Role == iamMember.role {
			binding = b
			break
		}
	}

	if binding == nil {
		return false
	}

	for _, mem := range binding.Members {
		if mem == iamMember.member {
			return true
		}
	}
	return false
}
