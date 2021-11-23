package sensor

import (
	"os"
	"path"
)

func ReadFileOnHostFileSystem(fileName string) ([]byte, error) {
	return os.ReadFile(path.Join(HostFileSystemDefaultLocation, fileName))
}
