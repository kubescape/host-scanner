package sensor

import (
	"fmt"

	ds "github.com/kubescape/host-scanner/sensor/datastructures"
	"github.com/kubescape/host-scanner/sensor/internal/utils"
	"go.uber.org/zap"
)

const (
	kubeProxyExe = "kube-proxy"
)

// KubeProxyInfo holds information about kube-proxy process
type KubeProxyInfo struct {
	// Information about the kubeconfig file of kube-proxy
	KubeConfigFile *ds.FileInfo `json:"kubeConfigFile,omitempty"`

	// Raw cmd line of kubelet process
	CmdLine string `json:"cmdLine"`
}

// SenseKubeProxyInfo return `KubeProxyInfo`
func SenseKubeProxyInfo() (*KubeProxyInfo, error) {
	ret := KubeProxyInfo{}

	// Get process
	proc, err := utils.LocateProcessByExecSuffix(kubeProxyExe)
	if err != nil {
		return &ret, fmt.Errorf("failed to locate kube-proxy process: %w", err)
	}

	// kubeconfig
	kubeConfigPath, ok := proc.GetArg(kubeConfigArgName)
	if ok {
		ret.KubeConfigFile = makeContaineredFileInfoVerbose(proc, kubeConfigPath, false,
			zap.String("in", "SenseKubeProxyInfo"),
		)
	}

	// cmd line
	ret.CmdLine = proc.RawCmd()

	return &ret, nil
}
