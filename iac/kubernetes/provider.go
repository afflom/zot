// kubernetes/provider.go
// This file handles the creation of a Kubernetes provider.

package kubernetes

import (
	p "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	k8s "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
)

// CreateProvider sets up a Kubernetes provider using provided configurations.
func CreateProvider(ctx *p.Context, kubeConfig, kubeContext string) (*k8s.Provider, error) {
	providerArgs := &k8s.ProviderArgs{}
	if kubeConfig != "" {
		providerArgs.Kubeconfig = p.String(kubeConfig)
	}
	if kubeContext != "" {
		providerArgs.Context = p.String(kubeContext)
	}
	return k8s.NewProvider(ctx, "kubeconfig", providerArgs)
}
