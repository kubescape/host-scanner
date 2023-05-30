//go:build kind

package e2e_test

import (
	"github.com/kubescape/host-scanner/sensor"
	ds "github.com/kubescape/host-scanner/sensor/datastructures"
)

var kubeletInfo = &sensor.KubeletInfo{
	ServiceFiles: []ds.FileInfo{
		{
			Ownership: &ds.FileOwnership{
				Err:       "",
				UID:       0,
				GID:       0,
				Username:  "root",
				Groupname: "root",
			},
			Path:        "/etc/systemd/system/kubelet.service.d/10-kubeadm.conf",
			Permissions: 420,
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
		Path:        "/etc/kubernetes/kubelet.conf",
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
		Path:        "/etc/kubernetes/pki/ca.crt",
		Permissions: 420,
	},
}
