//go:build aks

package e2e_test

import (
	ds "github.com/kubescape/host-scanner/sensor/datastructures"
)

var linuxSecurityHardening = &ds.LinuxSecurityHardeningStatus{
	AppArmor: "unloaded",
	SeLinux:  "not found",
}
