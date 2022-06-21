package sensor

import (
	"fmt"
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
	kubeConfigPath, ok := proc.GetArg(kubeConfigArgName)
	if ok {
		procRelativPath := proc.ContainredPath(kubeConfigPath)
		kubeConfigInfo, err := MakeFileInfo(procRelativPath, false)
		if err == nil {
			kubeConfigInfo.Path = kubeConfigPath
			ret.KubeConfigFile = kubeConfigInfo
		}
	}

	// cmd line
	ret.CmdLine = proc.RawCmd()

	return &ret, nil
}
