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
	noWhiteSpaceValidation := validation.StringDoesNotMatch(regexp.MustCompile(".*\\s.*"), "contains whitespace")
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

func getResourceIAMMember(d *schema.ResourceData) *iamMember {
	if d.Id() != "" {
		fields := strings.Fields(d.Id())
		return &iamMember{
			resource: fields[0],
			role:     fields[1],
			member:   fields[2],
		}
	}

	return &iamMember{
		resource: d.Get("resource").(string),
		role:     d.Get("role").(string),
		member:   d.Get("member").(string),
	}
}

type iamMember struct {
	resource string
	role     string
	member   string
}

func resourceMemberCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	IAMMember := getResourceIAMMember(d)

	p := m.(*policyUpdate)
	unlock := p.lock(resourceName(IAMMember.resource))
	defer unlock()

	policy, err := p.client.GetIamPolicy(ctx, &iam.GetIamPolicyRequest{Resource: IAMMember.resource, Options: nil})

	if err != nil {
		return diag.FromErr(err)
	}

	iampolicy.AddBinding(policy, IAMMember.role, IAMMember.member)

	_, err = p.client.SetIamPolicy(ctx, &iam.SetIamPolicyRequest{Resource: IAMMember.resource, Policy: policy})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s %s %s", IAMMember.resource, IAMMember.role, IAMMember.member))

	resourceMemberRead(ctx, d, m)

	return diags
}

func resourceMemberRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	IAMMember := getResourceIAMMember(d)

	p := m.(*policyUpdate)
	policy, err := p.client.GetIamPolicy(ctx, &iam.GetIamPolicyRequest{Resource: IAMMember.resource, Options: nil})
	if err != nil {
		return diag.FromErr(err)
	}

	var binding *iam.Binding
	for _, b := range policy.Bindings {
		if b.Role == IAMMember.role {
			binding = b
			break
		}
	}

	if binding == nil {
		d.SetId("")
		return diags
	}

	var member string
	for _, mem := range binding.Members {
		if mem == IAMMember.member {
			member = mem
		}
	}

	if member == "" {
		d.SetId("")
		return diags
	}

	if err := d.Set("member", member); err != nil {
		return diag.Errorf("error setting member: %s", err)
	}
	if err := d.Set("role", binding.Role); err != nil {
		return diag.Errorf("error setting role: %s", err)
	}
	if err := d.Set("resource", IAMMember.resource); err != nil {
		return diag.Errorf("error setting resource: %s", err)
	}
	return diags
}

func resourceMemberDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	IAMMember := getResourceIAMMember(d)

	p := m.(*policyUpdate)
	unlock := p.lock(resourceName(IAMMember.resource))
	defer unlock()

	policy, err := p.client.GetIamPolicy(ctx, &iam.GetIamPolicyRequest{Resource: IAMMember.resource, Options: nil})
	if err != nil {
		return diag.FromErr(err)
	}

	iampolicy.RemoveBinding(policy, IAMMember.role, IAMMember.member)

	_, err = p.client.SetIamPolicy(ctx, &iam.SetIamPolicyRequest{Resource: IAMMember.resource, Policy: policy})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
