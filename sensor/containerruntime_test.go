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
			path:        "testdata/testCNI/containerd.toml",
			expectedRes: CNIPaths{Conf_dir: "/etc/cni/net.mk", Bin_dirs: []string{"/opt/cni/bin"}},
			wantErr:     false,
		},
		{
			name:        "file_not_exit",
			path:        "testdata/testCNI/bla.toml",
			expectedRes: CNIPaths{},
			wantErr:     true,
		},
		{
			name:        "fileexists_noparams",
			path:        "testdata/testCNI/containerd_noparams.toml",
			expectedRes: CNIPaths{},
			wantErr:     false,
		},
	}

	crp, err := getContainerRuntimeProperties(containerdContainerRuntimeName)
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
			path:        "testdata/testCNI/crio.conf",
			expectedRes: CNIPaths{Conf_dir: "/etc/cni/net.d/", Bin_dirs: []string{"/opt/cni/bin/", "/rr/ff"}},
			wantErr:     false,
		},
		{
			name:        "file_not_exit",
			path:        "testdata/testCNI/bla.toml",
			expectedRes: CNIPaths{},
			wantErr:     true,
		},
		{
			name:        "fileexists_noparams",
			path:        "testdata/testCNI/crio_noparams.conf",
			expectedRes: CNIPaths{},
			wantErr:     false,
		},
	}

	crp, err := getContainerRuntimeProperties(crioContainerRuntimeName)
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

func Test_getCNIPathsFromPaths_crio(t *testing.T) {
	uid_tests := []struct {
		name        string
		path        string
		expectedRes CNIPaths
		wantErr     bool
	}{
		{
			name:        "crio_withparams",
			path:        "testdata/testCNI/crio.d",
			expectedRes: CNIPaths{Conf_dir: "/etc/cni/net.d/03", Bin_dirs: []string{"/opt/cni/bin/02", "/rr/ff/02"}},
			wantErr:     false,
		},
		{
			name:        "crio_noparams",
			path:        "testdata/testCNI/crio.d_noparams",
			expectedRes: CNIPaths{},
			wantErr:     true,
		},
	}

	for _, tt := range uid_tests {
		t.Run(tt.name, func(t *testing.T) {
			config_paths, err := makeConfigFilesList(tt.path)
			assert.NoError(t, err)

			cni_paths, err := getCNIPathsFromConfigPaths(config_paths, parseCNIPathsFromConfig_crio)

			// if cni_paths != nil {
			// 	fmt.Printf("%+v\n", cni_paths)
			// }

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
