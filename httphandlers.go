package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/kubescape/host-scanner/sensor"
	"go.uber.org/zap"
)

var BuildVersion string

func initHTTPHandlers() {
	// TODO: implement probe endpoint
	http.HandleFunc("/kubeletConfigurations", func(rw http.ResponseWriter, r *http.Request) {
		conf, err := sensor.SenseKubeletConfigurations()

		if err != nil {
			http.Error(rw, fmt.Sprintf("failed to sense kubelet conf: %v", err), http.StatusInternalServerError)
		} else {
			rw.WriteHeader(http.StatusOK)
			if _, err := rw.Write(conf); err != nil {
				zap.L().Error("In kubeletConfigurations handler failed to write", zap.Error(err))
			}
		}
	})
	http.HandleFunc("/kubeletCommandLine", func(rw http.ResponseWriter, r *http.Request) {
		proc, err := sensor.LocateKubeletProcess()

		if err != nil {
			http.Error(rw, fmt.Sprintf("failed to sense kubelet conf: %v", err), http.StatusInternalServerError)
			return
		}

		cmdLine := strings.Join(proc.CmdLine, " ")
		rw.WriteHeader(http.StatusOK)
		if _, err := rw.Write([]byte(cmdLine)); err != nil {
			zap.L().Error("In kubeletConfigurations handler failed to write", zap.Error(err))
		}

	})
	http.HandleFunc("/osRelease", osReleaseHandler)
	http.HandleFunc("/kernelVersion", kernelVersionHandler)
	http.HandleFunc("/linuxSecurityHardening", linuxSecurityHardeningHandler)
	http.HandleFunc("/openedPorts", openedPortsHandler)
	http.HandleFunc("/LinuxKernelVariables", LinuxKernelVariablesHandler)
	http.HandleFunc("/kubeletInfo", kubeletInfoHandler)
	http.HandleFunc("/kubeProxyInfo", kubeProxyHandler)
	http.HandleFunc("/controlPlaneInfo", controlPlaneHandler)
	http.HandleFunc("/cloudProviderInfo", cloudProviderHandler)
	http.HandleFunc("/version", versionHandler)
	http.HandleFunc("/CNIInfo", CNIHandler)

}

func CNIHandler(rw http.ResponseWriter, r *http.Request) {
	resp, err := sensor.SenseCNIInfo()
	GenericSensorHandler(rw, r, resp, err, "SenseCNIInfo")
}

func versionHandler(rw http.ResponseWriter, r *http.Request) {
	var err error
	if BuildVersion == "" {
		err = fmt.Errorf("host scanner BuildVersion is empty")
		BuildVersion = "unknown"
	}
	resp := BuildVersion
	GenericSensorHandler(rw, r, resp, err, "VersionHandler")
}

func cloudProviderHandler(rw http.ResponseWriter, r *http.Request) {
	resp, err := sensor.SenseCloudProviderInfo()
	GenericSensorHandler(rw, r, resp, err, "SenseCloudProviderInfo")
}

func controlPlaneHandler(rw http.ResponseWriter, r *http.Request) {
	resp, err := sensor.SenseControlPlaneInfo()
	GenericSensorHandler(rw, r, resp, err, "SenseControlPlaneInfo")
}

func kubeProxyHandler(rw http.ResponseWriter, r *http.Request) {
	resp, err := sensor.SenseKubeProxyInfo()
	GenericSensorHandler(rw, r, resp, err, "SenseKubeProxyInfo")
}

func kubeletInfoHandler(rw http.ResponseWriter, r *http.Request) {
	resp, err := sensor.SenseKubeletInfo()
	GenericSensorHandler(rw, r, resp, err, "SenseKubeletInfo")
}

func LinuxKernelVariablesHandler(rw http.ResponseWriter, r *http.Request) {
	resp, err := sensor.SenseKernelVariables()
	GenericSensorHandler(rw, r, resp, err, "SenseKernelVariables")
}

func openedPortsHandler(rw http.ResponseWriter, r *http.Request) {
	resp, err := sensor.SenseOpenPorts()
	GenericSensorHandler(rw, r, resp, err, "SenseOpenPorts")
}

func osReleaseHandler(rw http.ResponseWriter, r *http.Request) {
	fileContent, err := sensor.SenseOsRelease()
	if err != nil {
		http.Error(rw, fmt.Sprintf("failed to SenseOsRelease: %v", err), http.StatusInternalServerError)
	} else {
		rw.WriteHeader(http.StatusOK)
		if _, err := rw.Write(fileContent); err != nil {
			zap.L().Error("In SenseOsRelease handler failed to write", zap.Error(err))
		}
	}
}

func kernelVersionHandler(rw http.ResponseWriter, r *http.Request) {
	fileContent, err := sensor.SenseKernelVersion()
	if err != nil {
		http.Error(rw, fmt.Sprintf("failed to sense kernelVersionHandler: %v", err), http.StatusInternalServerError)
	} else {
		rw.WriteHeader(http.StatusOK)
		if _, err := rw.Write(fileContent); err != nil {
			zap.L().Error("In kernelVersionHandler handler failed to write", zap.Error(err))
		}
	}
}

func linuxSecurityHardeningHandler(rw http.ResponseWriter, r *http.Request) {
	resp, err := sensor.SenseLinuxSecurityHardening()
	GenericSensorHandler(rw, r, resp, err, "sense linuxSecurityHardeningHandler")
}

// GenericSensorHandler do the generic job of encoding the response and error handeling
func GenericSensorHandler(w http.ResponseWriter, r *http.Request, respContent interface{}, err error, senseName string) {

	// Response ok
	if err == nil {
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(respContent); err != nil {
			zap.L().Error(fmt.Sprintf("In %s handler failed to write", senseName), zap.Error(err))
		}
		return
	}

	// Handle errors
	senseErr, ok := err.(*sensor.SenseError)
	if !ok {
		http.Error(w, fmt.Sprintf("failed to %s: %v", senseName, err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(senseErr.Code)
	if err := json.NewEncoder(w).Encode(senseErr); err != nil {
		zap.L().Error(fmt.Sprintf("In %s handler failed to write", senseName), zap.Error(err))
	}
}
