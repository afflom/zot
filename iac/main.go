package main

import (
	"fmt"
	"os"
	"path/filepath"

	k8s "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	apps "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	core "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	meta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	p "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	pcfg "github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

const (
	DefaultAppName    = "nginx"
	DefaultAppImage   = "docker.io/library/nginx"
	DefaultAppVersion = "latest"
	DefaultReplicas   = 1
	DefaultKubeConfig = "~/.kube/config"
)

type App struct {
	Name    string
	Image   string
	Version string
}

func getKubeConfig(config *pcfg.Config) (string, string) {
	kubeConfig := config.Get("kubeConfig")
	if kubeConfig == "" {
		kubeConfig = os.Getenv("KUBECONFIG")
	}
	if kubeConfig == "" {
		kubeConfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}
	kubeContext := config.Get("kubeContext")
	return kubeConfig, kubeContext
}

func getReplicas(config *pcfg.Config) int {
	replicas := config.GetInt("replicas")
	if replicas == 0 {
		replicas = DefaultReplicas
	}
	return replicas
}

func getAppConfig(config *pcfg.Config) App {
	return App{
		Name:    IfEmpty(config.Get("appName"), DefaultAppName),
		Image:   IfEmpty(config.Get("appImage"), DefaultAppImage),
		Version: IfEmpty(config.Get("appVersion"), DefaultAppVersion),
	}
}

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

func CreateDeploymentArgs(app App, appLabels p.StringMap, replicas int) *apps.DeploymentArgs {
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

func CreateDeployment(ctx *p.Context, app App, provider *k8s.Provider, replicas int) (*apps.Deployment, error) {
	appLabels := CreateAppLabels(app)
	deploymentArgs := CreateDeploymentArgs(app, appLabels, replicas)
	options := CreateResourceOptions(provider)
	return apps.NewDeployment(ctx, app.Name, deploymentArgs, options...)
}

func CreateAppLabels(app App) p.StringMap {
	return p.StringMap{
		"app":       p.String(app.Name),
		"managedBy": p.String("Pulumi"),
		"version":   p.String(app.Version),
	}
}

func CreateResourceOptions(provider *k8s.Provider) []p.ResourceOption {
	options := []p.ResourceOption{p.DeleteBeforeReplace(true)}
	if provider != nil {
		options = append(options, p.Provider(provider))
	}
	return options
}

func IfEmpty(val, defaultVal string) string {
	if val == "" {
		return defaultVal
	}
	return val
}

func main() {
	p.Run(func(ctx *p.Context) error {
		config := pcfg.New(ctx, "")
		app := getAppConfig(config)
		kubeConfig, kubeContext := getKubeConfig(config)
		replicas := getReplicas(config)

		provider, err := CreateProvider(ctx, kubeConfig, kubeContext)
		if err != nil {
			return err
		}

		deployment, err := CreateDeployment(ctx, app, provider, replicas)
		if err != nil {
			return err
		}

		ctx.Export("name", p.String(deployment.Metadata.ElementType().Name()))
		return nil
	})
}
