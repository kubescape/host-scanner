package sensor

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	ds "github.com/armosec/host-sensor/sensor/datastructures"
	"github.com/armosec/host-sensor/sensor/internal/utils"
)

const (
	apiServerExe                   = "/kube-apiserver"
	controllerManagerExe           = "/kube-controller-manager"
	schedulerExe                   = "/kube-scheduler"
	etcdExe                        = "/etcd"
	etcdDataDirArg                 = "--data-dir"
	apiEncryptionProviderConfigArg = "--encryption-provider-config"
	auditPolicyFileArg             = "--audit-policy-file"

	// Default files paths according to https://workbench.cisecurity.org/benchmarks/8973/sections/1126652
	apiServerSpecsPath          = "/etc/kubernetes/manifests/kube-apiserver.yaml"
	controllerManagerSpecsPath  = "/etc/kubernetes/manifests/kube-controller-manager.yaml"
	controllerManagerConfigPath = "/etc/kubernetes/controller-manager.conf"
	schedulerSpecsPath          = "/etc/kubernetes/manifests/kube-scheduler.yaml"
	schedulerConfigPath         = "/etc/kubernetes/scheduler.conf"
	etcdConfigPath              = "/etc/kubernetes/manifests/etcd.yaml"
	adminConfigPath             = "/etc/kubernetes/admin.conf"
	pkiDir                      = "/etc/kubernetes/pki"

	// TODO: cni
)

var (
	ErrDataDirNotFound = errors.New("failed to find etcd data-dir")
)

// KubeProxyInfo holds information about kube-proxy process
type ControlPlaneInfo struct {
	APIServerInfo         *ApiServerInfo  `json:"APIServerInfo,omitempty"`
	ControllerManagerInfo *K8sProcessInfo `json:"controllerManagerInfo,omitempty"`
	SchedulerInfo         *K8sProcessInfo `json:"schedulerInfo,omitempty"`
	EtcdConfigFile        *ds.FileInfo    `json:"etcdConfigFile,omitempty"`
	EtcdDataDir           *ds.FileInfo    `json:"etcdDataDir,omitempty"`
	AdminConfigFile       *ds.FileInfo    `json:"adminConfigFile,omitempty"`
	PKIDIr                *ds.FileInfo    `json:"PKIDir,omitempty"`
	PKIFiles              []*ds.FileInfo  `json:"PKIFiles,omitempty"`
	CNIConfigFiles        []*ds.FileInfo  `json:"CNIConfigFiles,omitempty"`

	// The name of the running CNI
	CNIName string `json:"CNIName,omitempty"`
}

// K8sProcessInfo holds information about a k8s process
type K8sProcessInfo struct {
	// Information about the process specs file (if relevant)
	SpecsFile *ds.FileInfo `json:"specsFile,omitempty"`

	// Information about the process config file (if relevant)
	ConfigFile *ds.FileInfo `json:"configFile,omitempty"`

	// Information about the process kubeconfig file (if relevant)
	KubeConfigFile *ds.FileInfo `json:"kubeConfigFile,omitempty"`

	// Information about the process client ca file (if relevant)
	ClientCAFile *ds.FileInfo `json:"clientCAFile,omitempty"`

	// Raw cmd line of the process
	CmdLine string `json:"cmdLine"`
}

type ApiServerInfo struct {
	EncryptionProviderConfigFile *ds.FileInfo `json:"encryptionProviderConfigFile,omitempty"`
	AuditPolicyFile              *ds.FileInfo `json:"auditPolicyFile,omitempty"`
	*K8sProcessInfo              `json:",inline"`
}

// getEtcdDataDir find the `data-dir` path of etcd k8s component
func getEtcdDataDir() (string, error) {

	proc, err := utils.LocateProcessByExecSuffix(etcdExe)
	if err != nil {
		return "", fmt.Errorf("failed to locate kube-proxy process: %w", err)
	}

	dataDir, ok := proc.GetArg(etcdDataDirArg)
	if !ok || dataDir == "" {
		return "", ErrDataDirNotFound
	}

	return dataDir, nil
}

func makeProcessInfoVerbose(p *utils.ProcessDetails, specsPath, configPath, kubeConfigPath, clientCaPath string) *K8sProcessInfo {
	ret := K8sProcessInfo{}

	// init files
	files := []struct {
		data **ds.FileInfo
		path string
		file string
	}{
		{&ret.SpecsFile, specsPath, "specs"},
		{&ret.ConfigFile, configPath, "config"},
		{&ret.KubeConfigFile, kubeConfigPath, "kubeconfig"},
		{&ret.ClientCAFile, clientCaPath, "calient ca certificate"},
	}

	// get data
	for i := range files {
		file := &files[i]
		if file.path == "" {
			continue
		}

		*file.data = makeHostFileInfoVerbose(file.path, false,
			zap.String("in", "makeProcessInfoVerbose"),
			zap.String("path", file.path),
		)
	}

	if p != nil {
		ret.CmdLine = p.RawCmd()
	}

	// Return `nil` if wasn't able to find any data
	if ret == (K8sProcessInfo{}) {
		return nil
	}

	return &ret
}

// makeAPIserverEncryptionProviderConfigFile returns a ds.FileInfo object for the encryption provider config file of the API server. Required for https://workbench.cisecurity.org/sections/1126663/recommendations/1838675
func makeAPIserverEncryptionProviderConfigFile(p *utils.ProcessDetails) *ds.FileInfo {
	encryptionProviderConfigPath, ok := p.GetArg(apiEncryptionProviderConfigArg)
	if !ok {
		zap.L().Warn("failed to find encryption provider config path", zap.String("in", "makeAPIserverEncryptionProviderConfigFile"))
		return nil
	}

	fi, err := utils.MakeContaineredFileInfo(p, encryptionProviderConfigPath, true)
	if err != nil {
		zap.L().Warn("failed to create encryption provider config file info", zap.Error(err))
		return nil
	}

	// remove sensitive data
	data := map[string]interface{}{}
	err = yaml.Unmarshal(fi.Content, &data)
	if err != nil {
		err = json.Unmarshal(fi.Content, &data)
		if err != nil {
			zap.L().Warn("failed to unmarshal encryption provider config file")
			return nil
		}
	}

	removeEncryptionProviderConfigSecrets(data)

	// marshal back to yaml
	fi.Content, err = yaml.Marshal(data)
	if err != nil {
		zap.L().Warn("failed to marshal encryption provider config file", zap.Error(err))
		return nil
	}

	return fi
}

func removeEncryptionProviderConfigSecrets(data map[string]interface{}) {
	resources, ok := data["resources"].([]interface{})
	if !ok {
		return
	}

	for i := range resources {
		resource, ok := resources[i].(map[string]interface{})
		if !ok {
			continue
		}

		providers, ok := resource["providers"].([]interface{})
		if !ok {
			continue
		}

		for j := range providers {
			provider, ok := providers[j].(map[string]interface{})
			if !ok {
				continue
			}

			for key := range provider {
				object, ok := provider[key].(map[string]interface{})
				if !ok {
					continue
				}
				keys, ok := object["keys"].([]interface{})
				if !ok {
					continue
				}
				for k := range keys {
					key, ok := keys[k].(map[string]interface{})
					if !ok {
						continue
					}
					key["secret"] = "<REDACTED>"
					keys[k] = key
				}
				object["keys"] = keys
				provider[key] = object
			}
			providers[j] = provider
		}
		resource["providers"] = providers
		resources[i] = resource
	}
	data["resources"] = resources
}

// makeAPIserverAuditPolicyFile returns a ds.FileInfo object for an audit policy file of the API server. Required for https://workbench.cisecurity.org/sections/1126663/recommendations/1838675
func makeAPIserverAuditPolicyFile(p *utils.ProcessDetails) *ds.FileInfo {
	auditPolicyFilePath, ok := p.GetArg(auditPolicyFileArg)
	if !ok {
		zap.L().Info("audit-policy-file argument was not set ", zap.String("in", "makeAPIserverAuditPolicyFile"))
		return nil
	}

	return makeContaineredFileInfoVerbose(p, auditPolicyFilePath, true,
		zap.String("in", "makeAPIserverAuditPolicyFile"),
	)
}

// SenseControlPlaneInfo return `ControlPlaneInfo`
func SenseControlPlaneInfo() (*ControlPlaneInfo, error) {
	var err error
	ret := ControlPlaneInfo{}

	debugInfo := zap.String("in", "SenseControlPlaneInfo")

	apiProc, err := utils.LocateProcessByExecSuffix(apiServerExe)
	if err == nil {
		ret.APIServerInfo = &ApiServerInfo{}
		ret.APIServerInfo.K8sProcessInfo = makeProcessInfoVerbose(apiProc, apiServerSpecsPath, "", "", "")
		ret.APIServerInfo.EncryptionProviderConfigFile = makeAPIserverEncryptionProviderConfigFile(apiProc)
		ret.APIServerInfo.AuditPolicyFile = makeAPIserverAuditPolicyFile(apiProc)
	} else {
		zap.L().Error("SenseControlPlaneInfo", zap.Error(err))
	}

	controllerMangerProc, err := utils.LocateProcessByExecSuffix(controllerManagerExe)
	if err == nil {
		ret.ControllerManagerInfo = makeProcessInfoVerbose(controllerMangerProc, controllerManagerSpecsPath, controllerManagerConfigPath, "", "")
	} else {
		zap.L().Error("SenseControlPlaneInfo", zap.Error(err))
	}

	SchedulerProc, err := utils.LocateProcessByExecSuffix(schedulerExe)
	if err == nil {
		ret.SchedulerInfo = makeProcessInfoVerbose(SchedulerProc, schedulerSpecsPath, schedulerConfigPath, "", "")
	} else {
		zap.L().Error("SenseControlPlaneInfo", zap.Error(err))
	}

	// EtcdConfigFile
	ret.EtcdConfigFile = makeHostFileInfoVerbose(etcdConfigPath,
		false,
		debugInfo,
		zap.String("component", "EtcdConfigFile"),
	)

	// AdminConfigFile
	ret.AdminConfigFile = makeHostFileInfoVerbose(adminConfigPath,
		false,
		debugInfo,
		zap.String("component", "AdminConfigFile"),
	)

	// PKIDIr
	ret.PKIDIr = makeHostFileInfoVerbose(pkiDir,
		false,
		debugInfo,
		zap.String("component", "PKIDIr"),
	)

	// PKIFiles
	ret.PKIFiles, err = makeHostDirFilesInfoVerbose(pkiDir, true, nil, 0)
	if err != nil {
		zap.L().Error("SenseControlPlaneInfo failed to get PKIFiles info", zap.Error(err))
	}

	// etcd data-dir
	etcdDataDir, err := getEtcdDataDir()
	if err != nil {
		zap.L().Error("SenseControlPlaneInfo", zap.Error(ErrDataDirNotFound))
	} else {
		ret.EtcdDataDir = makeHostFileInfoVerbose(etcdDataDir,
			false,
			debugInfo,
			zap.String("component", "EtcdDataDir"),
		)
	}

	// make cni config files
	CNIConfigInfo, err := makeCNIConfigFilesInfo()

	if err != nil {
		zap.L().Error("SenseControlPlaneInfo", zap.Error(err))
	} else {
		ret.CNIConfigFiles = CNIConfigInfo
	}

	// get CNI name
	ret.CNIName = getCNIName()

	// If wasn't able to find any data - this is not a control plane
	if ret.APIServerInfo == nil &&
		ret.ControllerManagerInfo == nil &&
		ret.SchedulerInfo == nil &&
		ret.EtcdConfigFile == nil &&
		ret.EtcdDataDir == nil &&
		ret.AdminConfigFile == nil {
		return nil, &SenseError{
			Massage:  "not a control plane node",
			Function: "SenseControlPlaneInfo",
			Code:     http.StatusNotFound,
		}
	}

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
		zap.L().Debug("SenseControlPlaneInfo - no cni config files were found.",
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
