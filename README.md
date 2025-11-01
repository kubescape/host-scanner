# Kubescape Host-Scanner

[![Test Suite](https://github.com/kubescape/host-scanner/actions/workflows/test-suite.yaml/badge.svg)](https://github.com/kubescape/host-scanner/actions/workflows/test-suite.yaml) ![build](https://img.shields.io/github/actions/workflow/status/kubescape/host-scanner/build.yaml)

![code size](https://img.shields.io/github/languages/code-size/kubescape/host-scanner)

## Description
This component is a data acquisition component in the Kubescape project. Its goal is to collect information about the Kubernetes node host for further security posture evaluation in Kubescape.

## Deployment
Host-scanner is deployed as a privileged Kubernetes DaemonSet in the cluster. It publishes an API for clients to read host information.

## Supported APIs

|  endpoint  |  test-command |  description  | example |
|---|---|---|---|
| `/healthz` | `kubectl curl "http://<host-scanner-pod-name>:7888/healthz" -n <NAMESPACE>` | Returns liveness status of `host-scanner`. | [example] `{"alive": true}` |
| `/readyz` | `kubectl curl "http://<host-scanner-pod-name>:7888/readyz" -n <NAMESPACE>` | Returns readiness status of `host-scanner`. Return `503` in case `host-scanner` is not ready yet. | [example] `{"ready": true}` |
| `/controlplaneinfo` | `kubectl curl "http://<host-scanner-pod-name>:7888/controlplaneinfo" -n <NAMESPACE>` | Returns ControlPlane related information. | [example](docs/controlplaneinfo.json) |
| `/cniinfo` | `kubectl curl "http://<host-scanner-pod-name>:7888/cniinfo" -n <NAMESPACE>` | Returns container network interface information. | [example](docs/cniinfo.json) |
| `/kernelversion` | `kubectl curl "http://<host-scanner-pod-name>:7888/kernelversion" -n <NAMESPACE>` | Returns the kernel version. | [example](docs/kernelversion) |
| `/kubeletinfo` | `kubectl curl "http://<host-scanner-pod-name>:7888/kubeletinfo" -n <NAMESPACE>` | Returns **kubelet** information. | [example](docs/kubeletinfo.json) |
| `/kubeproxyinfo` | `kubectl curl "http://<host-scanner-pod-name>:7888/kubeproxyinfo" -n <NAMESPACE>` | Returns **kube-proxy** command line information. | [example](docs/kubeproxyinfo.json) |
| `/cloudproviderinfo` | `kubectl curl "http://<host-scanner-pod-name>:7888/cloudproviderinfo" -n <NAMESPACE>` | Returns cloud provider information metadata. | [example](docs/cloudprovider.json) |
| `/osrelease` | `kubectl curl "http://<host-scanner-pod-name>:7888/osrelease" -n <NAMESPACE>` | Returns information on the node's operating system. | [example](docs/osrelease) |
| `/openedports` | `kubectl curl "http://<host-scanner-pod-name>:7888/openedports" -n <NAMESPACE>` | Returns information on open ports. | [example](docs/openedports.json) |
| `/linuxsecurityhardening` | `kubectl curl "http://<host-scanner-pod-name>:7888/linuxsecurityhardening" -n <NAMESPACE>` | Returns information about security hardening feature. | [example](docs/linuxsecurityhardening.json) |
| `/version` | `kubectl curl "http://<host-scanner-pod-name>:7888/version" -n <NAMESPACE>` | Returns the build version of the `host-scanner`. | --- |

## Local usage - Setup, Build and Test

### 1. Prerequisites

* Clone the repository.
* [install kubectl](https://kubernetes.io/docs/tasks/tools/)
* [install kubectlcurl plugin](https://github.com/segmentio/kubectl-curl)
* Run K8s Cluster (select 1 of the following options): 
	* for minikube: [install minikube](https://minikube.sigs.k8s.io/docs/start/)
	* for cloud providers:
  		* Access to a remote private repository such as dockerhub.
  		* Access to a cloud provider running cluster.

### 2. Run Host-Scanner on local environment

#### Using Armo's built in script (with minikube)
>To build it:
>```
> python3 ./scripts/build-host-scanner-local.py --build
>```
>To revert it:
>```
> python3 ./scripts/build-host-scanner-local.py --revert
>```
>For Help:
>```
> python3 ./scripts/build-host-scanner-local.py --help
>```

#### Using private cloud providers
>
>***Setup***
> 
>
>Connect to your cluster (different cloud providers have different methods to connect the cluster). should be done once for each new cluster.
>
>Make sure your cluster is kubectl current context
>```
>kubectl config current-context
>```
>
>Create docker-registry secret. Namespace is defined in the deployment yaml. To find the full path for the docker config file run: ```ls ~/.docker/config.json```
>
>```
>>kubectl create secret docker-registry <MySecretName> --from-file=.dockerconfigjson=</path/to/.docker/config.json> -n <namespace>
>```
>
>***Login, Build and Push image to remote image repository***
>
><ins>Docker hub</ins>
>	
>Login to dockerhub
>```
>docker login
>```
>Create a new repository in dockerhub and mark it as private
>
>"myRepoName" should be the exact name of the image repository name in dockerhub
>```
>docker build -f build/Dockerfile . -t <myRepoName>:<MyImageTag>
>```
>
>Push image to a remote repository (dockerhub in our case)
>```
>docker push <myRepoName>:<MyImageTag>
>
>```
><ins>AWS ECR</ins>
>
> Follow [these instructions](https://docs.aws.amazon.com/AmazonECR/latest/userguide/docker-push-ecr-image.html).
>
>
>
>***Configure deployment.yaml***
>
>Configure [k8s-deployment.yaml](deployment/k8s-deployment.yaml) with the pushed image and and imagePullSecrets. Example:
>```yaml
>apiVersion: v1
>kind: Namespace
>metadata:
>  labels:
>    app: host-sensor
>    kubernetes.io/metadata.name: armo-kube-host-sensor
>    tier: armo-kube-host-sensor-control-plane
>  name: armo-kube-host-sensor
>
>---
>apiVersion: apps/v1
>kind: DaemonSet
>metadata:
>  name: host-sensor
>  namespace: armo-kube-host-sensor
>  labels:
>    k8s-app: armo-kube-host-sensor
>spec:
>  selector:
>    matchLabels:
>      name: host-sensor
>  template:
>    metadata:
>      labels:
>        name: host-sensor
>    spec:
>      tolerations:
>      # this toleration is to have the daemonset runnable on master nodes
>      # remove it if your masters can't run pods
>      - key: node-role.kubernetes.io/control-plane
>        operator: Exists
>        effect: NoSchedule
>      - key: node-role.kubernetes.io/master
>        operator: Exists
>        effect: NoSchedule
>      containers:
>      - name: host-sensor
>        image: myRepoName:MyImageTag
>        securityContext:
>          privileged: true
>          readOnlyRootFilesystem: true
>        ports:
>          - name: http
>            hostPort: 7888
>            containerPort: 7888
>        resources:
>          limits:
>            cpu: 1m
>            memory: 200Mi
>          requests:
>            cpu: 1m
>            memory: 200Mi
>        volumeMounts:
>        - mountPath: /host_fs
>          name: host-filesystem
>      imagePullSecrets:
>      - name: MySecretName
>      terminationGracePeriodSeconds: 120
>      dnsPolicy: ClusterFirstWithHostNet
>      automountServiceAccountToken: false
>      volumes:
>      - hostPath:
>          path: /
>          type: Directory
>        name: host-filesystem
>      hostNetwork: true
>      hostPID: true
>      hostIPC: true
>```
### 3. Apply and Test


Create host-scanner pod
```
kubectl apply -f https://raw.githubusercontent.com/kubescape/kubescape/master/core/pkg/hostsensorutils/hostsensor.yaml
```

If command failed, use the below instead:
```
kubectl apply -f deployment/k8s-deployment.yaml
```

Verify pod is launched successfully
```
kubectl get pods -A
```

Test an API
```
kubectl curl "http://<host-scanner-pod-name>:7888/<api-endpoint>" --namespace <namespace>
```

View pod logs

On a new terminal, view pod logs:
```
kubectl logs <host-scanner-pod-name> --namespace <namespace> -f
```

## Contributions

Thanks to all our contributors! Check out our [CONTRIBUTING](https://github.com/kubescape/kubescape/blob/master/CONTRIBUTING.md) file to learn how to join them.

* Feel free to pick a task from the [issues](https://github.com/kubescape/host-scanner/issues?q=is%3Aissue+is%3Aopen+label%3A%22open+for+contribution%22), roadmap or suggest a feature of your own.
* [Open an issue](https://github.com/kubescape/host-scanner/issues/new/choose): we aim to respond to all issues within 48 hours.
* [Join the CNCF Slack](https://slack.cncf.io/) and then our [users](https://cloud-native.slack.com/archives/C04EY3ZF9GE) or [developers](https://cloud-native.slack.com/archives/C04GY6H082K) channel.
