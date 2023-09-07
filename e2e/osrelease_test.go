//go:build kind

package e2e_test

import (
	"io"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OsRelease", func() {
	var (
		res     *http.Response
		err     error
		resBody []byte
		// this is usually the /etc/os-release content of github-actions workflows
		expectedResult = `NAME="Ubuntu"`
	)

	Context("testing /osrelease endpoint", func() {
		It("should respond to a GET request", func() {
			requestURL := url + "/osrelease"
			res, err = http.Get(requestURL)
			Expect(err).ToNot(HaveOccurred())
		})
		It("should return a 200 status code", func() {
			Expect(res.StatusCode).To(BeEquivalentTo(200))
		})
		It("should return the expected value", func() {
			resBody, err = io.ReadAll(res.Body)
			Ω(string(resBody)).Should(ContainSubstring(expectedResult))
		})
	})
})
