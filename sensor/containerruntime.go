package sensor

import (
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

	kubeletContainerRuntime         = "--container-runtime"
	kubeletContainerRuntimeEndPoint = "--container-runtime-endpoint"
	kubeletCNIConfigDir             = "--cni-conf-dir"
)

// Constant values for different types of Container Runtimes
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

// General properties for container runtimes
type containerRuntimeProperties struct {
	Name string

	// Process for cuflag stom config file.
	ConfigArgName string

	// Process flag for custom configuration directory.
	ConfigDirArgName string

	// Default config path
	DefaultConfigPath string

	// default configuration directory
	DefaultConfigDir string

	// suffix of container runtime process
	ProcessSuffix string

	// the socket suffix - used to identify the container runtime from kubelet
	Socket string

	// process pararm for CNI configuration directory
	CNIConfigDirArgName string

	// extract CNI info function
	ParseCNIFromConfigFunc func(string) (string, error)
}

// Struct to hold all information of a container runtime
type ContainerRuntimeInfo struct {
	properties *containerRuntimeProperties

	// process pointer
	process *ProcessDetails

	// root
	rootDir string
}

// getCNIConfigPath returns CNI config dir from a running Container Runtimes. Flow:
// 1. Find CNI config dir through kubelet flag (--container-runtime-endpoint). If not found:
// 2. Find CNI config dir through process of supported container runtimes. If not found:
// 3. return CNI config dir default.
func getCNIConfigPath() string {

	// Attempting to find CR from kubelet.
	CNIConfigDir := CNIConfigDirFromKubelet()

	if CNIConfigDir != "" {
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

// getConfigDirPath - returns container runtime config directory through process flag. If not found returns default.
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
// 1. If not default config is set, return nils. else:
// 2. Looking for config file through process cmdline, if not found:
// 3. Use default config path.
// 4. Extract the CNI config dir from config through a custom function of the Container Runtime. If not found, return an empty string.
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

	//appding config file to the end of the list as it always has the lowest priority.
	if configPath != "" {
		configDirFilesFullPath = append(configDirFilesFullPath, configPath)
	}

	CNIConfigDir := cr.getCNIConfigDirFromConfigPaths(configDirFilesFullPath)

	if CNIConfigDir == "" {
		zap.L().Debug("getCNIConfigDirFromConfig didn't find CNI Config dir in container runtime configs", zap.String("Container Runtime Name", cr.properties.Name))
	}

	return CNIConfigDir

}

// getCNIConfigDirFromConfigPaths - Get a list of configpaths and a parsing function and returns CNI config dir.
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

// getCNIConfigDirFromProcess - returns CNI config dir from process cmdline flags if defined.
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

// getCNIConfigDir - returns CNI config dir of the container runtime
// 1. Try to get paths from process flags. If not found:
// 2. Try to get paths from config file. If not found:
// 3. return default
func (cr *ContainerRuntimeInfo) getCNIConfigDir() string {

	CNIConfigDir := cr.getCNIConfigDirFromProcess()

	if CNIConfigDir != "" {
		return CNIConfigDir
	}

	CNIConfigDir = cr.getCNIConfigDirFromConfig()

	return CNIConfigDir
}

// containerdProps - returns container runtime "containerd" properties
func newContainerdProps() *containerRuntimeProperties {
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

// crioProps - returns container runtime "cri-o" properties
func newCrioProps() *containerRuntimeProperties {
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

// Constructor for ContainerRuntime object. Constructor will fail if process wasn't found for container runtime.
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
	cr.rootDir = hostFileSystemDefaultLocation
	// cr.updateCNIConfigDir()

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

// parseCNIConfigDirFromConfigContainerd - returns cni config dir from a containerd config structure. If not found returns empty string.
func parseCNIConfigDirFromConfigContainerd(configPath string) (string, error) {

	cniConfig := struct {
		Plugings map[string]struct {
			CNI struct {
				CNIConfigDir string `toml:"conf_dir"`
			} `toml:"cni"`
		} `toml:"plugins"`
	}{}

	_, err := toml.DecodeFile(configPath, &cniConfig)

	if err != nil {
		return "", err
	}

	return cniConfig.Plugings[containerdConfigSection].CNI.CNIConfigDir, nil
}

// parseCNIConfigDirFromConfigCrio - returns cni config dir from a cri-o config structure. If not found returns empty string.
func parseCNIConfigDirFromConfigCrio(configPath string) (string, error) {

	cniConfig := struct {
		Crio map[string]struct {
			CNIConfigDir string `toml:"network_dir"`
		} `toml:"crio"`
	}{}

	_, err := toml.DecodeFile(configPath, &cniConfig)

	if err != nil {
		return "", err
	}

	return cniConfig.Crio["network"].CNIConfigDir, nil
}

// parseCNIConfigDirFromConfigCridockerd - Not implemented.
func parseCNIConfigDirFromConfigCridockerd(configPath string) (string, error) {
	return "", fmt.Errorf("parseCNIConfigDirFromConfigCridockerd not implemented")
}

// CNIConfigDirFromKubelet - returns cni config dir by kubelet --container-runtime-endpoint flag.
func CNIConfigDirFromKubelet() string {

	var containerProcessSock string
	proc, err := LocateKubeletProcess()
	if err != nil {
		zap.L().Debug("CNIConfigDirFromKubelet - failed to locate kube-proxy process")
		return ""
	}

	CNIConfigDir, _ := proc.GetArg(kubeletCNIConfigDir)

	if CNIConfigDir != "" {
		return CNIConfigDir
	}

	crEndpoint, crEndPointOK := proc.GetArg(kubeletContainerRuntimeEndPoint)

	if crEndpoint == "" {
		cr, crOK := proc.GetArg(kubeletContainerRuntime)

		if (!crEndPointOK && !crOK) || (cr != "remote") {
			// From docs: "If your nodes use Kubernetes v1.23 and earlier and these flags aren't present
			// or if the --container-runtime flag is not remote, you use the dockershim socket with Docker Engine."
			zap.L().Debug("CNIConfigDirFromKubelet - no kubelet flags or --container-runtime not 'remote' means dockershim.sock which is not supported")
			return ""

		}
		// Uknown
		zap.L().Debug("CNIConfigDirFromKubelet - failed to find Container Runtime EndPoint")
		return ""

	}
	// there is crEndpoint
	zap.L().Debug("crEndPoint from kubelete found", zap.String("crEndPoint", crEndpoint))

	containerProcessSock = crEndpoint

	if strings.HasSuffix(crEndpoint, cridockerdSock) {
		// Check specific case where the end point is cri-dockerd. If so, then in the absence of cni paths configuration for cri-dockerd process,
		// we check containerd (which is using cri-dockerd as a CRI plugin)
		containerProcessSock = containerdSock

	}

	crObj, err := newContainerRuntime(containerProcessSock)

	if err != nil {
		return ""
	}

	return crObj.getCNIConfigDir()
}
