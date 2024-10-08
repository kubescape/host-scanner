package e2e_test

import (
	"io"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version", func() {
	var (
		res     *http.Response
		err     error
		resBody []byte
		// beign compiled for test purpose, we set the version to "latest"
		expectedResult = "\"test\"\n"
	)

	Context("testing /version endpoint", func() {
		It("should respond to a GET request", func() {
			requestURL := url + "/version"
			res, err = http.Get(requestURL)
			Expect(err).ToNot(HaveOccurred())
		})
		It("should return a 200 status code", func() {
			Expect(res.StatusCode).To(BeEquivalentTo(200))
		})
		It("should return the expected value", func() {
			resBody, err = io.ReadAll(res.Body)
			Expect(string(resBody)).To(BeEquivalentTo(expectedResult))
		})
	})
})
