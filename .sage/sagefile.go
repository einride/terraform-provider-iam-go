package main

import (
	"context"
	"os"
	"path/filepath"

	"go.einride.tech/sage/sg"
	"go.einride.tech/sage/sgtool"
	"go.einride.tech/sage/tools/sgconvco"
	"go.einride.tech/sage/tools/sggit"
	"go.einride.tech/sage/tools/sggo"
	"go.einride.tech/sage/tools/sggolangcilint"
	"go.einride.tech/sage/tools/sggoreview"
	"go.einride.tech/sage/tools/sgmarkdownfmt"
	"go.einride.tech/sage/tools/sgterraform"
)

const (
	hostname  = "hashicorp.com"
	namespace = "einride"
	pkgName   = "iamgo"
	name      = "iam-go"
	version   = "0.1.0"
	binary    = "terraform-provider-" + name + "_v" + version
	osArch    = "linux_amd64"
)

func main() {
	sg.GenerateMakefiles(
		sg.Makefile{
			Path:          sg.FromGitRoot("Makefile"),
			DefaultTarget: All,
		},
	)
}

func All(ctx context.Context) error {
	sg.Deps(ctx, ConvcoCheck, GoLint, GoReview, GoTest)
	sg.SerialDeps(ctx, TfDocPlugin, FormatMarkdown, GoModTidy, GitVerifyNoDiff)
	return nil
}

func LocalInstall(ctx context.Context) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	binaryDir := filepath.Join(homeDir, ".terraform.d/plugins", hostname, namespace, name, version, osArch)
	if err := os.MkdirAll(
		binaryDir,
		0o755,
	); err != nil {
		return err
	}
	sg.Logger(ctx).Printf("installing provider to %s...\n", binaryDir)
	return sg.Command(ctx, "go", "build", "-o", filepath.Join(binaryDir, binary)).Run()
}

func TfDocPlugin(ctx context.Context) error {
	sg.Deps(ctx, sgterraform.PrepareCommand)
	sg.Logger(ctx).Println("generating tfdocs...")
	if err := os.RemoveAll(sg.FromGitRoot("docs")); err != nil {
		return err
	}
	gen, err := sgtool.GoInstall(ctx, "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs", "latest")
	if err != nil {
		return err
	}
	return sg.Command(ctx, gen).Run()
}

func GoModTidy(ctx context.Context) error {
	sg.Logger(ctx).Println("tidying Go module files...")
	return sg.Command(ctx, "go", "mod", "tidy", "-v").Run()
}

func GoTest(ctx context.Context) error {
	sg.Logger(ctx).Println("running Go tests...")
	return sggo.TestCommand(ctx).Run()
}

func GoReview(ctx context.Context) error {
	sg.Logger(ctx).Println("reviewing Go files...")
	return sggoreview.Command(ctx, "-c", "1", "./...").Run()
}

func GoLint(ctx context.Context) error {
	sg.Logger(ctx).Println("linting Go files...")
	return sggolangcilint.Run(ctx)
}

func FormatMarkdown(ctx context.Context) error {
	sg.Logger(ctx).Println("formatting Markdown files...")
	return sgmarkdownfmt.Command(ctx, "-w", ".").Run()
}

func ConvcoCheck(ctx context.Context) error {
	sg.Logger(ctx).Println("checking git commits...")
	return sgconvco.Command(ctx, "check", "origin/master..HEAD").Run()
}

func GitVerifyNoDiff(ctx context.Context) error {
	sg.Logger(ctx).Println("verifying that git has no diff...")
	return sggit.VerifyNoDiff(ctx)
}
