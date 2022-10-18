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
)

//Constant values for different types of Container Runtimes
const (
	//container runtimes
	containerdContainerRuntimeName = "containerd"
	containerdSock                 = "/containerd.sock"
	containerdConfigSection        = "io.containerd.grpc.v1.cri"

	crioContainerRuntimeName = "crio"
	crioSock                 = "/crio.sock"

	dockershimSock = "/dockershim.sock"

	//container runtime interfaces
	cridockerdContainerRuntimeName = "cri-dockerd"
	cridockerdSock                 = "/cri-dockerd.sock"
)

//CNIPath struct
type CNIPaths struct {

	//Where we found the paths. Currently not in use.
	Source string `json:"Source,omitempty"`

	//The location of the CNI config files
	Conf_dir string `json:"CNIPathsConfigDir,omitempty"`

	//The location(s) of the binaries. It is a list because some of the CRs configure more than one dir as list.
	Bin_dirs []string `json:"CNIPathsBinDirs,omitempty"`
}

//General properties for container runtimes
type containerRuntimeProperties struct {
	Name string

	//if false, container runtime constructor will fail
	Supported bool

	//whether container runtime supports config files (cri-dockerd is an example doesn't support it)
	ConfigSupported bool

	//Process for cuflag stom config file.
	ConfigArgName string

	//Process flag for custom configuration directory.
	ConfigDirArgName string

	//Default config path
	DefaultConfigPath string

	//default configuration directory
	DefaultConfigDir string

	//suffix of container runtime process
	ProcessSuffix string

	//the socket suffix - used to identify the container runtime from kubelet
	Socket string

	//process pararm for CNI configuration directory
	CNIConfigDirArgName string

	//Process flag for CNI plugins directory
	CNIPluginDirArgName string

	//extract CNI info function
	ParseCNIFromConfigFunc func(string) (*CNIPaths, error)
}

// Struct to hold configurations of the container runtime
// Currently holds only CNIPaths, can be later expanded to additional properties.
type containerRuntimeConfigurations struct {

	//CNI files paths
	CNI_files *CNIPaths
}

//Struct to hold all information of a container runtime
type ContainerRuntimeInfo struct {
	properties *containerRuntimeProperties

	// process pointer
	process *ProcessDetails

	//CR onfig information if exist.
	config containerRuntimeConfigurations

	//root
	rootDir string
}

//list of container runtime properties.
var containersRuntimeProperties = []containerRuntimeProperties{
	{
		Name:                   containerdContainerRuntimeName,
		Supported:              true,
		ConfigSupported:        true,
		DefaultConfigPath:      "/etc/containerd/config.toml",
		ProcessSuffix:          "/containerd",
		Socket:                 "/containerd.sock",
		ConfigArgName:          "--config",
		ConfigDirArgName:       "",
		DefaultConfigDir:       "/etc/containerd/containerd.conf.d",
		CNIConfigDirArgName:    "",
		CNIPluginDirArgName:    "",
		ParseCNIFromConfigFunc: parseCNIPathsFromConfig_containerd,
	},
	{
		Name:                   crioContainerRuntimeName,
		Supported:              true,
		ConfigSupported:        true,
		DefaultConfigPath:      "/etc/crio/crio.conf",
		ProcessSuffix:          "/crio",
		Socket:                 "/crio.sock",
		ConfigArgName:          "--config",
		ConfigDirArgName:       "--config-dir",
		DefaultConfigDir:       "/etc/crio/crio.conf.d",
		CNIConfigDirArgName:    "--cni-config-dir",
		CNIPluginDirArgName:    "--cni-plugin-dir",
		ParseCNIFromConfigFunc: parseCNIPathsFromConfig_crio,
	},
	{
		Name:                   cridockerdContainerRuntimeName,
		Supported:              true,
		ConfigSupported:        false,
		DefaultConfigPath:      "",
		ProcessSuffix:          "/cri-dockerd",
		Socket:                 "/cri-dockerd.sock",
		ConfigArgName:          "",
		ConfigDirArgName:       "",
		DefaultConfigDir:       "",
		CNIConfigDirArgName:    "--cni-conf-dir",
		CNIPluginDirArgName:    "--cni-bin-dir",
		ParseCNIFromConfigFunc: parseCNIPathsFromConfig_cridockerd,
	},
}

//Get default CNIPaths - in use in case we couldnt find the paths through configs.
func getCNIPathsDefault() *CNIPaths {
	bin_dirs := []string{CNIDefaultBinDir}
	return &CNIPaths{Conf_dir: CNIDefaultConfigDir, Bin_dirs: bin_dirs}
}

// Get CNI paths from running Container Runtimes. Flow:
// 1. Find CNI through kubelet flag (--container-runtime_endpoint). If not found:
// 2. Find CNI through process of supported container runtimes. If not found:
// 3. return CNI default paths.
func getContainerRuntimeCNIPaths() *CNIPaths {

	// //Attempting to find CR from kubelet.
	cni_paths := CNIPathsFromKubelet()

	if cni_paths == nil {
		//Could construct container runtime from kubelet
		zap.L().Debug("getContainerRuntimeCNIPaths - failed to get CNI Paths through kubelete, trying through process")

		//Attempting to find CR through process.
		cr, err := getContainerRuntimeFromProcess(false)

		if err == nil {
			cni_paths := cr.getCNIPaths()
			if cni_paths == nil {
				return getCNIPathsDefault()
			} else {
				return cni_paths
			}

		} else {
			//Failed to get container runtime from process
			zap.L().Debug("getContainerRuntimeCNIPaths - failed to get container runtime from process, return cni defaults",
				zap.Error(err))

			return getCNIPathsDefault()
		}
	} else {
		if cni_paths == nil {
			return getCNIPathsDefault()
		} else {
			return cni_paths
		}

	}

}

//Get/Set functions
func (cr *ContainerRuntimeInfo) getDefaultConfigPath() string {
	return cr.properties.DefaultConfigPath
}

func (cr *ContainerRuntimeInfo) getCNIPaths() *CNIPaths { return cr.config.CNI_files }

func (cr *ContainerRuntimeInfo) setCNIPaths(cni_paths *CNIPaths) { cr.config.CNI_files = cni_paths }

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

func (cr *ContainerRuntimeInfo) getConfigPath() string {
	configPath := cr.getArgFromProcess(cr.properties.ConfigArgName)
	if configPath == "" {
		zap.L().Debug("getConfigPath - custom config no found through process, taking default config path",
			zap.String("CR_name", cr.properties.Name),
			zap.String("defaultConfigPath", cr.getDefaultConfigPath()))
		configPath = cr.getDefaultConfigPath()

	} else {
		zap.L().Debug("getCNIPathsFromConfig - custom config found in process",
			zap.String("CR_name", cr.properties.Name),
			zap.String("configPath", configPath))
	}

	configPath = path.Join(cr.rootDir, configPath)
	return configPath
}

func (cr *ContainerRuntimeInfo) getArgFromProcess(argName string) string {
	if argName == "" {
		return ""
	}

	p := cr.process
	if p != nil {
		res, ok := p.GetArg(argName)
		if !ok || res == "" {
			return ""
		} else {
			return res
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

	var configDirFilesFullPath []string

	if !cr.properties.ConfigSupported {
		return nil, nil
	}

	//Getting all config files in drop in folder if exist.
	configDirPath := cr.getConfigDirPath()
	configDirFilesFullPath = makeConfigFilesList(configDirPath)

	if configPath == "" {
		configPath = cr.getConfigPath()
	}

	//appding config file to the end of the list as it always has the lowest priority.
	if configPath != "" {
		configDirFilesFullPath = append(configDirFilesFullPath, configPath)
	}

	cni_paths := getCNIPathsFromConfigPaths(configDirFilesFullPath, cr.properties.ParseCNIFromConfigFunc)

	// if err != nil {
	// 	zap.L().Debug("getCNIPathsFromConfig - error looking for cni paths from configs",
	// 		zap.String("CR_name", cr.properties.Name),
	// 		zap.Error(err))
	// }

	if cni_paths != nil {
		zap.L().Debug("getCNIPathsFromConfig - found cni paths in configs for container runtime",
			zap.String("CR_name", cr.properties.Name))
	}

	return cni_paths, nil

}

//Get CNI Paths from process cmdline flags if such defined.
func (cr *ContainerRuntimeInfo) getCNIPathsFromProcess() *CNIPaths {
	p := cr.process
	if p == nil {
		zap.L().Debug("getCNIPathsFromProccess no process found for containe runtime",
			zap.String("ContainerRuntime name", cr.properties.Name))
		return nil
	}

	var conf_dir string
	var bin_dirs []string

	if cr.properties.CNIConfigDirArgName != "" {
		conf_dir, ok := p.GetArg(cr.properties.CNIConfigDirArgName)

		if !ok || conf_dir == "" {
			zap.L().Debug("getCNIPathsFromProccess no cni config dir found for process",
				zap.String("ContainerRuntime name", cr.properties.Name),
				zap.String("CNIConfigDirArgName", cr.properties.CNIConfigDirArgName))
		}
	}

	if cr.properties.CNIPluginDirArgName != "" {
		bin_dir, ok := p.GetArg(cr.properties.CNIPluginDirArgName)

		if !ok || bin_dir == "" {
			zap.L().Debug("getCNIPathsFromProccess no cni plugin dir found for process",
				zap.String("ContainerRuntime name", cr.properties.Name),
				zap.String("CNIConfigDirArgName", cr.properties.CNIPluginDirArgName))
		} else {
			if bin_dir != "" {
				bin_dirs = append(bin_dirs, bin_dir)
			}
		}
	}

	if len(bin_dirs) == 0 && conf_dir == "" {
		return nil
	} else {
		return &CNIPaths{Conf_dir: conf_dir, Bin_dirs: bin_dirs}
	}

}

// update CNI paths property. Flow:
// 1. Try to get paths from process flags. If not found:
// 2. Try to get paths from config file. If not found:
// 3. return defaults
func (cr *ContainerRuntimeInfo) updateCNIPaths() {

	var err error
	var cni_paths *CNIPaths
	// var cni_paths_source string

	CR_name := cr.properties.Name

	cni_paths = cr.getCNIPathsFromProcess()

	if cni_paths == nil {
		zap.L().Debug("updateCNIPaths couldn't get cni paths from process, trying through configs", zap.String("ContainerRuntime", CR_name))

		cni_paths, err = cr.getCNIPathsFromConfig("")

		if err != nil {
			//If didn't succeed to get cni paths from config, return global.
			zap.L().Debug("updateCNIPaths Failed to get paths from config", zap.String("ContainerRuntime", CR_name), zap.Error(err))
		}

		if cni_paths != nil {
			zap.L().Debug("updateCNIPaths found CNIPaths in configs", zap.String("ContainerRuntime", CR_name))
		}
	} else {
		zap.L().Debug("updateCNIPaths found CNIPaths in process",
			zap.String("ContainerRuntime", CR_name))
	}

	if cni_paths != nil {

		// Found config file, checking that paths are not empty.
		if cni_paths.Conf_dir == "" {
			zap.L().Debug("updateCNIPaths ContainerRuntime has config without conf_dir definition, taking default",
				zap.String("ContainerRuntime", CR_name),
			)
			cni_paths.Conf_dir = CNIDefaultConfigDir
		}

		if cni_paths.Bin_dirs == nil {
			zap.L().Debug("updateCNIPaths ContainerRuntime has config without Bin_dirs definition, taking default",
				zap.String("ContainerRuntime", CR_name),
			)
			cni_paths.Bin_dirs = append(cni_paths.Bin_dirs, CNIDefaultBinDir)
		}

	}

	cr.setCNIPaths(cni_paths)
}

//Constructor for ContainerRuntime object. Constructor will fail if process wasn't found for container runtime.
func NewContainerRuntime(properties containerRuntimeProperties, root_dir string) (*ContainerRuntimeInfo, error) {

	if properties.Supported == false {
		return nil, fmt.Errorf("Container runtime %s is not supported.", properties.Name)
	}
	cr := &ContainerRuntimeInfo{}
	cr.rootDir = root_dir
	cr.setProperties(&properties)
	p, err := LocateProcessByExecSuffix(cr.properties.ProcessSuffix)

	//if process wasn't find, fail to construct object
	if err != nil {
		return cr, fmt.Errorf("NewContainerRuntime - Failed to locate process for %s", cr.properties.Name)
	} else {
		cr.process = p
	}

	cr.updateCNIPaths()

	return cr, nil

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

// Try to get CNI paths from CR process.
// If there are multiple CR processes found - if takeLastIfMultiple is true, return the last one, otherwise nil with error.
// Search for process excludes cri-dockerd as if it is present there should be anyway another process for the main container runtime.
func getContainerRuntimeFromProcess(takeLastIfMultiple bool) (*ContainerRuntimeInfo, error) {

	//count processes
	sumProcess := 0
	lastObj := &ContainerRuntimeInfo{}

	for _, crp := range containersRuntimeProperties {
		if crp.Supported && crp.Name != cridockerdContainerRuntimeName {
			cr_obj, err := NewContainerRuntime(crp, hostFileSystemDefaultLocation)
			if err == nil {
				sumProcess += 1
				lastObj = cr_obj
			}
		}
	}

	if sumProcess == 1 {
		zap.L().Debug("getContainerRuntimeFromProcess found one process",
			zap.String("CR_name", lastObj.properties.Name))
		return lastObj, nil
	}

	if sumProcess == 0 {
		return nil, fmt.Errorf("getContainerRuntimeFromProcess didn't find Container Runtime process")
	}

	// we got a process for more than 1 container runtimes
	if takeLastIfMultiple {
		zap.L().Debug("getContainerRuntimeFromProcess got more than one Container Runtime process, choosing last one",
			zap.String("CR_name", lastObj.properties.Name))
		return lastObj, nil
	}
	return nil, fmt.Errorf("getContainerRuntimeFromProcess got more than one Container Runtime process")

}

// Read Containerd specific config structure to extract CNI paths.
func parseCNIPathsFromConfig_containerd(configPath string) (*CNIPaths, error) {

	// configPath = path.Join(rootPath, configPath)
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
		return &cni_paths, err
	}

	cni_paths.Conf_dir = cni_config.Plug[containerdConfigSection].CR_Plugin.Conf_dir

	bin_dirs := cni_config.Plug[containerdConfigSection].CR_Plugin.Bin_dir

	if bin_dirs != "" {
		cni_paths.Bin_dirs = append(cni_paths.Bin_dirs, bin_dirs)
	}

	return &cni_paths, err
}

// Read cri-o specific config structure to extract CNI paths.
func parseCNIPathsFromConfig_crio(configPath string) (*CNIPaths, error) {
	// configPath = path.Join(rootPath, configPath)

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
func parseCNIPathsFromConfig_cridockerd(configPath string) (*CNIPaths, error) {
	return nil, fmt.Errorf("parseCNIPathsFromConfig_cridockerd not implemented")
}

// Get CNI Paths from container runtime defined for kubelet.
// Container runtime is expected to be found in --container-runtime-endpoint.
func CNIPathsFromKubelet() *CNIPaths {
	proc, err := LocateProcessByExecSuffix(kubeletProcessSuffix)
	if err != nil {
		zap.L().Debug("CNIPathsFromKubelet - failed to locate kube-proxy process")
		return nil
	}

	crEndpoint, crEndPointOK := proc.GetArg(kubeletContainerRuntimeEndPoint)

	if crEndpoint == "" {
		cr, crOK := proc.GetArg(kubeletContainerRuntime)

		if (!crEndPointOK && !crOK) || (cr != "remote") {
			// From docs: "If your nodes use Kubernetes v1.23 and earlier and these flags aren't present
			// or if the --container-runtime flag is not remote, you use the dockershim socket with Docker Engine."
			zap.L().Debug("CNIPathsFromKubelet - no kubelet flags or --container-runtime not 'remote' means dockershim.sock which is not supported")
			return nil

		}
		//Uknown
		zap.L().Debug(ErrCRNotFound.Error())
		return nil

	}
	//there is crEndpoint
	zap.L().Debug("crEndPoint from kubelete found", zap.String("crEndPoint", crEndpoint))
	cr_obj, err := getContainerRuntime(crEndpoint)

	if err != nil {
		return nil
	} else {
		//Successfully created a Container runtime object. Try to get CNI paths
		cni_paths := cr_obj.getCNIPaths()

		if cni_paths == nil {
			// Didn't find CNIPaths.
			// Check specific case where the end point is cri-dockerd. If so, then in the absence of cni paths configuration for cri-dockerd process,
			// we check containerd (which is using cri-dockerd as a CRI plugin)

			if cr_obj.properties.Name == cridockerdContainerRuntimeName {

				cr_obj, err := getContainerRuntime(containerdSock)

				if err == nil {
					cni_paths := cr_obj.getCNIPaths()
					return cni_paths
				}

			}
		}
		return cni_paths
	}

}

// Get a list of configpaths and a parsing function and returns CNIPaths.
// iteration is done by the original order of the configpaths.
func getCNIPathsFromConfigPaths(configPaths []string, parseFunc func(string) (*CNIPaths, error)) *CNIPaths {

	conf_dir := ""
	bin_dirs := []string{}

	for _, config_path := range configPaths {
		cni_paths, err := parseFunc(config_path)

		if err != nil {
			zap.L().Debug("getCNIPathsFromConfigPaths - Failed to parse config file", zap.String("config_path", config_path))
		} else {
			if cni_paths.Conf_dir != "" && conf_dir == "" {
				conf_dir = cni_paths.Conf_dir
			}

			if len(cni_paths.Bin_dirs) > 0 && len(bin_dirs) == 0 {
				bin_dirs = cni_paths.Bin_dirs
			}

			if conf_dir != "" && len(bin_dirs) > 0 {
				return &CNIPaths{Conf_dir: conf_dir, Bin_dirs: bin_dirs}
			}

		}

	}

	return &CNIPaths{}

}