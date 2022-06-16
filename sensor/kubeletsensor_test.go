package sensor

import (
	"testing"
)

func TestLocateKubelet(t *testing.T) {
	// res, err := LocateKubeletProcess()
	// if err != nil {
	// 	t.Errorf("failed to LocateKubeletProcess: %v", err)
	// }
	// if res.PID < 1 {
	// 	t.Errorf("failed to LocateKubeletProcess: %v", res)
	// }
}

const clientCAKubeletConf string = `apiVersion: kubelet.config.k8s.io/v1beta1
authentication:
  anonymous:
    enabled: false
  webhook:
    cacheTTL: 0s
    enabled: true
  x509:
    clientCAFile: /var/lib/minikube/certs/ca.crt
evictionHard:
  imagefs.available: 0%
  nodefs.available: 0%
  nodefs.inodesFree: 0%
logging:
  flushFrequency: 0
  options:
    json:
      infoBufferSize: "0"
  verbosity: 0
`
const clientCAKubeletConf2 string = `apiVersion: kubelet.config.k8s.io/v1beta1
authentication:
  anonymous:
    enabled: false
  webhook:
    cacheTTL: 0s
    enabled: true
evictionHard:
  imagefs.available: 0%
  nodefs.available: 0%
  nodefs.inodesFree: 0%
logging:
  flushFrequency: 0
  options:
    json:
      infoBufferSize: "0"
  verbosity: 0
`

const clientCAKubeletConf3 string = `apiVersion: kubelet.config.k8s.io/v1beta1
evictionHard:
  imagefs.available: 0%
  nodefs.available: 0%
  nodefs.inodesFree: 0%
logging:
  flushFrequency: 0
  options:
    json:
      infoBufferSize: "0"
  verbosity: 0
`

func Test_kubeletExtractCAFileFromConf(t *testing.T) {
	type args struct {
		content []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "simple exist",
			args: args{
				content: []byte(clientCAKubeletConf),
			},
			want:    "/var/lib/minikube/certs/ca.crt",
			wantErr: false,
		},
		{
			name: "simple not exist",
			args: args{
				content: []byte(clientCAKubeletConf2),
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "simple not exist 2",
			args: args{
				content: []byte(clientCAKubeletConf3),
			},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := kubeletExtractCAFileFromConf(tt.args.content)
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
