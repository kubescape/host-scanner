package sensor

import (
	"os"
	"testing"
)

// this test should run only on real k8s pod.
// TODO: add to system-tests. Check if this data arrived to the backend
func TestLocateKubelet(t *testing.T) {
	res, err := LocateKubeletProcess()
	if err != nil {
		t.Errorf("failed to LocateKubeletProcess: %v", err)
	}
	if res.PID < 1 {
		t.Errorf("failed to LocateKubeletProcess: %v", res)
	}
}

func Test_kubeletExtractCAFileFromConf(t *testing.T) {

	tests := []struct {
		name     string
		dataPath string
		want     string
		wantErr  bool
	}{
		{
			name:     "simple exist",
			dataPath: "testdata/clientCAKubeletConf.yaml",
			want:     "/var/lib/minikube/certs/ca.crt",
			wantErr:  false,
		},
		{
			name:     "simple not exist",
			dataPath: "testdata/clientCAKubeletConf_2.yaml",
			want:     "",
			wantErr:  false,
		},
		{
			name:     "simple not exist 2",
			dataPath: "testdata/clientCAKubeletConf_3.yaml",
			want:     "",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(tt.dataPath)
			if err != nil {
				t.Errorf("kubeletExtractCAFileFromConf() failed to read testdata. %v", err)
				return
			}
			got, err := kubeletExtractCAFileFromConf(data)
			if (err != nil) != tt.wantErr {
				t.Errorf("kubeletExtractCAFileFromConf() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("kubeletExtractCAFileFromConf() = %v, want %v", got, tt.want)
			}
		})
	}
}
