package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/kubescape/go-logger"
	"github.com/kubescape/go-logger/helpers"
	"github.com/kubescape/host-scanner/sensor"
)

var BuildVersion string

func initHTTPHandlers() {
	// TODO: implement probe endpoint
	http.HandleFunc("/kubeletconfigurations", func(rw http.ResponseWriter, r *http.Request) {
		conf, err := sensor.SenseKubeletConfigurations()

		if err != nil {
			http.Error(rw, fmt.Sprintf("failed to sense kubelet conf: %v", err), http.StatusInternalServerError)
		} else {
			rw.WriteHeader(http.StatusOK)
			if _, err := rw.Write(conf); err != nil {
				logger.L().Ctx(r.Context()).Error("In kubeletConfigurations handler failed to write", helpers.Error(err))
			}
		}
	})
	http.HandleFunc("/kubeletcommandline", func(rw http.ResponseWriter, r *http.Request) {
		proc, err := sensor.LocateKubeletProcess()

		if err != nil {
			http.Error(rw, fmt.Sprintf("failed to sense kubelet conf: %v", err), http.StatusInternalServerError)
			return
		}

		cmdLine := strings.Join(proc.CmdLine, " ")
		rw.WriteHeader(http.StatusOK)
		if _, err := rw.Write([]byte(cmdLine)); err != nil {
			logger.L().Ctx(r.Context()).Error("In kubeletConfigurations handler failed to write", helpers.Error(err))
		}

	})
	http.HandleFunc("/osrelease", osReleaseHandler)
	http.HandleFunc("/kernelversion", kernelVersionHandler)
	http.HandleFunc("/linuxsecurityhardening", linuxSecurityHardeningHandler)
	http.HandleFunc("/openedports", openedPortsHandler)
	http.HandleFunc("/linuxkernelvariables", LinuxKernelVariablesHandler)
	http.HandleFunc("/kubeletinfo", kubeletInfoHandler)
	http.HandleFunc("/kubeproxyinfo", kubeProxyHandler)
	http.HandleFunc("/controlplaneinfo", controlPlaneHandler)
	http.HandleFunc("/cloudproviderinfo", cloudProviderHandler)
	http.HandleFunc("/version", versionHandler)
	http.HandleFunc("/cniinfo", CNIHandler)

}

func CNIHandler(rw http.ResponseWriter, r *http.Request) {
	resp, err := sensor.SenseCNIInfo(r.Context())
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
	resp, err := sensor.SenseControlPlaneInfo(r.Context())
	GenericSensorHandler(rw, r, resp, err, "SenseControlPlaneInfo")
}

func kubeProxyHandler(rw http.ResponseWriter, r *http.Request) {
	resp, err := sensor.SenseKubeProxyInfo(r.Context())
	GenericSensorHandler(rw, r, resp, err, "SenseKubeProxyInfo")
}

func kubeletInfoHandler(rw http.ResponseWriter, r *http.Request) {
	resp, err := sensor.SenseKubeletInfo(r.Context())
	GenericSensorHandler(rw, r, resp, err, "SenseKubeletInfo")
}

func LinuxKernelVariablesHandler(rw http.ResponseWriter, r *http.Request) {
	resp, err := sensor.SenseKernelVariables(r.Context())
	GenericSensorHandler(rw, r, resp, err, "SenseKernelVariables")
}

func openedPortsHandler(rw http.ResponseWriter, r *http.Request) {
	resp, err := sensor.SenseOpenPorts(r.Context())
	GenericSensorHandler(rw, r, resp, err, "SenseOpenPorts")
}

func osReleaseHandler(rw http.ResponseWriter, r *http.Request) {
	fileContent, err := sensor.SenseOsRelease()
	if err != nil {
		http.Error(rw, fmt.Sprintf("failed to SenseOsRelease: %v", err), http.StatusInternalServerError)
	} else {
		rw.WriteHeader(http.StatusOK)
		if _, err := rw.Write(fileContent); err != nil {
			logger.L().Ctx(r.Context()).Error("In SenseOsRelease handler failed to write", helpers.Error(err))
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
			logger.L().Ctx(r.Context()).Error("In kernelVersionHandler handler failed to write", helpers.Error(err))
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
			logger.L().Ctx(r.Context()).Error(fmt.Sprintf("In %s handler failed to write", senseName), helpers.Error(err))
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
		logger.L().Ctx(r.Context()).Error(fmt.Sprintf("In %s handler failed to write", senseName), helpers.Error(err))
	}
}
