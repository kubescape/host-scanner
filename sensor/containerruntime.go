package sensor

import (
	"fmt"
	"path"
	"strings"

	"github.com/BurntSushi/toml"
	"go.uber.org/zap"
)

// CNI default constants
const (
	CNIDefaultConfigDir string = "/etc/cni/"
)

// Constant values for different types of Container Runtimes
const (
	// container runtimes
	containerdContainerRuntimeName = "containerd"
	containerdSock                 = "/containerd.sock"
	containerdConfigSection        = "io.containerd.grpc.v1.cri"

	crioContainerRuntimeName = "crio"
	crioSock                 = "/crio.sock"

	dockershimSock = "/dockershim.sock"

	// container runtime interfaces
	cridockerdContainerRuntimeName = "cri-dockerd"
	cridockerdSock                 = "/cri-dockerd.sock"
)

// General properties for container runtimes
type containerRuntimeProperties struct {
	Name string

	// whether container runtime supports config files (cri-dockerd is an example doesn't support it)
	ConfigSupported bool

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

	// The location of the CNI config files
	CNIConfigDir string `json:"CNIConfigDir,omitempty"`

	// root
	rootDir string
}

// list of container runtime properties.
var containersRuntimeProperties = []containerRuntimeProperties{
	{
		Name:                   containerdContainerRuntimeName,
		ConfigSupported:        true,
		DefaultConfigPath:      "/etc/containerd/config.toml",
		ProcessSuffix:          "/containerd",
		Socket:                 "/containerd.sock",
		ConfigArgName:          "--config",
		ConfigDirArgName:       "",
		DefaultConfigDir:       "/etc/containerd/containerd.conf.d",
		CNIConfigDirArgName:    "",
		ParseCNIFromConfigFunc: parseCNIConfigDirFromConfigContainerd,
	},
	{
		Name:                   crioContainerRuntimeName,
		ConfigSupported:        true,
		DefaultConfigPath:      "/etc/crio/crio.conf",
		ProcessSuffix:          "/crio",
		Socket:                 "/crio.sock",
		ConfigArgName:          "--config",
		ConfigDirArgName:       "--config-dir",
		DefaultConfigDir:       "/etc/crio/crio.conf.d",
		CNIConfigDirArgName:    "--cni-config-dir",
		ParseCNIFromConfigFunc: parseCNIConfigDirFromConfigCrio,
	},
	{
		Name:                   cridockerdContainerRuntimeName,
		ConfigSupported:        false,
		DefaultConfigPath:      "",
		ProcessSuffix:          "/cri-dockerd",
		Socket:                 "/cri-dockerd.sock",
		ConfigArgName:          "",
		ConfigDirArgName:       "",
		DefaultConfigDir:       "",
		CNIConfigDirArgName:    "--cni-conf-dir",
		ParseCNIFromConfigFunc: parseCNIConfigDirFromConfigCridockerd,
	},
}

// Get CNI config dir from running Container Runtimes. Flow:
// 1. Find CNI config dir through kubelet flag (--container-runtime-endpoint). If not found:
// 2. Find CNI config dir through process of supported container runtimes. If not found:
// 3. return CNI config dir default.
func getContainerRuntimeCNIConfigPath() string {

	// Attempting to find CR from kubelet.
	CNIConfigDir := CNIConfigDirFromKubelet()

	if CNIConfigDir != "" {
		return CNIConfigDir
	}

	// Could construct container runtime from kubelet
	zap.L().Debug("getContainerRuntimeCNIConfigPath - failed to get CNI config dir through kubelete, trying through process")

	// Attempting to find CR through process.
	cr, err := getContainerRuntimeFromProcess()

	if err == nil {
		if cr.CNIConfigDir == "" {
			return CNIDefaultConfigDir
		}
		return cr.CNIConfigDir
	}

	//Failed to get container runtime from process
	zap.L().Debug("getContainerRuntimeCNIConfigPath - failed to get container runtime from process, return cni config dir default",
		zap.Error(err))

	return CNIDefaultConfigDir

}

func (cr *ContainerRuntimeInfo) setProperties(properies *containerRuntimeProperties) {
	cr.properties = properies
}

// get config directory. First try through process, if wasn't found taking default.
func (cr *ContainerRuntimeInfo) getConfigDirPath() string {
	configDirPath := cr.getArgFromProcess(cr.properties.ConfigDirArgName)

	if configDirPath == "" {
		configDirPath = path.Join(cr.rootDir, cr.properties.DefaultConfigDir)
	}

	return configDirPath
}

// Getting container runtime config path through process flag. If not found, return default config path.
func (cr *ContainerRuntimeInfo) getConfigPath() string {
	configPath := cr.getArgFromProcess(cr.properties.ConfigArgName)
	if configPath == "" {
		zap.L().Debug("getConfigPath - custom config no found through process, taking default config path",
			zap.String("Container Runtime Name", cr.properties.Name),
			zap.String("defaultConfigPath", cr.properties.DefaultConfigPath))
		configPath = cr.properties.DefaultConfigPath

	} else {
		zap.L().Debug("getConfigPath - custom config found in process",
			zap.String("Container Runtime Name", cr.properties.Name),
			zap.String("configPath", configPath))
	}

	return path.Join(cr.rootDir, configPath)
}

func (cr *ContainerRuntimeInfo) getArgFromProcess(argName string) string {
	if argName == "" {
		return ""
	}

	res, ok := cr.process.GetArg(argName)
	if !ok || res == "" {
		return ""
	} else {
		return res
	}

}

// Extract CNI Config dir information from the container runtime config file if exist.
// flow:
// 1. If not default config is set, return nils. else:
// 2. Looking for config file through process cmdline, if not found:
// 3. Use default config path.
// 4. Extract CNI config dur from config through a custom function of the Container Runtime. If not found, return empty string
func (cr *ContainerRuntimeInfo) getCNIConfigDirFromConfig() string {

	var configDirFilesFullPath []string

	if !cr.properties.ConfigSupported {
		return ""
	}

	//Getting all config files in drop in folder if exist.
	configDirPath := cr.getConfigDirPath()
	configDirFilesFullPath = makeSortedFilesList(configDirPath, false)

	configPath := cr.getConfigPath()

	//appding config file to the end of the list as it always has the lowest priority.
	if configPath != "" {
		configDirFilesFullPath = append(configDirFilesFullPath, configPath)
	}

	CNIConfigDir := cr.getCNIConfigDirFromConfigPaths(configDirFilesFullPath)

	if CNIConfigDir != "" {
		zap.L().Debug("getCNIConfigDirFromConfig - found cni paths in configs for container runtime",
			zap.String("Container Runtime Name", cr.properties.Name))
	}

	return CNIConfigDir

}

// Get a list of configpaths and a parsing function and returns CNI config dir.
// iteration is done by the original order of the configpaths.
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

// Get CNI config dir from process cmdline flags if such defined.
func (cr *ContainerRuntimeInfo) getCNIConfigDirFromProcess() string {
	var CNIConfigDir string

	if cr.properties.CNIConfigDirArgName != "" {
		CNIConfigDir, ok := cr.process.GetArg(cr.properties.CNIConfigDirArgName)

		if !ok || CNIConfigDir == "" {
			zap.L().Debug("getCNIConfigDirFromProcess - no cni config dir found for process",
				zap.String("ContainerRuntime name", cr.properties.Name),
				zap.String("CNIConfigDirArgName", cr.properties.CNIConfigDirArgName))
		}
	}

	return CNIConfigDir

}

// update CNI paths property. Flow:
// 1. Try to get paths from process flags. If not found:
// 2. Try to get paths from config file. If not found:
// 3. return defaults
func (cr *ContainerRuntimeInfo) updateCNIConfigDir() {

	CNIConfigDir := cr.getCNIConfigDirFromProcess()

	if CNIConfigDir != "" {
		zap.L().Debug("updateCNIConfigDir found CNI Config Dir in process", zap.String("Container Runtime Name", cr.properties.Name))
		cr.CNIConfigDir = CNIConfigDir
		return
	}

	CNIConfigDir = cr.getCNIConfigDirFromConfig()

	if CNIConfigDir != "" {
		zap.L().Debug("updateCNIConfigDir found CNI Config dir in configs", zap.String("Container Runtime Name", cr.properties.Name))
		cr.CNIConfigDir = CNIConfigDir
	}
}

// Constructor for ContainerRuntime object. Constructor will fail if process wasn't found for container runtime.
func NewContainerRuntime(properties containerRuntimeProperties, rootDir string) (*ContainerRuntimeInfo, error) {

	p, err := LocateProcessByExecSuffix(properties.ProcessSuffix)

	// if process wasn't find, fail to construct object
	if err != nil || p == nil {
		return nil, fmt.Errorf("NewContainerRuntime - Failed to locate process for %s", properties.Name)
	}

	cr := &ContainerRuntimeInfo{}
	cr.process = p
	cr.rootDir = rootDir
	cr.setProperties(&properties)
	cr.updateCNIConfigDir()

	return cr, nil

}

// Return container runtime properties by name
func getContainerRuntimeProperties(containerRuntimeName string) (*containerRuntimeProperties, error) {
	for _, crp := range containersRuntimeProperties {
		if crp.Name == containerRuntimeName {
			return &crp, nil
		}
	}

	return nil, fmt.Errorf("ContainerRuntimeName %s not found", containerRuntimeName)
}

// Get container runtime end point (i.e. [name].sock) and returns container runtime object if supported / exists.
func getContainerRuntime(crEndpoint string) (*ContainerRuntimeInfo, error) {
	for _, crp := range containersRuntimeProperties {
		if strings.HasSuffix(crEndpoint, crp.Socket) {
			return NewContainerRuntime(crp, hostFileSystemDefaultLocation)
		}
	}
	return nil, fmt.Errorf("getContainerRuntime End point '%s' is not supported", crEndpoint)
}

// Returns first container runtime found
// Search for process excludes cri-dockerd as if it is present there should be another process for the main container runtime.
func getContainerRuntimeFromProcess() (*ContainerRuntimeInfo, error) {

	for _, crp := range containersRuntimeProperties {
		if crp.Name != cridockerdContainerRuntimeName {
			crObj, err := NewContainerRuntime(crp, hostFileSystemDefaultLocation)
			if err == nil {
				return crObj, nil
			}
		}
	}

	return nil, fmt.Errorf("getContainerRuntimeFromProcess got more than one Container Runtime process")
}

// Read Containerd specific config structure to extract CNI paths.
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

// Read cri-o specific config structure to extract CNI paths.
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

// Not implemented.
func parseCNIConfigDirFromConfigCridockerd(configPath string) (string, error) {
	return "", fmt.Errorf("parseCNIConfigDirFromConfigCridockerd not implemented")
}

// Get CNI Paths from container runtime defined for kubelet.
// Container runtime is expected to be found in --container-runtime-endpoint.
func CNIConfigDirFromKubelet() string {
	proc, err := LocateProcessByExecSuffix(kubeletProcessSuffix)
	if err != nil {
		zap.L().Debug("CNIConfigDirFromKubelet - failed to locate kube-proxy process")
		return ""
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
		zap.L().Debug(ErrCRNotFound.Error())
		return ""

	}
	// there is crEndpoint
	zap.L().Debug("crEndPoint from kubelete found", zap.String("crEndPoint", crEndpoint))
	crObj, err := getContainerRuntime(crEndpoint)

	if err != nil {
		return ""
	}
	// Successfully created a Container runtime object, get cni config dir
	CNIConfigDir := crObj.CNIConfigDir

	if CNIConfigDir == "" {
		// Didn't find cni config dir.
		// Check specific case where the end point is cri-dockerd. If so, then in the absence of cni paths configuration for cri-dockerd process,
		// we check containerd (which is using cri-dockerd as a CRI plugin)

		if crObj.properties.Name == cridockerdContainerRuntimeName {

			crObj, err := getContainerRuntime(containerdSock)

			if err == nil {
				return crObj.CNIConfigDir
			}

		}
	}
	return CNIConfigDir
}
