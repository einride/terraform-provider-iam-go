# Terraform Provider IAM Go

A terraform provider for https://github.com/einride/iam-go

## Requirements
-	[Terraform](https://www.terraform.io/downloads.html) >= 0.13.x
-	[Go](https://golang.org/doc/install) >= 1.17

## Using the provider

TODO: add link to terraform registry

## Building and testing the provider

First, build and install the provider.

```shell
make install
```

Then, run the follwing command in the workspace you want to test it.

```shell
rm .terraform.lock.hc && terraform init
```
