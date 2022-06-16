package sensor

import (
	"errors"
	"os"
	"path"
	"syscall"

	"go.uber.org/zap"
)

const (
	kubeConfigArgName = "--kubeconfig"
)

var (
	ErrNotUnixFS = errors.New("unix operations are not supported")
)

func ReadFileOnHostFileSystem(fileName string) ([]byte, error) {
	return os.ReadFile(hostPath(fileName))
}

func hostPath(filePath string) string {
	return path.Join(HostFileSystemDefaultLocation, filePath)
}

// GetFilePermissions returns file permissions as int.
// On filesystem error, it returns the error as is.
func GetFilePermissions(filePath string) (int, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}

	permInt := int(info.Mode().Perm())

	return permInt, nil
}

// GetFileUNIXOwnership returns the user id and group of a file.
// On error, it return values of -1 for the ids.
// On filesystem error, it returns the error as is.
// If the filesystem not support UNIX ownership (like FAT), it returns ErrNotUnixFS.
func GetFileUNIXOwnership(filePath string) (int64, int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return -1, -1, err
	}

	asUnix, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return -1, -1, ErrNotUnixFS
	}

	user := int64(asUnix.Uid)
	group := int64(asUnix.Gid)

	return user, group, nil
}

// IsPathExists returns true if a given path exist and false otherwise
func IsPathExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// MakeFileInfo returns a `FileInfo` object for given path
// If `readContent` is set to `true`, it adds the file content
// On access error, it returns the error as is
func MakeFileInfo(filePath string, readContent bool) (*FileInfo, error) {
	ret := FileInfo{Path: filePath}

	zap.L().Debug("making file info", zap.String("path", filePath))

	// Permissions
	perms, err := GetFilePermissions(filePath)
	if err != nil {
		return nil, err
	}
	ret.Permissions = perms

	// Ownership
	uid, gid, err := GetFileUNIXOwnership(filePath)
	ret.Ownership = &FileOwnership{UID: uid, GID: gid}
	if err != nil {
		ret.Ownership.Err = err.Error()
	}

	// Content
	if readContent {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		ret.Content = content
	}

	return &ret, nil
}

// MakeHostFileInfo is a wrapper of `MakeFileInfo` for host files
func MakeHostFileInfo(filePath string, readContent bool) (*FileInfo, error) {
	obj, err := MakeFileInfo(hostPath(filePath), readContent)
	if err == nil {
		obj.Path = filePath
	}
	return obj, err
}
