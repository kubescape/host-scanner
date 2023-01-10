package sensor

import (
	"fmt"

	"go.uber.org/zap"

	ds "github.com/kubescape/host-scanner/sensor/datastructures"
	"github.com/kubescape/host-scanner/sensor/internal/utils"
)

// KubeProxyInfo holds information about kube-proxy process
type CNIInfo struct {
	CNIConfigFiles []*ds.FileInfo `json:"CNIConfigFiles,omitempty"`

	// The name of the running CNI
	CNIName string `json:"CNIName,omitempty"`
}

// SenseCNIInfo return `CNIInfo`
func SenseCNIInfo() (*CNIInfo, error) {
	var err error
	ret := CNIInfo{}

	// make cni config files
	CNIConfigInfo, err := makeCNIConfigFilesInfo()

	if err != nil {
		zap.L().Error("SenseCNIInfo", zap.Error(err))
	} else {
		ret.CNIConfigFiles = CNIConfigInfo
	}

	// get CNI name
	ret.CNIName = getCNIName()

	return &ret, nil
}

// makeCNIConfigFilesInfo - returns a list of FileInfos of cni config files.
func makeCNIConfigFilesInfo() ([]*ds.FileInfo, error) {
	// *** Start handling CNI Files
	kubeletProc, err := LocateKubeletProcess()
	if err != nil {
		return nil, err
	}

	CNIConfigDir := utils.GetCNIConfigPath(kubeletProc)

	if CNIConfigDir == "" {
		return nil, fmt.Errorf("no CNI Config dir found in getCNIConfigPath")
	}

	//Getting CNI config files
	CNIConfigInfo, err := makeHostDirFilesInfoVerbose(CNIConfigDir, true, nil, 0)

	if err != nil {
		return nil, fmt.Errorf("failed to makeHostDirFilesInfo for CNIConfigDir %s: %w", CNIConfigDir, err)
	}

	if len(CNIConfigInfo) == 0 {
		zap.L().Debug("SenseCNIInfo - no cni config files were found.",
			zap.String("path", CNIConfigDir))
	}

	return CNIConfigInfo, nil
}

// getCNIName - looking for CNI process and return CNI name, or empty if not found.
func getCNIName() string {
	supportedCNIs := []struct {
		name          string
		processSuffix string
	}{
		{"aws", "aws-k8s-agent"}, // aws VPC CNI agent
		// 'canal' CNI "sets up Calico to handle policy management and Flannel to manage the network itself". Therefore we will first
		// check "calico" (which supports network policies and indicates for either 'canal' or 'calico') and then flannel.
		{"Calico", "calico-node"},
		{"Flannel", "flanneld"},
		{"Cilium", "cilium-agent"},
		{"WeaveNet", "weave-net"},
	}

	for _, cni := range supportedCNIs {
		p, err := utils.LocateProcessByExecSuffix(cni.processSuffix)

		if p != nil {
			zap.L().Debug("CNI process found", zap.String("name", cni.name))
			return cni.name
		}

		if err != nil {
			zap.L().Error("getCNIName- Failed to locate process for cni",
				zap.String("cni name", cni.name),
				zap.Error(err))
		}

	}

	zap.L().Debug("No supported CNI process was found")

	return ""
}
