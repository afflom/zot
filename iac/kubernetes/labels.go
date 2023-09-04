// kubernetes/labels.go
// This file handles the creation of labels and resource options for Kubernetes resources.

package kubernetes

import (
	"github.com/emporous/uor-zot/iac/types"
	k8s "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	p "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// CreateAppLabels creates a set of labels for an application.
func CreateAppLabels(app types.App) p.StringMap {
	return p.StringMap{
		"app":       p.String(app.Name),
		"managedBy": p.String("Pulumi"),
		"version":   p.String(app.Version),
	}
}

// CreateResourceOptions sets up resource options for Kubernetes resources.
func CreateResourceOptions(provider *k8s.Provider) []p.ResourceOption {
	options := []p.ResourceOption{p.DeleteBeforeReplace(true)}
	if provider != nil {
		options = append(options, p.Provider(provider))
	}
	return options
}
