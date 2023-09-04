// kubernetes/deployment.go
// This file handles the creation of Kubernetes deployments.

package kubernetes

import (
	"fmt"

	"github.com/emporous/uor-zot/iac/types"
	k8s "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	apps "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	core "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	meta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	p "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// CreateDeployment sets up a Kubernetes deployment based on provided application and replica configs.
func CreateDeployment(ctx *p.Context, app types.App, provider *k8s.Provider, replicas int) (*apps.Deployment, error) {
	appLabels := CreateAppLabels(app)
	deploymentArgs := CreateDeploymentArgs(app, appLabels, replicas)
	options := CreateResourceOptions(provider)
	return apps.NewDeployment(ctx, app.Name, deploymentArgs, options...)
}

// CreateDeploymentArgs prepares the arguments needed for a Kubernetes Deployment resource.
func CreateDeploymentArgs(app types.App, appLabels p.StringMap, replicas int) *apps.DeploymentArgs {
	image := fmt.Sprintf("%s:%s", app.Image, app.Version)
	return &apps.DeploymentArgs{
		Metadata: &meta.ObjectMetaArgs{
			Labels: appLabels,
		},
		Spec: &apps.DeploymentSpecArgs{
			Replicas: p.Int(replicas),
			Selector: &meta.LabelSelectorArgs{
				MatchLabels: appLabels,
			},
			Template: &core.PodTemplateSpecArgs{
				Metadata: &meta.ObjectMetaArgs{
					Labels: appLabels,
				},
				Spec: &core.PodSpecArgs{
					Containers: core.ContainerArray{
						&core.ContainerArgs{
							Name:  p.String(app.Name),
							Image: p.String(image),
						},
					},
				},
			},
		},
	}
}
