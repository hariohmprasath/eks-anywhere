package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type deleteClusterOptions struct {
	clusterOptions
	wConfig      string
	forceCleanup bool
}

var dc = &deleteClusterOptions{}

var deleteClusterCmd = &cobra.Command{
	Use:          "cluster (<cluster-name>|-f <config-file>)",
	Short:        "Workload cluster",
	Long:         "This command is used to delete workload clusters created by eksctl anywhere",
	PreRunE:      preRunDeleteCluster,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := dc.validate(cmd.Context(), args); err != nil {
			return err
		}
		if err := dc.deleteCluster(cmd.Context()); err != nil {
			return fmt.Errorf("failed to delete cluster: %v", err)
		}
		return nil
	},
}

func preRunDeleteCluster(cmd *cobra.Command, args []string) error {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		err := viper.BindPFlag(flag.Name, flag)
		if err != nil {
			log.Fatalf("Error initializing flags: %v", err)
		}
	})
	return nil
}

func init() {
	deleteCmd.AddCommand(deleteClusterCmd)
	deleteClusterCmd.Flags().StringVarP(&dc.fileName, "filename", "f", "", "Filename that contains EKS-A cluster configuration, required if <cluster-name> is not provided")
	deleteClusterCmd.Flags().StringVarP(&dc.wConfig, "w-config", "w", "", "Kubeconfig file to use when deleting a workload cluster")
	deleteClusterCmd.Flags().BoolVar(&dc.forceCleanup, "force-cleanup", false, "Force deletion of previously created bootstrap cluster")
	deleteClusterCmd.Flags().StringVar(&dc.managementKubeconfig, "kubeconfig", "", "kubeconfig file pointing to a management cluster")
}

func (dc *deleteClusterOptions) validate(ctx context.Context, args []string) error {
	if dc.fileName == "" {
		clusterName, err := validations.ValidateClusterNameArg(args)
		if err != nil {
			return fmt.Errorf("please provide either a valid <cluster-name> or -f <config-file>")
		}
		filename := fmt.Sprintf("%[1]s/%[1]s-eks-a-cluster.yaml", clusterName)
		if !validations.FileExists(filename) {
			return fmt.Errorf("clusterconfig file %s for cluster: %s not found, please provide the clusterconfig path manually using -f <config-file>", filename, clusterName)
		}
		dc.fileName = filename
	}
	clusterConfig, err := commonValidation(ctx, dc.fileName)
	if err != nil {
		return err
	}
	if !validations.KubeConfigExists(clusterConfig.Name, clusterConfig.Name, dc.wConfig, kubeconfigPattern) {
		return fmt.Errorf("KubeConfig doesn't exists for cluster %s", clusterConfig.Name)
	}
	return nil
}

func (dc *deleteClusterOptions) deleteCluster(ctx context.Context) error {
	clusterSpec, err := newClusterSpec(dc.clusterOptions)
	if err != nil {
		return fmt.Errorf("unable to get cluster config from file: %v", err)
	}

	deps, err := dependencies.ForSpec(ctx, clusterSpec).
		WithBootstrapper().
		WithClusterManager().
		WithProvider(dc.fileName, clusterSpec.Cluster, cc.skipIpCheck).
		WithFluxAddonClient(ctx, clusterSpec.Cluster, clusterSpec.GitOpsConfig).
		WithWriter().
		Build()
	if err != nil {
		return err
	}

	deleteCluster := workflows.NewDelete(
		deps.Bootstrapper,
		deps.Provider,
		deps.ClusterManager,
		deps.FluxAddonClient,
	)

	var cluster *types.Cluster
	if clusterSpec.ManagementCluster == nil {
		cluster = &types.Cluster{
			Name:           clusterSpec.Name,
			KubeconfigFile: uc.kubeConfig(clusterSpec.Name),
		}
	} else {
		cluster = &types.Cluster{
			Name:           clusterSpec.Name,
			KubeconfigFile: clusterSpec.ManagementCluster.KubeconfigFile,
		}
	}

	err = deleteCluster.Run(ctx, cluster, clusterSpec, dc.forceCleanup, dc.managementKubeconfig)
	if err == nil {
		deps.Writer.CleanUp()
	}
	return err
}
