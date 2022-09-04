package sensor

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"go.uber.org/zap"
)

const (
	procSysKernelDir = "/proc/sys/kernel"
)

type KernelVariable struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Source string `json:"source"`
}

func SenseProcSysKernel() ([]KernelVariable, error) {
	procDir, err := os.Open(procSysKernelDir)
	if err != nil {
		return nil, fmt.Errorf("failed to procSysKernelDir dir(%s): %v", procSysKernelDir, err)
	}
	defer procDir.Close()

	return walkVarsDir(procSysKernelDir, procDir)
}

func walkVarsDir(dirPath string, procDir *os.File) ([]KernelVariable, error) {
	var varsNames []string
	varsList := make([]KernelVariable, 0, 128)

	var err error
	for varsNames, err = procDir.Readdirnames(100); err == nil; varsNames, err = procDir.Readdirnames(100) {
		for varIdx := range varsNames {
			varFileName := path.Join(dirPath, varsNames[varIdx])
			varFile, err := os.Open(varFileName)
			if err != nil {
				if strings.Contains(err.Error(), "permission denied") {
					zap.L().Error("In walkVarsDir failed to open file", zap.String("varFileName", varFileName),
						zap.Error(err))
					continue
				}
				return nil, fmt.Errorf("failed to open file (%s): %v", varFileName, err)
			}
			defer varFile.Close()
			fileInfo, err := varFile.Stat()
			if err != nil {
				return nil, fmt.Errorf("failed to stat file (%s): %v", varFileName, err)
			}
			if fileInfo.IsDir() {
				// CAUTION: recursive call!!!
				innerVars, err := walkVarsDir(varFileName, varFile)
				if err != nil {
					return nil, fmt.Errorf("failed to walkVarsDir file (%s): %v", varFileName, err)
				}
				if len(innerVars) > 0 {
					varsList = append(varsList, innerVars...)
				}
			} else if fileInfo.Mode().IsRegular() {
				strBld := strings.Builder{}
				if _, err := io.Copy(&strBld, varFile); err != nil {
					if strings.Contains(err.Error(), "operation not permitted") {
						zap.L().Error("In walkVarsDir failed to Copy file", zap.String("varFileName", varFileName),
							zap.Error(err))
						continue
					}
					return nil, fmt.Errorf("failed to copy file (%s): %v", varFileName, err)
				}
				varsList = append(varsList, KernelVariable{
					Key:    varsNames[varIdx],
					Value:  strBld.String(),
					Source: varFileName,
				})
			}
		}
	}
	if err != io.EOF {
		return nil, fmt.Errorf("failed to Readdirnames of procSysKernelDir dir(%s): %v; found so far: %v", procSysKernelDir, err, varsList)
	}
	return varsList, nil
}

func SenseKernelConfs() ([]KernelVariable, error) {
	varsList := make([]KernelVariable, 0, 16)

	return varsList, nil
}

func SenseKernelVariables() ([]KernelVariable, error) {
	vars, err := SenseProcSysKernel()
	if confVars, err := SenseKernelConfs(); err != nil {
		zap.L().Error("In SenseKernelVariables failed to SenseKernelConfs", zap.Error(err))
	} else {
		vars = append(vars, confVars...)
	}
	return vars, err
}
