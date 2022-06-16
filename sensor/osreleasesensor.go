package sensor

import (
	"fmt"
	"os"
	"path"
	"strings"

	"go.uber.org/zap"
)

const (
	etcDirName               = "/etc"
	osReleaseFileSuffix      = "os-release"
	appArmorProfilesFileName = "/sys/kernel/security/apparmor/profiles"
	seLinuxConfigFileName    = "/etc/selinux/semanage.conf"
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
	defer etcDir.Close()
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

func getAppArmorStatus() string {
	statusStr := "unloaded"
	hostAppArmorProfilesFileName := path.Join(HostFileSystemDefaultLocation, appArmorProfilesFileName)
	profFile, err := os.Open(hostAppArmorProfilesFileName)
	if err == nil {
		defer profFile.Close()
		statusStr = "stopped"
		content, err := ReadFileOnHostFileSystem(appArmorProfilesFileName)
		if err == nil && len(content) > 0 {
			statusStr = string(content)
		}
	}
	return statusStr
}

func getSELinuxStatus() string {
	statusStr := "not found"
	hostAppArmorProfilesFileName := path.Join(HostFileSystemDefaultLocation, seLinuxConfigFileName)
	conFile, err := os.Open(hostAppArmorProfilesFileName)
	if err == nil {
		defer conFile.Close()
		content, err := ReadFileOnHostFileSystem(appArmorProfilesFileName)
		if err == nil && len(content) > 0 {
			statusStr = string(content)
		}
	}
	return statusStr
}

func SenseLinuxSecurityHardening() (*LinuxSecurityHardeningStatus, error) {
	res := LinuxSecurityHardeningStatus{}

	res.AppArmor = getAppArmorStatus()
	res.SeLinux = getSELinuxStatus()

	return &res, nil
}
