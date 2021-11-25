package iam_go

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/genproto/googleapis/iam/v1"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"address": {
				Type:     schema.TypeString,
				Required: true,
			},
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("IAM_GO_TOKEN", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"iam-go_member": resourceMember(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func setupConnection(ctx context.Context, address string, token string) (iam.IAMPolicyClient, error) {
	connection, err := Connect(ctx, address, token)
	if err != nil {
		return nil, err
	}
	return iam.NewIAMPolicyClient(connection), nil
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	address := d.Get("address").(string)
	token := d.Get("token").(string)

	client, err := setupConnection(ctx, address, token)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return newPolicyUpdate(client), diags
}
