package sensor

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getContainerdCNIPaths(t *testing.T) {
	uid_tests := []struct {
		name        string
		path        string
		expectedRes CNIPaths
		wantErr     bool
	}{
		{
			name:        "fileexists_paramsexist",
			path:        "testCNI/containerd.toml",
			expectedRes: CNIPaths{Conf_dir: "/etc/cni/net.mk", Bin_dirs: []string{"/opt/cni/bin"}},
			wantErr:     false,
		},
		{
			name:        "file_not_exit",
			path:        "testCNI/bla.toml",
			expectedRes: CNIPaths{},
			wantErr:     true,
		},
		{
			name:        "fileexists_noparams",
			path:        "testCNI/containerd_noparams.toml",
			expectedRes: CNIPaths{},
			wantErr:     false,
		},
	}

	crp, err := getContainerRuntimeProperties(CONTAINERD_CONTAINER_RUNTIME_NAME)
	assert.NoError(t, err)
	cr, err := NewContainerRuntime(*crp, "")
	for _, tt := range uid_tests {
		t.Run(tt.name, func(t *testing.T) {
			cni_paths, err := cr.getCNIPathsFromConfig(tt.path)

			if err != nil {
				if tt.wantErr {
					fmt.Println(err)
				} else {
					assert.NoError(t, err)
				}

			} else {
				assert.NoError(t, err)
				assert.Equal(t, &tt.expectedRes, cni_paths)
			}

		})
	}

}

func Test_getCrioCNIPaths(t *testing.T) {
	uid_tests := []struct {
		name        string
		path        string
		expectedRes CNIPaths
		wantErr     bool
	}{
		{
			name:        "fileexists_paramsexist",
			path:        "testCNI/crio.conf",
			expectedRes: CNIPaths{Conf_dir: "/etc/cni/net.d/", Bin_dirs: []string{"/opt/cni/bin/", "/rr/ff"}},
			wantErr:     false,
		},
		{
			name:        "file_not_exit",
			path:        "testCNI/bla.toml",
			expectedRes: CNIPaths{},
			wantErr:     true,
		},
		{
			name:        "fileexists_noparams",
			path:        "testCNI/crio_noparams.conf",
			expectedRes: CNIPaths{},
			wantErr:     false,
		},
	}

	crp, err := getContainerRuntimeProperties(CRIO_CONTAINER_RUNTIME_NAME)
	assert.NoError(t, err)
	cr, err := NewContainerRuntime(*crp, "")
	for _, tt := range uid_tests {
		t.Run(tt.name, func(t *testing.T) {
			cni_paths, err := cr.getCNIPathsFromConfig(tt.path)

			if err != nil {
				if tt.wantErr {
					fmt.Println(err)
				} else {
					assert.NoError(t, err)
				}

			} else {
				assert.NoError(t, err)
				assert.Equal(t, &tt.expectedRes, cni_paths)
			}

		})
	}

}
