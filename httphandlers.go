package main

import (
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
}

func osReleaseHandler(rw http.ResponseWriter, r *http.Request) {
	fileContent, err := sensor.SenseOsRelease()
	if err != nil {
		http.Error(rw, fmt.Sprintf("failed to sense kubelet conf: %v", err), http.StatusInternalServerError)
	} else {
		rw.WriteHeader(http.StatusOK)
		if _, err := rw.Write(fileContent); err != nil {
			zap.L().Error("In kubeletConfigurations handler failed to write", zap.Error(err))
		}
	}
}
