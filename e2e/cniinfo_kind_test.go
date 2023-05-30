//go:build kind

package e2e_test

import (
	"github.com/kubescape/host-scanner/sensor"
	ds "github.com/kubescape/host-scanner/sensor/datastructures"
)

var cniInfo = &sensor.CNIInfo{
	CNIConfigFiles: []*ds.FileInfo{
		{
			Ownership: &ds.FileOwnership{
				Err:       "",
				UID:       0,
				GID:       0,
				Username:  "root",
				Groupname: "root",
			},
			Path:        "/etc/cni/net.d/10-kindnet.conflist",
			Permissions: 420,
		},
		{
			Ownership: &ds.FileOwnership{
				Err:       "",
				UID:       0,
				GID:       0,
				Username:  "root",
				Groupname: "root",
			},
			Path:        "/etc/cni/net.d",
			Permissions: 448,
		},
	},
	CNINames: []string{"Kindnet"},
}
