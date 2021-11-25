package main

import (
	iamgo "terraform-provider-iam-go/iam-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{ProviderFunc: iamgo.Provider})
}
