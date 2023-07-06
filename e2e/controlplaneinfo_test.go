//go:build kind

package e2e_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	sensor "github.com/kubescape/host-scanner/sensor"
	"github.com/kubescape/go-logger"
)

var _ = Describe("ControlPlaneInfo", func() {
	var (
		res     *http.Response
		err     error
		resBody []byte
	)

	Context("testing /controlplaneinfo endpoint", func() {
		It("should respond to a GET request", func() {
			requestURL := url + "/controlplaneinfo"
			res, err = http.Get(requestURL)
			Expect(err).ToNot(HaveOccurred())
		})
		It("should return a 200 status code", func() {
			Expect(res.StatusCode).To(BeEquivalentTo(200))
		})
		It("should return the expected value of PKIDir and PKIFiles", func() {
			resultBody := &sensor.ControlPlaneInfo{}

			resBody, err = ioutil.ReadAll(res.Body)
			Expect(err).ToNot(HaveOccurred())

			err = json.Unmarshal(resBody, resultBody)
			Expect(err).ToNot(HaveOccurred())

			Expect(resultBody.PKIDIr.Path).To(Equal(pkiFiles.PKIDIr.Path))

			// (leave it there for debugging)
			for i := range resultBody.PKIFiles {
				logger.L().Info(resultBody.PKIFiles[i].Path)
			}
			for i := range resultBody.PKIFiles {
				Expect(resultBody.PKIFiles[i].Path).To(Equal(pkiFiles.PKIFiles[i].Path))
			}
		})
		It("should return the expected value of ApiServerInfo", func() {
			resultBody := &sensor.ControlPlaneInfo{}

			err = json.Unmarshal(resBody, resultBody)
			Expect(err).ToNot(HaveOccurred())

			Expect(resultBody.APIServerInfo.K8sProcessInfo.SpecsFile).
				To(Equal(apiServerInfo.APIServerInfo.K8sProcessInfo.SpecsFile))
		})
		It("should return the expected value of ControllerManagerInfo", func() {
			resultBody := &sensor.ControlPlaneInfo{}

			err = json.Unmarshal(resBody, resultBody)
			Expect(err).ToNot(HaveOccurred())

			Expect(resultBody.ControllerManagerInfo.SpecsFile).
				To(Equal(controllerManagerInfo.ControllerManagerInfo.SpecsFile))
			Expect(resultBody.ControllerManagerInfo.ConfigFile).
				To(Equal(controllerManagerInfo.ControllerManagerInfo.ConfigFile))
			Expect(resultBody.ControllerManagerInfo.KubeConfigFile).
				To(Equal(controllerManagerInfo.ControllerManagerInfo.KubeConfigFile))
			Expect(resultBody.ControllerManagerInfo.ClientCAFile).
				To(Equal(controllerManagerInfo.ControllerManagerInfo.ClientCAFile))
		})
		It("should return the expected value of SchedulerInfo", func() {
			resultBody := &sensor.ControlPlaneInfo{}

			err = json.Unmarshal(resBody, resultBody)
			Expect(err).ToNot(HaveOccurred())

			Expect(resultBody.SchedulerInfo.SpecsFile).
				To(Equal(schedulerInfo.SchedulerInfo.SpecsFile))
			Expect(resultBody.SchedulerInfo.ConfigFile).
				To(Equal(schedulerInfo.SchedulerInfo.ConfigFile))
			Expect(resultBody.SchedulerInfo.KubeConfigFile).
				To(Equal(schedulerInfo.SchedulerInfo.KubeConfigFile))
			Expect(resultBody.SchedulerInfo.ClientCAFile).
				To(Equal(schedulerInfo.SchedulerInfo.ClientCAFile))
		})
		It("should return the expected value of EtcdConfigFile and EtcdDataDir", func() {
			resultBody := &sensor.ControlPlaneInfo{}

			err = json.Unmarshal(resBody, resultBody)
			Expect(err).ToNot(HaveOccurred())

			Expect(resultBody.EtcdConfigFile).
				To(Equal(etcdConfigFile.EtcdConfigFile))
			Expect(resultBody.EtcdDataDir).
				To(Equal(etcdConfigFile.EtcdDataDir))
		})
		It("should return the expected value of AdminConfigFile", func() {
			resultBody := &sensor.ControlPlaneInfo{}

			err = json.Unmarshal(resBody, resultBody)
			Expect(err).ToNot(HaveOccurred())

			Expect(resultBody.AdminConfigFile).
				To(Equal(adminConfigFile.AdminConfigFile))
		})
	})
})
