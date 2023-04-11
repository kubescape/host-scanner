package utils

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetCNIConfigPath(t *testing.T) {
	uid_tests := []struct {
		name     string
		process  string
		pid      int32
		expected string
	}{
		{
			name:     "kubelet_kind",
			process:  "/usr/bin/kubelet --bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --kubeconfig=/etc/kubernetes/kubelet.conf --config=/var/lib/kubelet/config.yaml --container-runtime=remote --container-runtime-endpoint=unix:///run/containerd/containerd.sock --node-ip=172.18.0.2 --node-labels= --pod-infra-container-image=registry.k8s.io/pause:3.8 --provider-id=kind://docker/cis-test/cis-test-control-plane --fail-swap-on=false --cgroup-root=/kubelet",
			pid:      15,
			expected: "/etc/cni/",
		},
		{
			name:     "kubelet_manual_installation",
			process:  "/usr/local/bin/kubelet --config=/var/lib/kubelet/kubelet-config.yaml --container-runtime=remote --container-runtime-endpoint=unix:///var/run/containerd/containerd.sock --image-pull-progress-deadline=2m --kubeconfig=/var/lib/kubelet/kubeconfig --network-plugin=cni --register-node=true --v=2",
			pid:      15,
			expected: "/etc/cni/",
		},
		{
			name:     "kubelet_manual_installation_with_custom_runtime_endpoint",
			process:  "/usr/local/bin/kubelet --config=/var/lib/kubelet/kubelet-config.yaml --container-runtime=remote --container-runtime-endpoint=unix:///var/run/containerd/containerd.sock --image-pull-progress-deadline=2m --kubeconfig=/var/lib/kubelet/kubeconfig --network-plugin=cni --container-runtime-endpoint=/run/containerd/containerd.sock",
			pid:      15,
			expected: "/run/containerd/",
		},
		{
			name:     "kubelet_manual_installation_with_custom_cni_dir",
			process:  "/usr/local/bin/kubelet --config=/var/lib/kubelet/kubelet-config.yaml --container-runtime=remote --container-runtime-endpoint=unix:///var/run/containerd/containerd.sock --image-pull-progress-deadline=2m --kubeconfig=/var/lib/kubelet/kubeconfig --network-plugin=cni --container-runtime-endpoint=/run/containerd/containerd.sock --cni-conf-dir=/var/lib/cni/",
			pid:      15,
			expected: "/var/lib/cni/",
		},
	}

	for _, tt := range uid_tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			// create ProcessDetails object to pass it to GetCNIConfigPath
			proc := &ProcessDetails{
				CmdLine: strings.Split(tt.process, " "),
				PID:     tt.pid,
			}
			cniConfigPath := GetCNIConfigPath(ctx, proc)

			if !assert.Equal(t, tt.expected, cniConfigPath) {
				t.Logf("%s has different output\n", tt.name)
			}
		})
	}
}

func Test_parseCNIPathsFromConfigContainerd(t *testing.T) {
	uid_tests := []struct {
		name        string
		path        string
		expectedRes string
		wantErr     bool
	}{
		{
			name:        "fileexists_paramsexist",
			path:        "testdata/testCNI/containerd.toml",
			expectedRes: "/etc/cni/net.mk",
			wantErr:     false,
		},
		{
			name:        "file_not_exit",
			path:        "testdata/testCNI/bla.toml",
			expectedRes: "",
			wantErr:     true,
		},
		{
			name:        "fileexists_noparams",
			path:        "testdata/testCNI/containerd_noparams.toml",
			expectedRes: "",
			wantErr:     false,
		},
	}

	for _, tt := range uid_tests {
		t.Run(tt.name, func(t *testing.T) {
			CNIConfigDir, err := parseCNIConfigDirFromConfigContainerd(tt.path)

			if err != nil {
				if !tt.wantErr {
					assert.NoError(t, err)
				}

			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, CNIConfigDir)
			}

		})
	}

}

func Test_parseCNIPathsFromConfigCrio(t *testing.T) {
	uid_tests := []struct {
		name        string
		path        string
		expectedRes string
		wantErr     bool
	}{
		{
			name:        "fileexists_paramsexist",
			path:        "testdata/testCNI/crio.conf",
			expectedRes: "/etc/cni/net.d/",
			wantErr:     false,
		},
		{
			name:        "file_not_exit",
			path:        "testdata/testCNI/bla.toml",
			expectedRes: "",
			wantErr:     true,
		},
		{
			name:        "fileexists_noparams",
			path:        "testdata/testCNI/crio_noparams.conf",
			expectedRes: "",
			wantErr:     false,
		},
	}

	for _, tt := range uid_tests {
		t.Run(tt.name, func(t *testing.T) {
			CNIConfigDir, err := parseCNIConfigDirFromConfigCrio(tt.path)

			if err != nil {
				if !tt.wantErr {
					assert.NoError(t, err)
				}

			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, CNIConfigDir)
			}

		})
	}

}
