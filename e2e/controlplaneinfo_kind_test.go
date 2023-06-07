//go:build kind

package e2e_test

import (
	sensor "github.com/kubescape/host-scanner/sensor"
	ds "github.com/kubescape/host-scanner/sensor/datastructures"
)

var pkiFiles = &sensor.ControlPlaneInfo{
	PKIDIr: &ds.FileInfo{Path: "/etc/kubernetes/pki"},
	PKIFiles: []*ds.FileInfo{
		{
			Path: "/etc/kubernetes/pki/etcd/peer.crt",
		},
		{
			Path: "/etc/kubernetes/pki/etcd/healthcheck-client.crt",
		},
		{
			Path: "/etc/kubernetes/pki/etcd/ca.crt",
		},
		{
			Path: "/etc/kubernetes/pki/etcd/ca.key",
		},
		{
			Path: "/etc/kubernetes/pki/etcd/server.crt",
		},
		{
			Path: "/etc/kubernetes/pki/etcd/server.key",
		},
		{
			Path: "/etc/kubernetes/pki/etcd/healthcheck-client.key",
		},
		{
			Path: "/etc/kubernetes/pki/etcd/peer.key",
		},
		{
			Path: "/etc/kubernetes/pki/ca.crt",
		},
		{
			Path: "/etc/kubernetes/pki/apiserver.crt",
		},
		{
			Path: "/etc/kubernetes/pki/ca.key",
		},
		{
			Path: "/etc/kubernetes/pki/front-proxy-client.crt",
		},
		{
			Path: "/etc/kubernetes/pki/front-proxy-ca.key",
		},
		{
			Path: "/etc/kubernetes/pki/apiserver-etcd-client.crt",
		},
		{
			Path: "/etc/kubernetes/pki/front-proxy-ca.crt",
		},
		{
			Path: "/etc/kubernetes/pki/apiserver-kubelet-client.crt",
		},
		{
			Path: "/etc/kubernetes/pki/apiserver-etcd-client.key",
		},
		{
			Path: "/etc/kubernetes/pki/apiserver.key",
		},
		{
			Path: "/etc/kubernetes/pki/front-proxy-client.key",
		},
		{
			Path: "/etc/kubernetes/pki/sa.key",
		},
		{
			Path: "/etc/kubernetes/pki/sa.pub",
		},
		{
			Path: "/etc/kubernetes/pki/apiserver-kubelet-client.key",
		},
	},
}

var apiServerInfo = &sensor.ControlPlaneInfo{
	APIServerInfo: &sensor.ApiServerInfo{
		K8sProcessInfo: &sensor.K8sProcessInfo{
			SpecsFile: &ds.FileInfo{
				Ownership: &ds.FileOwnership{
					Err:       "",
					UID:       0,
					GID:       0,
					Username:  "root",
					Groupname: "root",
				},
				Path:        "/etc/kubernetes/manifests/kube-apiserver.yaml",
				Permissions: 384,
			},
		},
	},
}

var controllerManagerInfo = &sensor.ControlPlaneInfo{
	ControllerManagerInfo: &sensor.K8sProcessInfo{
		SpecsFile: &ds.FileInfo{
			Ownership: &ds.FileOwnership{
				Err:       "",
				UID:       0,
				GID:       0,
				Username:  "root",
				Groupname: "root",
			},
			Path:        "/etc/kubernetes/manifests/kube-controller-manager.yaml",
			Permissions: 384,
		},
		ConfigFile: &ds.FileInfo{
			Ownership: &ds.FileOwnership{
				Err:       "",
				UID:       0,
				GID:       0,
				Username:  "root",
				Groupname: "root",
			},
			Path:        "/etc/kubernetes/controller-manager.conf",
			Permissions: 384,
		},
	},
}

var schedulerInfo = &sensor.ControlPlaneInfo{
	SchedulerInfo: &sensor.K8sProcessInfo{
		SpecsFile: &ds.FileInfo{
			Ownership: &ds.FileOwnership{
				Err:       "",
				UID:       0,
				GID:       0,
				Username:  "root",
				Groupname: "root",
			},
			Path:        "/etc/kubernetes/manifests/kube-scheduler.yaml",
			Permissions: 384,
		},
		ConfigFile: &ds.FileInfo{
			Ownership: &ds.FileOwnership{
				Err:       "",
				UID:       0,
				GID:       0,
				Username:  "root",
				Groupname: "root",
			},
			Path:        "/etc/kubernetes/scheduler.conf",
			Permissions: 384,
		},
	},
}

var etcdConfigFile = &sensor.ControlPlaneInfo{
	EtcdConfigFile: &ds.FileInfo{
		Ownership: &ds.FileOwnership{
			Err:       "",
			UID:       0,
			GID:       0,
			Username:  "root",
			Groupname: "root",
		},
		Path:        "/etc/kubernetes/manifests/etcd.yaml",
		Permissions: 384,
	},
	EtcdDataDir: &ds.FileInfo{
		Ownership: &ds.FileOwnership{
			Err:       "",
			UID:       0,
			GID:       0,
			Username:  "root",
			Groupname: "root",
		},
		Path:        "/var/lib/etcd",
		Permissions: 448,
	},
}

var adminConfigFile = &sensor.ControlPlaneInfo{
	AdminConfigFile: &ds.FileInfo{
		Ownership: &ds.FileOwnership{
			Err:       "",
			UID:       0,
			GID:       0,
			Username:  "root",
			Groupname: "root",
		},
		Path:        "/etc/kubernetes/admin.conf",
		Permissions: 384,
	},
}
