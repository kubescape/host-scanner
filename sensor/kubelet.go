package sensor

import (
	"fmt"

	ds "github.com/kubescape/host-scanner/sensor/datastructures"
	"github.com/kubescape/host-scanner/sensor/internal/utils"
	"go.uber.org/zap"
	"sigs.k8s.io/yaml"
)

const (
	procDirName            = "/proc"
	kubeletProcessSuffix   = "/kubelet"
	kubeletConfigArgName   = "--config"
	kubeletClientCAArgName = "--client-ca-file"

	// Default paths
	kubeletConfigDefaultPath     = "/var/lib/kubelet/config.yaml"
	kubeletKubeConfigDefaultPath = "/etc/kubernetes/kubelet.conf"
)

// KubeletInfo holds information about kubelet
type KubeletInfo struct {
	// ServiceFile is a list of files used to configure the kubelet service.
	// Most of the times it will be a single file, under /etc/systemd/system/kubelet.service.d.
	ServiceFiles []ds.FileInfo `json:"serviceFiles,omitempty"`

	// Information about kubelete config file
	ConfigFile *ds.FileInfo `json:"configFile,omitempty"`

	// Information about the kubeconfig file of kubelet
	KubeConfigFile *ds.FileInfo `json:"kubeConfigFile,omitempty"`

	// Information about the client ca file of kubelet (if exist)
	ClientCAFile *ds.FileInfo `json:"clientCAFile,omitempty"`

	// Raw cmd line of kubelet process
	CmdLine string `json:"cmdLine"`
}

func LocateKubeletProcess() (*utils.ProcessDetails, error) {
	return utils.LocateProcessByExecSuffix(kubeletProcessSuffix)
}

func ReadKubeletConfig(kubeletConfArgs string) ([]byte, error) {
	conte, err := utils.ReadFileOnHostFileSystem(kubeletConfArgs)
	zap.L().Debug("raw content", zap.ByteString("cont", conte))
	return conte, err
}

func makeKubeletServiceFilesInfo(pid int) []ds.FileInfo {
	files, err := utils.GetKubeletServiceFiles(pid)
	if err != nil {
		zap.L().Warn("failed to getKubeletServiceFiles", zap.Error(err))
		return nil
	}

	serviceFiles := []ds.FileInfo{}
	for _, file := range files {
		info := makeHostFileInfoVerbose(file, false, zap.String("in", "makeProcessInfoVerbose"))
		if info != nil {
			serviceFiles = append(serviceFiles, *info)
		}
	}

	if len(serviceFiles) == 0 {
		return nil
	}

	return serviceFiles
}

// SenseKubeletInfo return varius information about the kubelet service
func SenseKubeletInfo() (*KubeletInfo, error) {
	ret := KubeletInfo{}

	kubeletProcess, err := LocateKubeletProcess()
	if err != nil {
		return &ret, fmt.Errorf("failed to LocateKubeletProcess: %w", err)
	}

	// Serivce files
	ret.ServiceFiles = makeKubeletServiceFilesInfo(int(kubeletProcess.PID))

	// Kubelet config
	configPath := kubeletConfigDefaultPath
	p, ok := kubeletProcess.GetArg(kubeletConfigArgName)
	if ok {
		configPath = p
	}
	ret.ConfigFile = makeContaineredFileInfoVerbose(kubeletProcess, configPath, true,
		zap.String("in", "SenseKubeletInfo"),
	)

	// Kubelet kubeconfig
	kubeConfigPath := kubeletConfigDefaultPath
	p, ok = kubeletProcess.GetArg(kubeConfigArgName)
	if ok {
		kubeConfigPath = p
	}
	ret.KubeConfigFile = makeContaineredFileInfoVerbose(kubeletProcess, kubeConfigPath, false,
		zap.String("in", "SenseKubeletInfo"),
	)

	// Kubelet client ca certificate
	caFilePath, ok := kubeletProcess.GetArg(kubeletClientCAArgName)
	if !ok && ret.ConfigFile != nil && ret.ConfigFile.Content != nil {
		zap.L().Debug("extracting kubelet client ca certificate from config")
		extracted, err := kubeletExtractCAFileFromConf(ret.ConfigFile.Content)
		if err == nil {
			caFilePath = extracted
		}
	}
	if caFilePath != "" {
		ret.ClientCAFile = makeContaineredFileInfoVerbose(kubeletProcess, caFilePath, false,
			zap.String("in", "SenseKubeletInfo"),
		)
	}

	// Cmd line
	ret.CmdLine = kubeletProcess.RawCmd()

	return &ret, nil
}

// kubeletExtractCAFileFromConf extract the client ca file path from kubelet config
func kubeletExtractCAFileFromConf(content []byte) (string, error) {
	var kubeletConfig struct {
		Authentication struct {
			X509 struct {
				ClientCAFile string
			}
		}
	}

	err := yaml.Unmarshal(content, &kubeletConfig)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal kubelet config: %w", err)
	}

	return kubeletConfig.Authentication.X509.ClientCAFile, nil
}

// Deprecated: use SenseKubeletInfo for more information.
// Return the content of kubelet config file
func SenseKubeletConfigurations() ([]byte, error) {
	kubeletProcess, err := LocateKubeletProcess()
	if err != nil {
		return nil, fmt.Errorf("failed to LocateKubeletProcess: %w", err)
	}
	kubeletConfFileLocation, ok := kubeletProcess.GetArg(kubeletConfigArgName)
	if !ok || kubeletConfFileLocation == "" {
		return nil, fmt.Errorf("in SenseKubeletConfigurations failed to find kubelet config File location")
	}

	zap.L().Debug("config loaction", zap.String("kubeletConfFileLocation", kubeletConfFileLocation))
	return ReadKubeletConfig(kubeletProcess.ContaineredPath(kubeletConfFileLocation))
}
