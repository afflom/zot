package kind

import (
	"github.com/pulumi/pulumi-command/sdk/v3/go/command/local"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func NewKindCluster(ctx *pulumi.Context,
	name string, args *KindClusterArgs, opts ...pulumi.ResourceOption) (*KindCluster, error) {
	if args == nil {
		args = &KindClusterArgs{}
	}

	component := &KindCluster{}
	err := ctx.RegisterComponentResource("my:kind:KindCluster", name, component, opts...)
	if err != nil {
		return nil, err
	}

	// Create a Pulumi local command resource for kind create cluster
	createCluster, err := local.NewCommand(ctx, name+"-createCluster", &local.CommandArgs{
		Create: pulumi.StringPtr("kind create cluster --name " + args.ClusterName),
		Dir:    pulumi.StringPtr(args.WorkingDir),
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	// Create a Pulumi local command resource for kind delete cluster
	deleteCluster, err := local.NewCommand(ctx, name+"-deleteCluster", &local.CommandArgs{
		Delete: pulumi.StringPtr("kind delete cluster --name " + args.ClusterName),
		Dir:    pulumi.StringPtr(args.WorkingDir),
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	// Populate component outputs
	component.ClusterName = pulumi.ToOutput(pulumi.String(args.ClusterName)).(pulumi.StringOutput)
	component.CreateStdout = createCluster.Stdout
	component.DeleteStdout = deleteCluster.Stdout

	// Register resource outputs
	err = ctx.RegisterResourceOutputs(component, pulumi.Map{
		"clusterName":  component.ClusterName,
		"createStdout": component.CreateStdout,
		"deleteStdout": component.DeleteStdout,
	})
	if err != nil {
		return nil, err
	}

	return component, nil
}

type KindClusterArgs struct {
	ClusterName string
	WorkingDir  string
}

type KindCluster struct {
	pulumi.ResourceState

	ClusterName  pulumi.StringOutput
	CreateStdout pulumi.StringOutput
	DeleteStdout pulumi.StringOutput
}
