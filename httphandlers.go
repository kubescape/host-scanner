package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/kubescape/go-logger"
	"github.com/kubescape/go-logger/helpers"
	"github.com/kubescape/host-scanner/sensor"
)

var (
	BuildVersion string
	healthzEP    = "/healthz"
	readyzEP     = "/readyz"
)

func initHTTPHandlers() {
	// setup readiness probe.
	isReady := &atomic.Value{}
	setupReadyz(isReady)

	// enable handlers for liveness and readiness probes.
	http.HandleFunc(healthzEP, healthzHandler)
	http.HandleFunc(readyzEP, readyzHandler(isReady))
	// WARNING: the below http requests are used by library: kubescape/core/pkg/hostsensorutils/hostsensorgetfrompod.go
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

// healthzHandler is a liveness probe.
func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write([]byte(`{"alive": true}`))
	if err != nil {
		logger.
			L().
			Ctx(r.Context()).
			Error("failed to write response")
	}
}

// setupReadyz set the atomic value to start checking the probe.
func setupReadyz(isReady *atomic.Value) {
	isReady.Store(false)
	go func() {
		logger.
			L().
			Ctx(context.Background()).
			Info("Setting up readyz probe")
		isReady.Store(true)
		logger.
			L().
			Ctx(context.Background()).
			Info("readyz probe is positive")
	}()
}

// readyzHandler is a readiness probe.
func readyzHandler(isReady *atomic.Value) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if isReady == nil || !isReady.Load().(bool) {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"ready": true}`))
		if err != nil {
			logger.
				L().
				Ctx(r.Context()).
				Error("failed to write response")
		}
	}
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
		return
	}
	rw.WriteHeader(http.StatusOK)
	if _, err := rw.Write(fileContent); err != nil {
		logger.L().Ctx(r.Context()).Error("In SenseOsRelease handler failed to write", helpers.Error(err))
	}
}

func kernelVersionHandler(rw http.ResponseWriter, r *http.Request) {
	fileContent, err := sensor.SenseKernelVersion()
	if err != nil {
		http.Error(rw, fmt.Sprintf("failed to sense kernelVersionHandler: %v", err), http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusOK)
	if _, err := rw.Write(fileContent); err != nil {
		logger.L().Ctx(r.Context()).Error("In kernelVersionHandler handler failed to write", helpers.Error(err))
	}
}

func linuxSecurityHardeningHandler(rw http.ResponseWriter, r *http.Request) {
	resp, err := sensor.SenseLinuxSecurityHardening()
	GenericSensorHandler(rw, r, resp, err, "sense linuxSecurityHardeningHandler")
}

// GenericSensorHandler do the generic job of encoding the response and error handeling
func GenericSensorHandler(w http.ResponseWriter, r *http.Request, respContent interface{}, respErr error, senseName string) {

	// Response ok
	if respErr == nil {
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(respContent); err != nil {
			logger.L().Ctx(r.Context()).Error(fmt.Sprintf("In %s handler failed to write", senseName), helpers.Error(err))
		}
		return
	}

	// Handle errors
	senseErr, ok := respErr.(*sensor.SenseError)
	if !ok {
		http.Error(w, fmt.Sprintf("failed to %s: %v", senseName, respErr), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(senseErr.Code)
	if err := json.NewEncoder(w).Encode(senseErr); err != nil {
		logger.L().Ctx(r.Context()).Error(fmt.Sprintf("In %s handler failed to write", senseName), helpers.Error(err))
	}
}
