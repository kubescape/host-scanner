package sensor

import (
	"fmt"
	"os"
	"path"
	"strings"

	"go.uber.org/zap"
)

const (
	etcDirName          = "/etc"
	osReleaseFileSuffix = "os-release"
)

func SenseOsRelease() ([]byte, error) {
	osFileName, err := getOsReleaseFile()
	if err == nil {
		return ReadFileOnHostFileSystem(path.Join(etcDirName, osFileName))
	}
	return []byte{}, fmt.Errorf("failed to find os-release file: %v", err)
}

func getOsReleaseFile() (string, error) {
	hostEtcDir := path.Join(HostFileSystemDefaultLocation, etcDirName)
	etcDir, err := os.Open(hostEtcDir)
	if err != nil {
		return "", fmt.Errorf("failed to open etc dir: %v", err)
	}
	etcSons := make([]string, 0)
	for etcSons, err = etcDir.Readdirnames(100); err == nil; etcSons, err = etcDir.Readdirnames(100) {
		for idx := range etcSons {
			if strings.HasSuffix(etcSons[idx], osReleaseFileSuffix) {
				zap.L().Debug("os release file found", zap.String("filename", etcSons[idx]))
				return etcSons[idx], nil
			}
		}
	}
	return "", err
}

func SenseKernelVersion() ([]byte, error) {
	return ReadFileOnHostFileSystem(path.Join(procDirName, "version"))
}
