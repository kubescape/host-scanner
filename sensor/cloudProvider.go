package sensor

import (
	"net/http"

	"github.com/armosec/utils-go/httputils"
)

// CloudProviderInfo holds information about the Cloud Provider
type CloudProviderInfo struct {
	// Has access to cloud provider meta data API
	ProviderMetaDataAPIAccess bool `json:"providerMetaDataAPIAccess,omitempty"`
}

// SenseCloudProviderInfo returns `CloudProviderInfo`
func SenseCloudProviderInfo() (*CloudProviderInfo, error) {

	ret := CloudProviderInfo{}

	ret.ProviderMetaDataAPIAccess = hasMetaDataAPIAccess()

	return &ret, nil
}

// hasMetaDataAPIAccess - checks if there is an access to cloud provider meta data
func hasMetaDataAPIAccess() bool {
	client := &http.Client{}

	res, err := httputils.HttpGet(client, "http://169.254.169.254/computeMetadata/v1/?alt=json&recursive=true", map[string]string{"Metadata-Flavor": "Google"})

	if err == nil && res.StatusCode == 200 {
		return true
	}

	res, err = httputils.HttpGet(client, "http://169.254.169.254/metadata/instance?api-version=2021-02-01", map[string]string{"Metadata": "true"})

	if err == nil && res.StatusCode == 200 {
		return true
	}

	res, err = httputils.HttpGet(client, "http://169.254.169.254/latest/meta-data/local-hostname", nil)

	if err == nil && res.StatusCode == 200 {
		return true
	}

	return false

}
