package sensor

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	ds "github.com/armosec/host-sensor/sensor/datastructures"
	"github.com/armosec/host-sensor/sensor/internal/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	maxRecursionDepth = 10
)

// makeHostFileInfoVerbose makes a file info object
// for the given path on the host file system, and with error logging.
// It returns nil on error.
func makeHostFileInfoVerbose(path string, readContent bool, failMsgs ...zap.Field) *ds.FileInfo {
	return makeChangedRootFileInfoVerbose(utils.HostFileSystemDefaultLocation, path, readContent, failMsgs...)
}

// makeContaineredFileInfoVerbose makes a file info object
// for a given process file system view, and with error logging.
// It returns nil on error.
func makeContaineredFileInfoVerbose(p *utils.ProcessDetails, filePath string, readContent bool, failMsgs ...zap.Field) *ds.FileInfo {
	return makeChangedRootFileInfoVerbose(p.RootDir(), filePath, readContent, failMsgs...)
}

// makeChangedRootFileInfoVerbose makes a file info object
// for the given path on the given root directory, and with error logging.
func makeChangedRootFileInfoVerbose(rootDir string, path string, readContent bool, failMsgs ...zap.Field) *ds.FileInfo {
	fileInfo, err := utils.MakeChangedRootFileInfo(rootDir, path, readContent)
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
func makeHostDirFilesInfoVerbose(dir string, recursive bool, fileInfos *([]*ds.FileInfo), recursionLevel int) ([]*ds.FileInfo, error) {
	dirInfo, err := os.Open(utils.HostPath(dir))
	if err != nil {
		return nil, fmt.Errorf("failed to open dir at %s: %w", dir, err)
	}
	defer dirInfo.Close()

	if fileInfos == nil {
		fileInfos = &([]*ds.FileInfo{})
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
			stats, err := os.Stat(utils.HostPath(filePath))
			if err != nil {
				zap.L().Error("failed to get file stats",
					zap.String("in", "makeHostDirFilesInfo"),
					zap.String("path", filePath))
				continue
			}
			if stats.IsDir() {
				if recursionLevel+1 == maxRecursionDepth {
					zap.L().Error("max recursion depth exceeded",
						zap.String("in", "makeHostDirFilesInfo"),
						zap.String("path", filePath))
					continue
				}
				makeHostDirFilesInfoVerbose(filePath, recursive, fileInfos, recursionLevel+1)
			}
		}
	}

	if errors.Is(err, io.EOF) {
		err = nil
	}

	return *fileInfos, err
}
