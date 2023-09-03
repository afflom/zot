package main

import (
	"github.com/pulumi/pulumi-docker/sdk/v3/go/docker"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Read general config
		conf := config.New(ctx, "")
		pushImage := conf.GetBool("pushImage")
		zotUser := conf.Get("zotUser")
		if zotUser == "" {
			zotUser = "defaultUser"
		}

		zotPasswd := conf.GetSecret("zotPasswd")
		var actualZotPasswd pulumi.StringOutput
		actualZotPasswd = zotPasswd.ApplyT(func(val string) string {
			if val == "" {
				return "defaultPassword"
			}
			return val
		}).(pulumi.StringOutput)

		var imageArgs *docker.ImageArgs

		// If pushImage is true, include registry details
		if pushImage {
			imageArgs = &docker.ImageArgs{
				Build: &docker.DockerBuildArgs{
					Context: pulumi.String("../build"),
				},
				ImageName: pulumi.String("emporous/uor-zot:dev"),
				Registry: &docker.ImageRegistryArgs{
					Server:   pulumi.String("ghcr.io"),
					Username: pulumi.String("emporous"),
					Password: pulumi.String("api-key-string"),
				},
			}
		} else {
			// Otherwise, just build the image locally
			imageArgs = &docker.ImageArgs{
				Build: &docker.DockerBuildArgs{
					Context: pulumi.String("../build"),
				},
			}
		}

		// Build and optionally publish the Docker image
		image, err := docker.NewImage(ctx, "ghcr.io/emporous/uor-zot", imageArgs)
		if err != nil {
			return err
		}

		// Create Kubernetes provider
		provider, err := kubernetes.NewProvider(ctx, "k8s-provider", &kubernetes.ProviderArgs{
			Kubeconfig: pulumi.String("~/.kube/config"),
		})
		if err != nil {
			return err
		}

		// Deploy Zot using the built Docker image
		_, err = kubernetes.NewDeployment(ctx, "zot-deployment", &kubernetes.DeploymentArgs{
			Metadata: &kubernetes.ObjectMetaArgs{
				Name: pulumi.String("zot"),
			},
			Spec: &kubernetes.DeploymentSpecArgs{
				Template: &kubernetes.PodTemplateSpecArgs{
					Metadata: &kubernetes.ObjectMetaArgs{
						Labels: pulumi.StringMap{"app": pulumi.String("zot")},
					},
					Spec: &kubernetes.PodSpecArgs{
						Containers: kubernetes.ContainerArray{
							&kubernetes.ContainerArgs{
								Name:  pulumi.String("zot"),
								Image: image.ImageName,
								Env: kubernetes.EnvVarArray{
									&kubernetes.EnvVarArgs{
										Name:  pulumi.String("ZOT_USER"),
										Value: pulumi.String(zotUser),
									},
									&kubernetes.EnvVarArgs{
										Name:      pulumi.String("ZOT_PASSWD"),
										ValueFrom: &kubernetes.EnvVarSourceArgs{SecretKeyRef: &kubernetes.SecretKeySelectorArgs{Key: pulumi.String("password"), Name: actualZotPasswd}},
									},
								},
							},
						},
					},
				},
			},
		}, pulumi.Provider(provider))

		if err != nil {
			return err
		}

		return nil
	})
}
