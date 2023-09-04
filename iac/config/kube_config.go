// config/kube_config.go
// This file handles the retrieval of Kubernetes configuration settings.

package config

import (
	"os"
	"path/filepath"

	cfg "github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// GetKubeConfig fetches Kubernetes configuration settings.
func GetKubeConfig(config *cfg.Config) (string, string) {
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
