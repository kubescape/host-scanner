//go:build kind

package e2e_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/kubescape/host-scanner/sensor"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/procspy"
)

var _ = Describe("Openedports", func() {
	var (
		res     *http.Response
		err     error
		resBody []byte
	)

	Context("testing /openedports endpoint", func() {
		It("should respond to a GET request", func() {
			requestURL := url + "/openedports"
			res, err = http.Get(requestURL)
			Expect(err).ToNot(HaveOccurred())
		})
		It("should return a 200 status code", func() {
			Expect(res.StatusCode).To(BeEquivalentTo(200))
		})
		It("should return the expected value of OpenedPortsStatus", func() {
			jsonToCompare := &sensor.OpenPortsStatus{
				TcpPorts: []procspy.Connection{
					{
						LocalPort: 7888,
					},
				},
				UdpPorts:  []procspy.Connection{},
				ICMPPorts: []procspy.Connection{},
			}
			jsonOpenedPortsInfo := &sensor.OpenPortsStatus{}

			resBody, err = ioutil.ReadAll(res.Body)
			Expect(err).ToNot(HaveOccurred())

			err = json.Unmarshal(resBody, jsonOpenedPortsInfo)
			Expect(err).ToNot(HaveOccurred())

			for i := range jsonOpenedPortsInfo.TcpPorts {
				Expect(jsonOpenedPortsInfo.TcpPorts[i].LocalPort).
					To(Equal(jsonToCompare.TcpPorts[i].LocalPort))
			}
			for i := range jsonOpenedPortsInfo.UdpPorts {
				Expect(jsonOpenedPortsInfo.UdpPorts[i].LocalPort).
					To(Equal(jsonToCompare.UdpPorts[i].LocalPort))
			}
			for i := range jsonOpenedPortsInfo.ICMPPorts {
				Expect(jsonOpenedPortsInfo.ICMPPorts[i].LocalPort).
					To(Equal(jsonToCompare.ICMPPorts[i].LocalPort))
			}
		})
	})
})
