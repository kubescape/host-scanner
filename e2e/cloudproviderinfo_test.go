package e2e_test

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/kubescape/host-scanner/sensor"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cloudproviderinfo", func() {

	var (
		res     *http.Response
		err     error
		resBody []byte
	)

	Context("testing /cloudproviderinfo endpoint", func() {
		It("should respond to a GET request", func() {
			requestURL := url + "/cloudproviderinfo"
			res, err = http.Get(requestURL)
			Expect(err).ToNot(HaveOccurred())
		})
		It("should return a 200 status code", func() {
			Expect(res.StatusCode).To(BeEquivalentTo(200))
		})
		It("should return the expected value of CloudProviderInfo", func() {
			resultBody := &sensor.CloudProviderInfo{}

			resBody, err = io.ReadAll(res.Body)
			Expect(err).ToNot(HaveOccurred())

			err = json.Unmarshal(resBody, resultBody)
			Expect(err).ToNot(HaveOccurred())

			Expect(resultBody).To(Equal(cloudProviderInfo))
		})
	})
})
