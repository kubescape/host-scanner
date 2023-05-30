//go:build kind

package e2e_test

import (
	"github.com/kubescape/host-scanner/sensor"
)

var cloudProviderInfo = &sensor.CloudProviderInfo{
	ProviderMetaDataAPIAccess: true,
}
