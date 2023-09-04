// config/replicas.go
// This file handles the retrieval of the number of replicas for the application.

package config

import (
	pcfg "github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

const (
	DefaultReplicas = 1
)

// GetReplicas fetches the number of replicas set in the configuration.
func GetReplicas(config *pcfg.Config) int {
	replicas := config.GetInt("replicas")
	if replicas == 0 {
		replicas = DefaultReplicas
	}
	return replicas
}
