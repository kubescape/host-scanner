package sensor

import (
	"github.com/armosec/utils-go/httputils"
	"github.com/kubescape/host-scanner/sensor/internal/utils"
)

// CloudProviderInfo holds information about the Cloud Provider
type CloudProviderInfo struct {
	// Has access to cloud provider meta data API
	ProviderMetaDataAPIAccess bool `json:"providerMetaDataAPIAccess,omitempty"`
}

// MetaDataAPIRequests - hold information on major cloud providers meta data access urls.
var MetaDataAPIRequests = []struct {
	url     string
	headers map[string]string
}{
	{
		"http://169.254.169.254/computeMetadata/v1/?alt=json&recursive=true",
		map[string]string{"Metadata-Flavor": "Google"},
	},
	{
		"http://169.254.169.254/metadata/instance?api-version=2021-02-01",
		map[string]string{"Metadata": "true"},
	},
	{
		"http://169.254.169.254/latest/meta-data/local-hostname",
		map[string]string{},
	},
}

// SenseCloudProviderInfo returns `CloudProviderInfo`
func SenseCloudProviderInfo() (*CloudProviderInfo, error) {

	ret := CloudProviderInfo{}

	ret.ProviderMetaDataAPIAccess = hasMetaDataAPIAccess()

	return &ret, nil
}

// hasMetaDataAPIAccess - checks if there is an access to cloud provider meta data
func hasMetaDataAPIAccess() bool {
	client := utils.GetHttpClient()
	client.Timeout = 1000000000

	for _, req := range MetaDataAPIRequests {
		res, err := httputils.HttpGet(client, req.url, req.headers)

		if err == nil && res.StatusCode == 200 {
			return true
		}
	}

	return false

}
