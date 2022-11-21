package utils

import (
	"errors"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"go.uber.org/zap"
)

// CNI default constants
const (
	CNIDefaultConfigDir string = "/etc/cni/"

	// kubelet flags for container runtime and cni configuration dir.
	kubeletContainerRuntime         = "--container-runtime"
	kubeletContainerRuntimeEndPoint = "--container-runtime-endpoint"
	kubeletCNIConfigDir             = "--cni-conf-dir"
)

// Types of supported container runtime processes.
const (
	// container runtimes names
	containerdContainerRuntimeName = "containerd"
	crioContainerRuntimeName       = "crio"

	containerdConfigSection = "io.containerd.grpc.v1.cri"

	// container runtime processes suffix
	crioSock       = "/crio.sock"
	containerdSock = "/containerd.sock"
	cridockerdSock = "/cri-dockerd.sock"
)

// Errors
var (
	// ErrDockershimRT means the used container runtime is dockershim.
	// no kubelet flags or --container-runtime not 'remote' means dockershim.sock which is not supported
	ErrDockershimRT = errors.New("dockershim runtime is not supported")

	// ErrCRINotFound means no container runtime was found.
	ErrCRINotFound = errors.New("no container runtime was found")
)

// A containerRuntimeProperties holds properties of a container runtime.
type containerRuntimeProperties struct {
	Name string

	// Process for cuflag stom config file.
	ConfigArgName string

	// Process flag for custom configuration directory.
	ConfigDirArgName string

	// Default config path.
	DefaultConfigPath string

	// default configuration directory.
	DefaultConfigDir string

	// suffix of container runtime process.
	ProcessSuffix string

	// the socket suffix - used to identify the container runtime from kubelet.
	Socket string

	// process pararm for CNI configuration directory.
	CNIConfigDirArgName string

	// extract CNI info function
	ParseCNIFromConfigFunc func(string) (string, error)
}

// A ContainerRuntimeInfo holds a container runtime properties and process info.
type ContainerRuntimeInfo struct {
	// container runtime properties
	properties *containerRuntimeProperties

	// process pointer
	process *ProcessDetails

	// root
	rootDir string
}

// getCNIConfigPath returns CNI config dir from a running container runtime. Flow:
//  1. Find CNI config dir through kubelet flag (--container-runtime-endpoint). If not found:
//  2. Find CNI config dir through process of supported container runtimes. If not found:
//  3. return CNI config dir default that is defined in the container runtime properties.
func GetCNIConfigPath(kubeletProc *ProcessDetails) string {

	// Attempting to find CR from kubelet.
	CNIConfigDir, err := CNIConfigDirFromKubelet(kubeletProc)

	if err == nil {
		return CNIConfigDir
	}

	// Could construct container runtime from kubelet
	zap.L().Debug("getCNIConfigPath - failed to get CNI config dir through kubelete flags.")

	// Attempting to find CR through process.
	cr, err := getContainerRuntimeFromProcess()

	if err != nil {
		//Failed to get container runtime from process
		zap.L().Debug("getCNIConfigPath - failed to get container runtime from process, return cni config dir default",
			zap.Error(err))

		return CNIDefaultConfigDir
	}

	CNIConfigDir = cr.getCNIConfigDir()
	if CNIConfigDir == "" {
		return CNIDefaultConfigDir
	}
	return CNIConfigDir

}

// getConfigDirPath - returns container runtime config directory through process flag. If not found returns default config directory from container runtime properties.
func (cr *ContainerRuntimeInfo) getConfigDirPath() string {
	configDirPath, _ := cr.process.GetArg(cr.properties.ConfigDirArgName)

	if configDirPath == "" {
		configDirPath = path.Join(cr.rootDir, cr.properties.DefaultConfigDir)
	}

	return configDirPath
}

// getConfigPath - returns container runtime config path through process flag. If not found returns default.
func (cr *ContainerRuntimeInfo) getConfigPath() string {
	configPath, _ := cr.process.GetArg(cr.properties.ConfigArgName)
	if configPath == "" {
		zap.L().Debug("getConfigPath - container runtime config file wasn't found through process flags, return default path",
			zap.String("Container Runtime Name", cr.properties.Name),
			zap.String("defaultConfigPath", cr.properties.DefaultConfigPath))
		configPath = cr.properties.DefaultConfigPath

	} else {
		zap.L().Debug("getConfigPath - container runtime config file found through process flags",
			zap.String("Container Runtime Name", cr.properties.Name),
			zap.String("configPath", configPath))
	}

	return path.Join(cr.rootDir, configPath)
}

// getCNIConfigDirFromConfig - returns CNI Config dir from the container runtime config file if exist.
// flow:
//  1. Getting container runtime configs directory path and container runtime config path.
//  2. Build a decending ordered list of configs from configs directory and adding the config path as last. This is the order of precedence for configuration.
//  3. Get CNI config path from ordered list. If not found, return empty string.
func (cr *ContainerRuntimeInfo) getCNIConfigDirFromConfig() string {

	var configDirFilesFullPath []string

	// Getting all config files in drop in folder if exist.
	configDirPath := cr.getConfigDirPath()

	// Call ReadDir to get all files.
	outputDirFiles, err := os.ReadDir(configDirPath)

	if err != nil {
		zap.L().Error("getCNIConfigDirFromConfig- Failed to Call ReadDir",
			zap.String("configDirPath", configDirPath),
			zap.Error(err))
	} else {
		configDirFilesFullPath = make([]string, len(outputDirFiles), len(outputDirFiles))

		// constuct and reverse sort config dir files full path
		for i, filename := range outputDirFiles {
			configDirFilesFullPath[i] = path.Join(configDirPath, filename.Name())
		}

		sort.Sort(sort.Reverse(sort.StringSlice(configDirFilesFullPath)))
	}

	configPath := cr.getConfigPath()

	//appending config file to the end of the list as it always has the lowest priority.
	if configPath != "" {
		configDirFilesFullPath = append(configDirFilesFullPath, configPath)
	}

	CNIConfigDir := cr.getCNIConfigDirFromConfigPaths(configDirFilesFullPath)

	if CNIConfigDir == "" {
		zap.L().Debug("getCNIConfigDirFromConfig didn't find CNI Config dir in container runtime configs", zap.String("Container Runtime Name", cr.properties.Name))
	}

	return CNIConfigDir

}

// getCNIConfigDirFromConfigPaths - Get a list of configpaths, run through the paths by order, parse the CNI config dir and return once found. If not found, return empty string.
func (cr *ContainerRuntimeInfo) getCNIConfigDirFromConfigPaths(configPaths []string) string {

	for _, configPath := range configPaths {
		CNIConfigDir, err := cr.properties.ParseCNIFromConfigFunc(configPath)

		if err != nil {
			zap.L().Debug("getCNIConfigDirFromConfigPaths - Failed to parse config file", zap.String("configPath", configPath), zap.Error(err))
			continue
		}

		if CNIConfigDir != "" {
			return CNIConfigDir
		}

	}

	return ""

}

// getCNIConfigDirFromProcess - returns CNI config dir from process cmdline flags if defined, otherwise returns empty string.
func (cr *ContainerRuntimeInfo) getCNIConfigDirFromProcess() string {

	if cr.properties.CNIConfigDirArgName != "" {
		CNIConfigDir, _ := cr.process.GetArg(cr.properties.CNIConfigDirArgName)
		if CNIConfigDir != "" {
			zap.L().Debug("getCNIConfigDir found CNI Config Dir in process", zap.String("Container Runtime Name", cr.properties.Name))
		}

		return CNIConfigDir
	}

	return ""

}

// getCNIConfigDir - returns CNI config dir of the container runtime.
//  1. Get dir from container runtime process flags. If not found:
//  2. Get dir from container runtime config file(s). If not found:
//  3. return default CNI config dir
func (cr *ContainerRuntimeInfo) getCNIConfigDir() string {

	CNIConfigDir := cr.getCNIConfigDirFromProcess()

	if CNIConfigDir != "" {
		return CNIConfigDir
	}

	CNIConfigDir = cr.getCNIConfigDirFromConfig()

	return CNIConfigDir
}

// containerdProps - returns container runtime "containerd" properties.
func containerdProps() *containerRuntimeProperties {
	return &containerRuntimeProperties{Name: containerdContainerRuntimeName,
		DefaultConfigPath:      "/etc/containerd/config.toml",
		ProcessSuffix:          "/containerd",
		Socket:                 "/containerd.sock",
		ConfigArgName:          "--config",
		ConfigDirArgName:       "",
		DefaultConfigDir:       "/etc/containerd/containerd.conf.d",
		CNIConfigDirArgName:    "",
		ParseCNIFromConfigFunc: parseCNIConfigDirFromConfigContainerd}

}

// crioProps - returns container runtime "cri-o" properties.
func crioProps() *containerRuntimeProperties {
	return &containerRuntimeProperties{Name: crioContainerRuntimeName,
		DefaultConfigPath:      "/etc/crio/crio.conf",
		ProcessSuffix:          "/crio",
		Socket:                 "/crio.sock",
		ConfigArgName:          "--config",
		ConfigDirArgName:       "--config-dir",
		DefaultConfigDir:       "/etc/crio/crio.conf.d",
		CNIConfigDirArgName:    "--cni-config-dir",
		ParseCNIFromConfigFunc: parseCNIConfigDirFromConfigCrio}

}

// newContainerRuntime is a constructor for ContainerRuntime object. Constructor will fail if process wasn't found for container runtime.
// Constructor accept CRIKind as parameter which can be either a container runtime name or container runtime process suffix.
func newContainerRuntime(CRIKind string) (*ContainerRuntimeInfo, error) {

	cr := &ContainerRuntimeInfo{}

	switch CRIKind {
	case containerdContainerRuntimeName, containerdSock:
		cr.properties = containerdProps()
	case crioContainerRuntimeName, crioSock:
		cr.properties = crioProps()

	default:
		return nil, fmt.Errorf("newContainerRuntime of kind '%s' is not supported", CRIKind)

	}
	p, err := LocateProcessByExecSuffix(cr.properties.ProcessSuffix)

	// if process wasn't find, fail to construct object
	if err != nil || p == nil {
		return nil, fmt.Errorf("newContainerRuntime - Failed to locate process for CRIKind %s", CRIKind)
	}

	cr.process = p
	cr.rootDir = HostFileSystemDefaultLocation

	return cr, nil

}

// getContainerRuntimeFromProcess - returns first container runtime found by process.
func getContainerRuntimeFromProcess() (*ContainerRuntimeInfo, error) {

	crObj, err := newContainerRuntime(containerdContainerRuntimeName)

	if err != nil {
		crObj, err = newContainerRuntime(crioContainerRuntimeName)

		if err != nil {
			return nil, fmt.Errorf("getContainerRuntimeFromProcess didnt find Container Runtime process")
		}
	}

	return crObj, nil

}

// parseCNIConfigDirFromConfigContainerd - parses and returns cni config dir from a containerd config structure. If not found returns empty string.
func parseCNIConfigDirFromConfigContainerd(configPath string) (string, error) {

	cniConfig := struct {
		Plugins struct {
			IoContainerdGrpcV1CRI struct {
				CNI struct {
					CNIConfigDir string `toml:"conf_dir"`
				} `toml:"cni"`
			} `toml:"io.containerd.grpc.v1.cri"`
		} `toml:"plugins"`
	}{}

	_, err := toml.DecodeFile(configPath, &cniConfig)

	if err != nil {
		return "", err
	}

	return cniConfig.Plugins.IoContainerdGrpcV1CRI.CNI.CNIConfigDir, nil
}

// parseCNIConfigDirFromConfigCrio - parses and returns cni config dir from a cri-o config structure. If not found returns empty string.
func parseCNIConfigDirFromConfigCrio(configPath string) (string, error) {
	cniConfig := struct {
		Crio struct {
			Network struct {
				NetworkDir string `toml:"network_dir"`
			} `toml:"network"`
		} `toml:"crio"`
	}{}

	_, err := toml.DecodeFile(configPath, &cniConfig)

	if err != nil {
		return "", err
	}

	return cniConfig.Crio.Network.NetworkDir, nil
}

// CNIConfigDirFromKubelet - returns cni config dir by kubelet --container-runtime-endpoint flag. Returns empty string if not found.
// A specific case is cri-dockerd.sock process which it's container runtime is determined by kubernetes docs.
func CNIConfigDirFromKubelet(proc *ProcessDetails) (string, error) {

	// Try from kubelet process flags
	CNIConfigDir, _ := proc.GetArg(kubeletCNIConfigDir)
	if CNIConfigDir != "" {
		return CNIConfigDir, nil
	}

	// Try from kubelet process CRI socket
	crEndpoint, crEndPointOK := proc.GetArg(kubeletContainerRuntimeEndPoint)
	if crEndpoint == "" {
		cr, crOK := proc.GetArg(kubeletContainerRuntime)
		if (!crEndPointOK && !crOK) || (cr != "remote") {
			// From k8s docs (https://kubernetes.io/docs/tasks/administer-cluster/migrating-from-dockershim/find-out-runtime-you-use/#which-endpoint):
			// 	"
			// 	If your nodes use Kubernetes v1.23 and earlier and these flags aren't present
			// 	or if the --container-runtime flag is not remote, you use the dockershim socket with Docker Engine.
			// 	"
			return "", ErrDockershimRT
		}
		// Unknown
		return "", ErrCRINotFound
	}

	containerProcessSock := crEndpoint
	if strings.HasSuffix(crEndpoint, cridockerdSock) {
		// Check specific case where the end point is cri-dockerd. If so, then in the absence of cni paths configuration for cri-dockerd process,
		// we check containerd (which is using cri-dockerd as a CRI plugin)
		containerProcessSock = containerdSock
	}

	crObj, err := newContainerRuntime(containerProcessSock)

	if err != nil {
		return "", err
	}

	return crObj.getCNIConfigDir(), nil
}
