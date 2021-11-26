package main

import (
	"github.com/einride/terraform-provider-iam-go/iamgo"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{ProviderFunc: iamgo.Provider})
}
