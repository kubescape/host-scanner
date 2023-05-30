//go:build aks

package e2e_test

import (
	"github.com/kubescape/host-scanner/sensor"
	ds "github.com/kubescape/host-scanner/sensor/datastructures"
)

var kubeProxyInfo = &sensor.KubeProxyInfo{
	KubeConfigFile: &ds.FileInfo{
		Ownership:   &ds.FileOwnership{Err: "", UID: 0, GID: 0, Username: "root", Groupname: "root"},
		Path:        "/var/lib/kubelet/kubeconfig",
		Content:     nil,
		Permissions: 384,
	},
}
