//go:build aks

package e2e_test

import (
	"github.com/kubescape/host-scanner/sensor"
	ds "github.com/kubescape/host-scanner/sensor/datastructures"
)

var kubeletInfo = &sensor.KubeletInfo{
	ServiceFiles: []ds.FileInfo{
		{
			Ownership:   &ds.FileOwnership{Err: "", UID: 0, GID: 0, Username: "root", Groupname: "root"},
			Path:        "/etc/systemd/system/kubelet.service.d/10-cgroupv2.conf",
			Content:     nil,
			Permissions: 420,
		},
		{
			Ownership:   &ds.FileOwnership{Err: "", UID: 0, GID: 0, Username: "root", Groupname: "root"},
			Path:        "/etc/systemd/system/kubelet.service.d/10-container-runtime-flag.conf",
			Content:     nil,
			Permissions: 420,
		},
		{
			Ownership:   &ds.FileOwnership{Err: "", UID: 0, GID: 0, Username: "root", Groupname: "root"},
			Path:        "/etc/systemd/system/kubelet.service.d/10-containerd-base-flag.conf",
			Content:     nil,
			Permissions: 420,
		},
		{
			Ownership:   &ds.FileOwnership{Err: "", UID: 0, GID: 0, Username: "root", Groupname: "root"},
			Path:        "/etc/systemd/system/kubelet.service.d/10-tlsbootstrap.conf",
			Content:     nil,
			Permissions: 384,
		},
	},
	ConfigFile: &ds.FileInfo{
		Ownership: &ds.FileOwnership{
			Err:       "",
			UID:       0,
			GID:       0,
			Username:  "root",
			Groupname: "root",
		},
		Path:        "/var/lib/kubelet/config.yaml",
		Permissions: 420,
	},
	KubeConfigFile: &ds.FileInfo{
		Ownership: &ds.FileOwnership{
			Err:       "",
			UID:       0,
			GID:       0,
			Username:  "root",
			Groupname: "root",
		},
		Path:        "/var/lib/kubelet/kubeconfig",
		Permissions: 420,
	},
	ClientCAFile: &ds.FileInfo{
		Ownership: &ds.FileOwnership{
			Err:       "",
			UID:       0,
			GID:       0,
			Username:  "root",
			Groupname: "root",
		},
		Path:        "/etc/kubernetes/certs/ca.crt",
		Permissions: 384,
	},
}
