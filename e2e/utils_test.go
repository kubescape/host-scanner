package e2e_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/transport/spdy"
)

func DeployCurlContainer(k8sClient *kubernetes.Clientset, namespace string) (*appsv1.Deployment, error) {
	// install test container to run curl againsta host-scanner
	labels := map[string]string{"app": "pod-exec"}
	replicas := int32(1)
	app := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "curl",
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:    "curl",
							Image:   "curlimages/curl",
							Command: []string{"sleep", "300"},
						},
					},
				},
			},
		},
	}
	deployment, err := k8sClient.AppsV1().
		Deployments(namespace).
		Create(context.TODO(), app, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return deployment, nil
}

func DeleteCurlContainer(k8sClient *kubernetes.Clientset, namespace string) error {
	err := k8sClient.AppsV1().
		Deployments(namespace).
		Delete(context.Background(), "curl", metav1.DeleteOptions{})
	return err
}

func WaitForPod(namespace string, labelSelector string) (watch.Interface, error) {
	watch, err := k8sClient.CoreV1().
		Pods(namespace).
		Watch(context.TODO(), metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, err
	}
	return watch, nil
}

func ListPod(namespace string, labelSelector string) (string, error) {
	// retrieve podName from curl container
	var pod string
	pods, err := k8sClient.CoreV1().
		Pods(namespace).
		List(context.TODO(), metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return "", err
	}
	for _, v := range pods.Items {
		pod = v.Name
	}
	return pod, nil
}

func GetPodIP(k8sClient *kubernetes.Clientset, namespace string, labelSelector string) (string, error) {
	var containerIP string
	pod, err := k8sClient.CoreV1().
		Pods(namespace).
		List(context.TODO(), metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return "", err
	}
	for _, v := range pod.Items {
		containerIP = v.Status.PodIP
	}
	return containerIP, nil
}

func GetPodName(k8sClient *kubernetes.Clientset, namespace string, labelSelector string) (string, error) {
	var podName string
	pod, err := k8sClient.CoreV1().
		Pods(namespace).
		List(context.TODO(), metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return "", err
	}
	for _, v := range pod.Items {
		podName = v.Name
	}
	return podName, nil
}

func ExecInPod(k8sClient *kubernetes.Clientset, namespace, pod, command string) (*bytes.Buffer, *bytes.Buffer, error) {
	buf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	request := k8sClient.CoreV1().
		RESTClient().
		Post().
		Namespace(namespace).
		Resource("pods").
		Name(pod).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Command: []string{"/bin/sh", "-c", command},
			Stdin:   false,
			Stdout:  true,
			Stderr:  true,
			TTY:     true,
		}, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(cfg, "POST", request.URL())
	if err != nil {
		return buf, errBuf, fmt.Errorf("error in initializing SPDY executor: %s", err)
	}
	err = exec.StreamWithContext(context.TODO(), remotecommand.StreamOptions{
		Stdout: buf,
		Stderr: errBuf,
	})
	if err != nil {
		return buf, errBuf, fmt.Errorf("error in executing command: %s", err)
	}
	return buf, errBuf, nil
}

func CreatePortForward(config *rest.Config, k8sClient *kubernetes.Clientset, namespace, pod string, ports []string) error {
	stopCh := make(<-chan struct{})
	readyCh := make(chan struct{})
	defer close(readyCh)

	req := k8sClient.CoreV1().
		RESTClient().
		Post().
		Resource("pods").
		Namespace(namespace).
		Name(pod).
		SubResource("portforward")

	transport, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return fmt.Errorf("spdy error: %v", err)
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", req.URL())
	fw, err := portforward.New(dialer, ports, stopCh, readyCh, os.Stdout, os.Stderr)
	if err != nil {
		return fmt.Errorf("portforward error: %v", err)
	}
	return fw.ForwardPorts()
}
