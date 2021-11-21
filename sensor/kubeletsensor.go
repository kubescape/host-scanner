package sensor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	"go.uber.org/zap"
	kubeletv1beta1 "k8s.io/kubelet/config/v1beta1"
)

const (
	procDirName          = "/proc"
	kubeletProcessSuffix = "/kubelet"
	kubeletConfigArgName = "--config"
)

type ProcessDetails struct {
	PID     int32    `json:"pid"`
	CmdLine []string `json:"cmdline"`
}

func LocateProcessByExecSuffix(processSuffix string) (*ProcessDetails, error) {
	procDir, err := os.Open(procDirName)
	if err != nil {
		return nil, fmt.Errorf("failed to open processes dir: %v", err)
	}
	pidDirs := make([]string, 0)
	for pidDirs, err = procDir.Readdirnames(100); err == nil; pidDirs, err = procDir.Readdirnames(100) {
		for pidIdx := range pidDirs {
			// since processes are about to die in the middle of the loop, we will ignore next errors
			pid, err := strconv.ParseInt(pidDirs[pidIdx], 10, 0)
			if err != nil {
				continue
			}
			specificProcessCMD := path.Join(procDirName, pidDirs[pidIdx], "cmdline")
			cmdLine, err := os.ReadFile(specificProcessCMD)
			if err != nil {
				continue
			}
			cmdLineSplitted := bytes.Split(cmdLine, []byte{00})
			if bytes.HasSuffix(cmdLineSplitted[0], []byte(processSuffix)) {
				zap.L().Debug("process found", zap.String("processSuffix", processSuffix))
				res := &ProcessDetails{PID: int32(pid), CmdLine: make([]string, 0, len(cmdLineSplitted))}
				for splitIdx := range cmdLineSplitted {
					res.CmdLine = append(res.CmdLine, string(cmdLineSplitted[splitIdx]))
				}
				return res, nil
			}
		}
	}
	if err != io.EOF {
		return nil, fmt.Errorf("failed to read processes dir names: %v", err)
	}
	return nil, fmt.Errorf("no process with given suffix found")
}

func LocateKubeletProcess() (*ProcessDetails, error) {
	return LocateProcessByExecSuffix(kubeletProcessSuffix)
}

func LocateKubeletConfig(kubeletConfArgs string) (*kubeletv1beta1.KubeletConfiguration, error) {
	res := &kubeletv1beta1.KubeletConfiguration{}
	confFile, err := os.OpenFile(kubeletConfArgs, os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open conf file: %v", err)
	}
	if err := json.NewDecoder(confFile).Decode(res); err != nil {
		return nil, fmt.Errorf("failed to decode conf file: %v", err)
	}
	return res, nil
}

func SenseKubeletConfigurations() (*kubeletv1beta1.KubeletConfiguration, error) {
	kubeletProcess, err := LocateKubeletProcess()
	if err != nil {
		return nil, fmt.Errorf("failed to LocateKubeletProcess: %v", err)
	}
	kubeletConfFileLocation := ""
	for argIdx := range kubeletProcess.CmdLine {
		if strings.HasPrefix(kubeletProcess.CmdLine[argIdx], kubeletConfigArgName) {
			kubeletConfFileLocation = kubeletProcess.CmdLine[argIdx][len(kubeletConfigArgName):]
			if strings.HasPrefix(kubeletConfFileLocation, "=") {
				kubeletConfFileLocation = kubeletConfFileLocation[1:]
			} else if argIdx+1 < len(kubeletProcess.CmdLine) {
				kubeletConfFileLocation = kubeletProcess.CmdLine[argIdx+1]
			} else {
				zap.L().Error("In SenseKubeletConfigurations failed to read kubeletConfFileLocation", zap.Any("kubeletProcess", kubeletProcess))
				return nil, fmt.Errorf("no valid config location argument found")
			}
		}
	}
	return LocateKubeletConfig(kubeletConfFileLocation)
}
