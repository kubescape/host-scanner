package sensor

import (
	"fmt"

	"go.uber.org/zap"
)

const (
	kubeProxyExe = "kube-proxy"
)

// KubeProxyInfo holds information about kube-proxy process
type KubeProxyInfo struct {
	// Information about the kubeconfig file of kube-proxy
	KubeConfigFile *FileInfo `json:"kubeConfigFile,omitempty"`

	// Raw cmd line of kubelet process
	CmdLine string `json:"cmdLine"`
}

// SenseKubeProxyInfo return `KubeProxyInfo`
func SenseKubeProxyInfo() (*KubeProxyInfo, error) {
	ret := KubeProxyInfo{}

	// Get process
	proc, err := LocateProcessByExecSuffix(kubeProxyExe)
	if err != nil {
		return &ret, fmt.Errorf("failed to locate kube-proxy process: %w", err)
	}

	// kubeconfig
	kubeConfigPath, ok := proc.GetArg("--config")
	if ok {
		kubeConfigInfo, err := makeContaineredFileInfo(kubeConfigPath, false, proc)
		ret.KubeConfigFile = kubeConfigInfo
		if err != nil {
			zap.L().Debug("SenseKubeProxyInfo failed to MakeFileInfo for kube-proxy kubeconfig",
				zap.String("path", kubeConfigPath),
				zap.Error(err),
			)
		}
	}

	// cmd line
	ret.CmdLine = proc.RawCmd()

	return &ret, nil
}
