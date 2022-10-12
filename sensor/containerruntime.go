package sensor

import (
	"fmt"
	"path"
	"strings"

	"github.com/BurntSushi/toml"
	"go.uber.org/zap"
)

//CNI default constants
const (
	CNIDefaultConfigDir string = "/etc/cni/"
	CNIDefaultBinDir    string = "/opt/cni/bin/"

	CNIPATHS_SOURCE_PROCESS_PARAMS string = "process_params"
	CNIPATHS_SOURCE_PROCESS_CONFIG string = "process_config"
	CNIPATHS_SOURCE_DEFAULT_CONFIG string = "default_config"
	CNIPATHS_SOURCE_DEFAULT        string = "default"
)

//CNIPath struct
type CNIPaths struct {

	//Where we found the paths.
	Source string `json:"Source,omitempty"`

	//The location of the CNI config files
	Conf_dir string `json:"CNIPathsConfigDir,omitempty"`

	//The location(s) of the binaries. It is a list because some of the CRs configure more than one dir as list.
	Bin_dirs []string `json:"CNIPathsBinDirs,omitempty"`
}

//Constant values for different types of Container Runtimes
const (
	//container runtimes
	CONTAINERD_CONTAINER_RUNTIME_NAME = "containerd"
	CONTAINERD_SOCK                   = "/containerd.sock"
	CONTAINERD_CONFIG_SECTION         = "io.containerd.grpc.v1.cri"

	CRIO_CONTAINER_RUNTIME_NAME = "crio"
	CRIO_SOCK                   = "/crio.sock"

	DOCKERSHIM_SOCK = "/dockershim.sock"

	//container runtime interfaces
	CRIDOCKERD_CONTAINER_RUNTIME_NAME = "cri-dockerd"
	CRIDOCKERD_SOCK                   = "/cri-dockerd.sock"
)

//General properties for container runtimes
type containerRuntimeProperties struct {
	Name string

	//if false, container runtime constructor will fail
	Supported bool

	ConfigParam       string
	DefaultConfigPath string
	ProcessSuffix     string

	//the socket suffix - used to identify the container runtime from kubelet
	Socket string

	CNIProcessConfigDirParam string
	CNIProcessPluginDirParam string

	//extract CNI info function
	ExtractCNIFromConfigFunc func(string, string) (*CNIPaths, error)
}

//list of container runtime properties.
var containersRuntimeProperties = []containerRuntimeProperties{
	{
		Name:                     CONTAINERD_CONTAINER_RUNTIME_NAME,
		Supported:                true,
		DefaultConfigPath:        "/etc/containerd/config.toml",
		ProcessSuffix:            "/containerd",
		Socket:                   "/containerd.sock",
		ConfigParam:              "--config",
		CNIProcessConfigDirParam: "",
		CNIProcessPluginDirParam: "",
		ExtractCNIFromConfigFunc: extractContainerdCNIPathsFromConfig,
	},
	{
		Name:                     CRIO_CONTAINER_RUNTIME_NAME,
		Supported:                true,
		DefaultConfigPath:        "/etc/crio/crio.conf",
		ProcessSuffix:            "/crio",
		Socket:                   "/crio.sock",
		ConfigParam:              "--config",
		CNIProcessConfigDirParam: "--cni-config-dir",
		CNIProcessPluginDirParam: "--cni-plugin-dir",
		ExtractCNIFromConfigFunc: extractCrioCNIPathsFromConfig,
	},
	{
		Name:                     CRIDOCKERD_CONTAINER_RUNTIME_NAME,
		Supported:                true,
		DefaultConfigPath:        "",
		ProcessSuffix:            "/cri-dockerd",
		Socket:                   "/cri-dockerd.sock",
		ConfigParam:              "",
		CNIProcessConfigDirParam: "--cni-conf-dir",
		CNIProcessPluginDirParam: "--cni-bin-dir",
		ExtractCNIFromConfigFunc: extractCriDockedCNIPathsFromConfig,
	},
}

//Return container runtime properties by name
func getContainerRuntimeProperties(containerRuntimeName string) (*containerRuntimeProperties, error) {
	for _, crp := range containersRuntimeProperties {
		if crp.Name == containerRuntimeName {
			return &crp, nil
		}
	}

	return nil, fmt.Errorf("ContainerRuntimeName %s not found", containerRuntimeName)
}

//Get container runtime end point (i.e. [name].sock) and returns container runtime object if supported / exists.
func getContainerRuntime(crEndpoint string) (*ContainerRuntimeInfo, error) {
	for _, crp := range containersRuntimeProperties {
		if strings.HasSuffix(crEndpoint, crp.Socket) && crp.Supported {
			return NewContainerRuntime(crp, hostFileSystemDefaultLocation)
		}
	}
	return nil, fmt.Errorf("getContainerRuntime End point '%s' is not supported", crEndpoint)
}

//Get default CNIPaths - in use in case we couldnt find the paths through configs.
func getCNIPathsDefault() *CNIPaths {
	bin_dirs := []string{CNIDefaultBinDir}
	return &CNIPaths{Conf_dir: CNIDefaultConfigDir, Bin_dirs: bin_dirs, Source: CNIPATHS_SOURCE_DEFAULT}
}

// Get CNI paths from running Container Runtimes. Flow:
// 1. Find CNI through kubelet params (--container-runtime_endpoint). If not found:
// 2. Find CNI through process of supported container runtimes. If not found:
// 3. return CNI default paths.
func getContainerRuntimeCNIPaths() (*CNIPaths, error) {

	// //Attempting to find CR from kubelet.
	cni_paths, err := CNIPathsFromKubelet()

	if err != nil {
		//Could construct container runtime from kubelet
		zap.L().Debug("getContainerRuntimeCNIPaths - failed to get CNI Paths through kubelete, trying through process",
			zap.Error(err))

		//Attempting to find CR through process.
		cr, err := getContainerRuntimeFromProcess(false)

		if err == nil {
			cni_paths := cr.getCNIPaths()
			if cni_paths == nil {
				return getCNIPathsDefault(), nil
			} else {
				return cni_paths, nil
			}

		} else {
			//Failed to get container runtime from process
			zap.L().Debug("getContainerRuntimeCNIPaths - failed to get container runtime from process, return cni defaults",
				zap.Error(err))

			return getCNIPathsDefault(), nil
		}
	} else {
		if cni_paths == nil {
			return getCNIPathsDefault(), nil
		} else {
			return cni_paths, nil
		}

	}

}

// Struct to hold config file information of the container runtime
type containerRuntimeConfig struct {

	//Container runtime config path.
	ConfigPath string `json:"ContainerRuntimeConfigPath,omitempty"`

	//configRootPath
	configRootPath string

	//CNI files paths
	CNI_files *CNIPaths
}

//Struct to hold all information of a container runtime
type ContainerRuntimeInfo struct {
	properties *containerRuntimeProperties

	// process pointer
	process *ProcessDetails

	//CR onfig information if exist.
	config containerRuntimeConfig
}

//Get/Set functions
func (cr *ContainerRuntimeInfo) getDefaultConfigPath() string {
	return cr.properties.DefaultConfigPath
}

func (cr *ContainerRuntimeInfo) getConfigPath() string { return cr.config.ConfigPath }

func (cr *ContainerRuntimeInfo) setConfigPath(configPath string) { cr.config.ConfigPath = configPath }

func (cr *ContainerRuntimeInfo) getProcessSuffix() string { return cr.properties.ProcessSuffix }

func (cr *ContainerRuntimeInfo) getProcess() *ProcessDetails { return cr.process }

func (cr *ContainerRuntimeInfo) setProcess(p *ProcessDetails) { cr.process = p }

func (cr *ContainerRuntimeInfo) getName() string { return cr.properties.Name }

func (cr *ContainerRuntimeInfo) getCNIPaths() *CNIPaths { return cr.config.CNI_files }

func (cr *ContainerRuntimeInfo) setCNIPaths(cni_paths *CNIPaths) { cr.config.CNI_files = cni_paths }

func (cr *ContainerRuntimeInfo) setProperties(properies *containerRuntimeProperties) {
	cr.properties = properies
}

//Getting the location of the config file of the container runtime through cmdline param.
// If param is not set, return empty string
func (cr *ContainerRuntimeInfo) getConfigPathFromProcess() string {
	if cr.properties.ConfigParam == "" {
		return ""
	}

	p := cr.getProcess()
	if p != nil {
		configPath, ok := p.GetArg(cr.properties.ConfigParam)
		if !ok || configPath == "" {
			return ""
		} else {
			return configPath
		}
	}

	return ""
}

//Extract CNI dirs information from the CR config file if exist.
// flow:
// 1. If not default config is set, return nils. else:
// 2. Looking for config file through process cmdline, if not found:
// 3. Use default config path.
// 4. Extrac CNI paths from config through a custom function of the Container Runtime. If not paths found, return nil
func (cr *ContainerRuntimeInfo) getCNIPathsFromConfig(configPath string) (*CNIPaths, error) {

	var cni_source string

	if cr.getDefaultConfigPath() == "" {
		return nil, nil
	}

	if configPath == "" {
		configPath = cr.getConfigPathFromProcess()
		if configPath == "" {
			zap.L().Debug("getCNIPathsFromConfig - custom config file not set for CR, taking default config path",
				zap.String("CR_name", cr.getName()),
				zap.String("defaultConfigPath", cr.getDefaultConfigPath()))
			configPath = cr.getDefaultConfigPath()
			cni_source = CNIPATHS_SOURCE_DEFAULT_CONFIG
		} else {
			cni_source = CNIPATHS_SOURCE_PROCESS_CONFIG
			zap.L().Debug("getCNIPathsFromConfig - custom config file was set for CR",
				zap.String("CR_name", cr.getName()),
				zap.String("configPath", configPath))
		}
	}

	cni_paths, err := cr.properties.ExtractCNIFromConfigFunc(configPath, cr.config.configRootPath)

	if err == nil {
		zap.L().Debug("getCNIPathsFromConfig - found config file for container runtime",
			zap.String("CR_name", cr.getName()),
			zap.String("configPath", configPath))
	}

	if cni_paths != nil {
		cni_paths.Source = cni_source
	}
	return cni_paths, err

}

//Get CNI Paths from process cmdline params if such defined.
func (cr *ContainerRuntimeInfo) getCNIPathsFromProcess() (*CNIPaths, error) {
	p := cr.getProcess()
	if p == nil {
		return nil, fmt.Errorf("No proccess found for %s", cr.getName())
	}

	var conf_dir string
	var bin_dirs []string

	if cr.properties.CNIProcessConfigDirParam != "" {
		conf_dir, ok := p.GetArg(cr.properties.CNIProcessConfigDirParam)

		if !ok || conf_dir == "" {
			zap.L().Debug("getCNIPathsFromProccess no cni config dir found for process",
				zap.String("ContainerRuntime name", cr.getName()),
				zap.String("CNIProcessConfigDirParam", cr.properties.CNIProcessConfigDirParam))
			// conf_dir = CNIDefaultConfigDir
		}
	}

	if cr.properties.CNIProcessPluginDirParam != "" {
		bin_dir, ok := p.GetArg(cr.properties.CNIProcessPluginDirParam)

		if !ok || bin_dir == "" {
			zap.L().Debug("getCNIPathsFromProccess no cni plugin dir found for process",
				zap.String("ContainerRuntime name", cr.getName()),
				zap.String("CNIProcessConfigDirParam", cr.properties.CNIProcessPluginDirParam))
			// conf_dir = CNIDefaultConfigDir
			// bin_dirs = append(bin_dirs, CNIDefaultBinDir)
		} else {
			if bin_dir != "" {
				bin_dirs = append(bin_dirs, bin_dir)
			}
		}
	}

	if len(bin_dirs) == 0 && conf_dir == "" {
		return nil, nil
	} else {
		return &CNIPaths{Conf_dir: conf_dir, Bin_dirs: bin_dirs, Source: CNIPATHS_SOURCE_PROCESS_PARAMS}, nil
	}

}

//Find process by container runtime process suffix
func (cr *ContainerRuntimeInfo) locateProcess() (*ProcessDetails, error) {
	p, err := LocateProcessByExecSuffix(cr.getProcessSuffix())

	if err == nil {
		cr.setProcess(p)
	}
	return p, err
}

// update CNI paths property. Flow:
// 1. Try to get paths from config file. If not found:
// 2. Try to get paths from process params. If not found:
// 3. return defaults
func (cr *ContainerRuntimeInfo) updateCNIPaths() {

	var err error
	var cni_paths *CNIPaths

	configPath := cr.getConfigPath()
	CR_name := cr.getName()

	cni_paths, err = cr.getCNIPathsFromConfig("")

	if err != nil {
		zap.L().Debug("updateCNIPaths Failed to get paths from config, trying through process",
			zap.String("ContainerRuntime", CR_name),
			zap.String("configPath", configPath),
			zap.Error(err),
		)

		cni_paths, err = cr.getCNIPathsFromProcess()

		if err != nil {
			//If didn't succeed to get cni paths from config, return global.
			zap.L().Debug("updateCNIPaths Failed to get paths from process, taking defaults",
				zap.String("ContainerRuntime", CR_name),
				zap.String("configPath", configPath),
				zap.Error(err))
			cni_paths, err = nil, nil
		}

	} else {

		if cni_paths != nil {
			// Found config file, checking that paths are not empty.
			if cni_paths.Conf_dir == "" {
				zap.L().Debug("updateCNIPaths ContainerRuntime has config without conf_dir definition, taking default",
					zap.String("ContainerRuntime", CR_name),
					zap.String("configPath", configPath),
				)
				cni_paths.Conf_dir = CNIDefaultConfigDir
			}

			if cni_paths.Bin_dirs == nil {
				zap.L().Debug("updateCNIPaths ContainerRuntime has config without Bin_dirs definition, taking default",
					zap.String("ContainerRuntime", CR_name),
					zap.String("configPath", configPath),
				)
				cni_paths.Conf_dir = CNIDefaultBinDir
			}
		}

	}

	cr.setCNIPaths(cni_paths)
}

//Constructor for ContainerRuntime object. Constructor will fail if process wasn't found for container runtime.
func NewContainerRuntime(properties containerRuntimeProperties, configRootPath string) (*ContainerRuntimeInfo, error) {

	if properties.Supported == false {
		return nil, fmt.Errorf("Container runtime %s is not supported.", properties.Name)
	}
	cr := &ContainerRuntimeInfo{}
	cr.config.configRootPath = configRootPath
	cr.setProperties(&properties)
	_, err := cr.locateProcess()

	//if process wasn't find, fail to construct object
	if err != nil {
		return cr, fmt.Errorf("NewContainerRuntime - Failed to locate process for %s", cr.getName())
	}

	cr.updateCNIPaths()

	return cr, nil

}

// Try to get CNI paths from CR process.
// If there are multiple CR processes found - if takeLastIfMultiple is true, return the last one, otherwise nil with error.
// Search for process excludes cri-dockerd as if it is present there should be anyway another process for the main container runtime.
func getContainerRuntimeFromProcess(takeLastIfMultiple bool) (*ContainerRuntimeInfo, error) {

	//count processes
	sumProcess := 0
	lastObj := &ContainerRuntimeInfo{}

	for _, crp := range containersRuntimeProperties {
		if crp.Supported && crp.Name != CRIDOCKERD_CONTAINER_RUNTIME_NAME {
			cr_obj, err := NewContainerRuntime(crp, hostFileSystemDefaultLocation)
			if err == nil {
				sumProcess += 1
				lastObj = cr_obj
			}
		}
	}

	if sumProcess == 1 {
		zap.L().Debug("getContainerRuntimeFromProcess found one process",
			zap.String("CR_name", lastObj.getName()))
		return lastObj, nil
	}

	if sumProcess == 0 {
		return nil, fmt.Errorf("getContainerRuntimeFromProcess didn't find Container Runtime process")
	}

	// we got a process for more than 1 container runtimes
	if takeLastIfMultiple {
		zap.L().Debug("getContainerRuntimeFromProcess got more than one Container Runtime process, choosing last one",
			zap.String("CR_name", lastObj.getName()))
		return lastObj, nil
	}
	return nil, fmt.Errorf("getContainerRuntimeFromProcess got more than one Container Runtime process")

}

// Read Containerd specific config structure to extract CNI paths.
func extractContainerdCNIPathsFromConfig(configPath string, rootPath string) (*CNIPaths, error) {

	configPath = path.Join(rootPath, configPath)
	cni_paths := CNIPaths{}

	cni_config := struct {
		Plug map[string]struct {
			CR_Plugin struct {
				Bin_dir  string `toml:"bin_dir"`
				Conf_dir string `toml:"conf_dir"`
			} `toml:"cni"`
		} `toml:"plugins"`
	}{}

	_, err := toml.DecodeFile(configPath, &cni_config)

	if err != nil {
		zap.L().Error("getContainerdCNIPaths", zap.Error(err))
		return nil, err
	}

	cni_paths.Conf_dir = cni_config.Plug[CONTAINERD_CONFIG_SECTION].CR_Plugin.Conf_dir

	bin_dirs := cni_config.Plug[CONTAINERD_CONFIG_SECTION].CR_Plugin.Bin_dir

	if bin_dirs != "" {
		cni_paths.Bin_dirs = append(cni_paths.Bin_dirs, bin_dirs)
	}

	return &cni_paths, err
}

// Read cri-o specific config structure to extract CNI paths.
func extractCrioCNIPathsFromConfig(configPath string, rootPath string) (*CNIPaths, error) {
	configPath = path.Join(rootPath, configPath)

	cni_paths := CNIPaths{}

	cni_config := struct {
		Plug map[string]struct {
			// CRI_Plugin struct {
			Conf_dir string   `toml:"network_dir"`
			Bin_dir  []string `toml:"plugin_dirs"`
			// } `toml:"network"`
		} `toml:"crio"`
	}{}

	_, err := toml.DecodeFile(configPath, &cni_config)

	if err != nil {
		zap.L().Error("getCrioCNIPaths", zap.Error(err))
		return nil, err
	}

	cni_paths.Conf_dir = cni_config.Plug["network"].Conf_dir

	bin_dirs := cni_config.Plug["network"].Bin_dir

	if bin_dirs != nil {
		cni_paths.Bin_dirs = append(cni_paths.Bin_dirs, bin_dirs...)
	}

	return &cni_paths, nil
}

//Not implemented.
func extractCriDockedCNIPathsFromConfig(configPath string, rootPath string) (*CNIPaths, error) {
	return nil, fmt.Errorf("extractCriDockedCNIPathsFromConfig not implemented")
}

// Get CNI Paths from container runtime defined for kubelet.
// Container runtime is expected to be found in --container-runtime-endpoint.
func CNIPathsFromKubelet() (*CNIPaths, error) {
	proc, err := LocateProcessByExecSuffix(kubeletProcessSuffix)
	if err != nil {
		return nil, fmt.Errorf("failed to locate kube-proxy process: %w", err)
	}

	crEndpoint, crEndPointOK := proc.GetArg(kubeletContainerRuntimeEndPoint)

	if crEndpoint == "" {
		cr, crOK := proc.GetArg(kubeletContainerRuntime)

		if (!crEndPointOK && !crOK) || (cr != "remote") {
			// From docs: "If your nodes use Kubernetes v1.23 and earlier and these flags aren't present
			// or if the --container-runtime flag is not remote, you use the dockershim socket with Docker Engine."
			return nil, fmt.Errorf("no kubelet params or --container-runtime not 'remote' means dockershim.sock which is not supported")

		}
		//Uknown
		return nil, ErrCRNotFound

	}
	//there is crEndpoint
	zap.L().Debug("crEndPoint from kubelete found", zap.String("crEndPoint", crEndpoint))
	cr_obj, err := getContainerRuntime(crEndpoint)

	if err == nil {
		//Successfully created a Container runtime object. Try to get CNI paths
		cni_paths := cr_obj.getCNIPaths()

		if cni_paths == nil {
			// Didn't find CNIPaths.
			// Check specific case where the end point is cri-dockerd. If so, then in the absence of cni paths configuration for cri-dockerd process,
			// we check containerd (which is using cri-dockerd as a CRI plugin)

			if cr_obj.getName() == CRIDOCKERD_CONTAINER_RUNTIME_NAME {

				cr_obj, err := getContainerRuntime(CONTAINERD_SOCK)

				if err == nil {
					cni_paths := cr_obj.getCNIPaths()
					return cni_paths, nil
				}

			}
		}
		return cni_paths, nil
	}

	return nil, err

}
