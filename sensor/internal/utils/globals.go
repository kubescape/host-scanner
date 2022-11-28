package utils

import (
	"net/http"
	"sync"
)

var (
	// Where the host sensor is expecting host fs to be mounted.
	// Defined as var for testing purposes only
	HostFileSystemDefaultLocation = "/host_fs"

	// global http.client instance to reduce object resource overuse.
	httpClient *http.Client

	lock = &sync.Mutex{}
)

// GetHttpClient - instantiate http.client object
func GetHttpClient() *http.Client {
	if httpClient == nil {
		lock.Lock()
		defer lock.Unlock()
		if httpClient == nil {
			httpClient = &http.Client{}
		}
	}
	return httpClient
}
