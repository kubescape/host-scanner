package sensor

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	kubeConfigArgName = "--kubeconfig"
	maxRecursionDepth = 10
)

var (
	ErrNotUnixFS = errors.New("operation not supported by the file system")
)

func ReadFileOnHostFileSystem(fileName string) ([]byte, error) {
	return os.ReadFile(hostPath(fileName))
}

func hostPath(filePath string) string {
	return path.Join(hostFileSystemDefaultLocation, filePath)
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

// MakeContaineredFileInfo is a wrapper of `MakeChangedRootFileInfo` for container files
func makeContaineredFileInfo(filePath string, readContent bool, p *ProcessDetails) (*FileInfo, error) {
	return makeChangedRootFileInfo(filePath, readContent, p.RootDir())
}

// MakeHostFileInfo is a wrapper of `MakeChangedRootFileInfo` for host files
func makeHostFileInfo(filePath string, readContent bool) (*FileInfo, error) {
	return makeChangedRootFileInfo(filePath, readContent, hostFileSystemDefaultLocation)
}

// MakeHostFileInfo is a wrapper of `MakeFileInfo` for rootDir/filePath
func makeChangedRootFileInfo(filePath string, readContent bool, rootDir string) (*FileInfo, error) {
	fullPath := path.Join(rootDir, filePath)
	obj, err := MakeFileInfo(fullPath, readContent)

	if err != nil {
		return obj, err
	}

	obj.Path = filePath

	// Username
	username, err := getUserName(obj.Ownership.UID, rootDir)
	obj.Ownership.Username = username

	if err != nil {
		zap.L().Error("MakeHostFileInfo", zap.Error(err))
	}

	// Groupname
	groupname, err := getGroupName(obj.Ownership.GID, rootDir)
	obj.Ownership.Groupname = groupname

	if err != nil {
		zap.L().Error("MakeHostFileInfo", zap.Error(err))
	}

	return obj, nil
}

// makeHostFileInfoVerbose is wrapper of `MakeHostFileInfo` with error logging
func makeHostFileInfoVerbose(path string, readContent bool, failMsgs ...zap.Field) *FileInfo {
	fileInfo, err := makeHostFileInfo(path, readContent)
	if err != nil {
		logArgs := append([]zapcore.Field{
			zap.String("path", path),
			zap.Error(err),
		},
			failMsgs...,
		)
		zap.L().Error("failed to MakeHostFileInfo", logArgs...)
	}
	return fileInfo
}

// makeHostDirFilesInfo iterate over a directory and make a list of
// file infos for all the files inside it. If `recursive` is set to true,
// the file infos will be added recursively until `maxRecursionDepth` is reached
func makeHostDirFilesInfo(dir string, recursive bool, fileInfos *([]*FileInfo), recursionLevel int) ([]*FileInfo, error) {
	dirInfo, err := os.Open(hostPath(dir))
	if err != nil {
		return nil, fmt.Errorf("failed to open dir at %s: %w", dir, err)
	}
	defer dirInfo.Close()

	if fileInfos == nil {
		fileInfos = &([]*FileInfo{})
	}

	var fileNames []string
	for fileNames, err = dirInfo.Readdirnames(100); err == nil; fileNames, err = dirInfo.Readdirnames(100) {
		for i := range fileNames {
			filePath := path.Join(dir, fileNames[i])
			fileInfo := makeHostFileInfoVerbose(filePath,
				false,
				zap.String("in", "makeHostDirFilesInfo"),
				zap.String("dir", dir),
			)

			if fileInfo != nil {
				*fileInfos = append(*fileInfos, fileInfo)
			}

			if !recursive {
				continue
			}

			// Check if is directory
			stats, err := os.Stat(hostPath(filePath))
			if err != nil {
				zap.L().Error("failed to get file stats",
					zap.String("in", "makeHostDirFilesInfo"),
					zap.String("path", filePath))
				continue
			}
			if stats.IsDir() {
				if recursionLevel+1 == maxRecursionDepth {
					zap.L().Error("max recusrion depth exceeded",
						zap.String("in", "makeHostDirFilesInfo"),
						zap.String("path", filePath))
					continue
				}
				makeHostDirFilesInfo(filePath, recursive, fileInfos, recursionLevel+1)
			}
		}
	}

	if errors.Is(err, io.EOF) {
		err = nil
	}

	return *fileInfos, err
}
