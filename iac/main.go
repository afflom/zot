// main.go
// This is the main entry point for the application.
package main

import (
	"github.com/emporous/uor-zot/iac/kind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Initialize Kind cluster arguments
		kindArgs := &kind.KindClusterArgs{
			ClusterName: "my-cluster",
			WorkingDir:  "./kind", // where your kind.yaml is located
		}

		// Create a new Kind cluster
		_, err := kind.NewKindCluster(ctx, "my-kind-cluster", kindArgs)
		if err != nil {
			return err
		}

		return nil
	})
}
