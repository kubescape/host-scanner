package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/armosec/host-sensor/sensor"
	"go.uber.org/zap"
)

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

		cmdLine := strings.Join(proc.CmdLine, " ")
		if err != nil {
			http.Error(rw, fmt.Sprintf("failed to sense kubelet conf: %v", err), http.StatusInternalServerError)
		} else {
			rw.WriteHeader(http.StatusOK)
			if _, err := rw.Write([]byte(cmdLine)); err != nil {
				zap.L().Error("In kubeletConfigurations handler failed to write", zap.Error(err))
			}
		}
	})
	http.HandleFunc("/osRelease", osReleaseHandler)
	http.HandleFunc("/kernelVersion", kernelVersionHandler)
	http.HandleFunc("/linuxSecurityHardening", linuxSecurityHardeningHandler)
	http.HandleFunc("/openedPorts", openedPortsHandler)
	http.HandleFunc("/LinuxKernelVariables", LinuxKernelVariablesHandler)
}

func LinuxKernelVariablesHandler(rw http.ResponseWriter, r *http.Request) {
	respContent, err := sensor.SenseKernelVariables()
	if err != nil {
		http.Error(rw, fmt.Sprintf("failed to SenseKernelVariables: %v", err), http.StatusInternalServerError)
	} else {
		rw.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(rw).Encode(respContent); err != nil {
			zap.L().Error("In LinuxKernelVariablesHandler handler failed to write", zap.Error(err))
		}
	}
}

func openedPortsHandler(rw http.ResponseWriter, r *http.Request) {
	respContent, err := sensor.SenseOpenPorts()
	if err != nil {
		http.Error(rw, fmt.Sprintf("failed to SenseOpenPorts: %v", err), http.StatusInternalServerError)
	} else {
		rw.WriteHeader(http.StatusOK)
		if _, err := rw.Write(respContent); err != nil {
			zap.L().Error("In openedPortsHandler handler failed to write", zap.Error(err))
		}
	}
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
	fileContent, err := sensor.SenseLinuxSecurityHardening()
	if err != nil {
		http.Error(rw, fmt.Sprintf("failed to sense linuxSecurityHardeningHandler: %v", err), http.StatusInternalServerError)
	} else {
		rw.WriteHeader(http.StatusOK)
		if _, err := rw.Write(fileContent); err != nil {
			zap.L().Error("In linuxSecurityHardeningHandler handler failed to write", zap.Error(err))
		}
	}
}
