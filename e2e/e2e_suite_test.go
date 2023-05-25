package e2e_test

import (
	"flag"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	cfg             *rest.Config
	k8sClient       *kubernetes.Clientset
	testEnv         *envtest.Environment
	deployment      *appsv1.Deployment
	namespace       string  = "kubescape-host-scanner"
	localServerPort *string = flag.String("port", "7888", "destination port")
	url             string
)

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "host-scanner e2e-tests")
}

var _ = BeforeSuite(func(done Done) {
	defer close(done)

	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter)))

	By("bootstrapping test environment")
	useCluster := true
	testEnv = &envtest.Environment{
		UseExistingCluster: &useCluster,
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	k8sClient, err = kubernetes.NewForConfig(cfg)
	Expect(k8sClient).ToNot(BeNil())

	//podName, err := GetPodName(k8sClient, namespace, "name=host-scanner")
	Expect(err).ToNot(HaveOccurred())

	url = fmt.Sprintf("http://localhost:%s", *localServerPort)
	//go func() {
	//	err = CreatePortForward(cfg, k8sClient, namespace, podName, []string{localServerPort, "7888"})
	//	Expect(err).ToNot(HaveOccurred())
	//	time.Sleep(120)
	//}()
	//err = CreatePortForward(cfg, k8sClient, namespace, podName, []string{localServerPort, "7888"})
	//Expect(err).ToNot(HaveOccurred())
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	Expect(testEnv.Stop()).ToNot(HaveOccurred())
})
