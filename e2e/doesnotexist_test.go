package e2e_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Doesnotexist", func() {
	var (
		res *http.Response
		err error
	)

	Context("testing /doesnotexist endpoint", func() {
		It("should respond to a GET request", func() {
			requestURL := url + "/doesnotexist"
			res, err = http.Get(requestURL)
			Expect(err).ToNot(HaveOccurred())
		})
		It("should return a 200 status code", func() {
			Expect(res.StatusCode).To(BeEquivalentTo(404))
		})
	})
})
