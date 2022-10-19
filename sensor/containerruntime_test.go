package sensor

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
				if tt.wantErr {
					fmt.Println(err)
				} else {
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
				if tt.wantErr {
					fmt.Println(err)
				} else {
					assert.NoError(t, err)
				}

			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, CNIConfigDir)
			}

		})
	}

}

