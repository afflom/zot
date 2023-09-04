// main.go
// This is the main entry point for the application.

package main

import (
	"github.com/emporous/uor-zot/iac/config"
    "github.com/emporous/uor-zot/iac/kubernetes"
	p "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	cfg "github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	p.Run(func(ctx *p.Context) error {
		// Load configurations
		configCtx := cfg.New(ctx, "")
		app := config.GetAppConfig(configCtx)
		kubeConfig, kubeContext := config.GetKubeConfig(configCtx)
		replicas := config.GetReplicas(configCtx)

		// Create Kubernetes provider
		provider, err := kubernetes.CreateProvider(ctx, kubeConfig, kubeContext)
		if err != nil {
			return err
		}

		// Create Kubernetes Deployment
		deployment, err := kubernetes.CreateDeployment(ctx, app, provider, replicas)
		if err != nil {
			return err
		}

		// Export deployment name
		ctx.Export("name", p.String(deployment.Metadata.ElementType().Name()))
		return nil
	})
}
