// config/app_config.go
// This file handles the retrieval of application-specific configuration settings.

package config

import (
	"github.com/emporous/uor-zot/iac/helpers"
	"github.com/emporous/uor-zot/iac/types"
	pcfg "github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

const (
	DefaultAppName    = "zot"
	DefaultAppImage   = "docker.io/library/nginx"
	DefaultAppVersion = "latest"
)

// GetAppConfig fetches application configuration settings.
func GetAppConfig(config *pcfg.Config) types.App {
	return types.App{
		Name:    helpers.IfEmpty(config.Get("appName"), DefaultAppName),
		Image:   helpers.IfEmpty(config.Get("appImage"), DefaultAppImage),
		Version: helpers.IfEmpty(config.Get("appVersion"), DefaultAppVersion),
	}
}
