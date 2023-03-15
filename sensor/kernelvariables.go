package sensor

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/kubescape/go-logger"
	"github.com/kubescape/go-logger/helpers"
)

const (
	procSysKernelDir = "/proc/sys/kernel"
	//TODO: add dir for macos (?)
	//TODO: add dir for windows (?)
)

type KernelVariable struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Source string `json:"source"`
}

func SenseProcSysKernel(ctx context.Context) ([]KernelVariable, error) {

	// open system kernel directory (only Linux OS)
	procDir, err := os.Open(procSysKernelDir)

	if err != nil {
		return nil, fmt.Errorf("failed to procSysKernelDir dir(%s): %v", procSysKernelDir, err)
	}
	defer procDir.Close()

	return walkVarsDir(ctx, procSysKernelDir, procDir)
}

func walkVarsDir(ctx context.Context, dirPath string, procDir *os.File) ([]KernelVariable, error) {
	var varsNames []string
	varsList := make([]KernelVariable, 0, 128)

	var err error
	for varsNames, err = procDir.Readdirnames(100); err == nil; varsNames, err = procDir.Readdirnames(100) {
		for varIdx := range varsNames {
			varFileName := path.Join(dirPath, varsNames[varIdx])
			varFile, err := os.Open(varFileName)
			if err != nil {
				if strings.Contains(err.Error(), "permission denied") {
					logger.L().Ctx(ctx).Error("In walkVarsDir failed to open file", helpers.String("varFileName", varFileName),
						helpers.Error(err))
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
				innerVars, err := walkVarsDir(ctx, varFileName, varFile)
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
						logger.L().Ctx(ctx).Error("In walkVarsDir failed to Copy file", helpers.String("varFileName", varFileName),
							helpers.Error(err))
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

func SenseKernelVariables(ctx context.Context) ([]KernelVariable, error) {
	vars, err := SenseProcSysKernel(ctx)
	if confVars, err := SenseKernelConfs(); err != nil {
		logger.L().Ctx(ctx).Error("In SenseKernelVariables failed to SenseKernelConfs", helpers.Error(err))
	} else {
		vars = append(vars, confVars...)
	}
	return vars, err
}
