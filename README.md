# Kubescape host-scanner
## Description
This component is a data acquisition component in the Kubescape project. Its goal is to collect information about the Kubernetes node host for further security posture evaluation in Kubescape.

## Deployment
Host-scanner is deployed as a privileged Kubernetes DaemonSet in the cluster. It publishes an API for clients to read host information.

## Supported APIs

**ControlPlaneInfo** - returns ControlPlane related information. Returns 404 if no information exist (=the node is not a control plane). [example](README.md#controlplaneinfo)
```
kubectl curl "http://<PODNAME>:7888/ControlPlaneInfo" -n <NAMESPACE>

```
**CNIInfo** - returns container network interface information. [example](README.md#cniinfo)
```
kubectl curl "http://<PODNAME>:7888/CNIInfo" -n <NAMESPACE>

```

**kernelVersion** - returns the kernel version. [example](README.md#kernelversion)
```
kubectl curl "http://<PODNAME>:7888/kernelVersion" -n <NAMESPACE>

```

**KubeletInfo** - returns kubelet information. [example](README.md#kubeletinfo)
```
kubectl curl "http://<PODNAME>:7888/KubeletInfo" -n <NAMESPACE>

```

**kubeProxyInfo** - returns kube-proxy command line info. [example](README.md#kubeproxyinfo)
```
kubectl curl "http://<PODNAME>:7888/kubeProxyInfo" -n <NAMESPACE>

```

**cloudProviderInfo** - returns information on cloud provider. [example](README.md#cloudproviderinfo)
```
kubectl curl "http://<PODNAME>:7888/cloudProviderInfo" -n <NAMESPACE>

```

**osRelease** - returns information on the node's operating system. [example](README.md#osrelease)
```
kubectl curl "http://<PODNAME>:7888/osRelease" -n <NAMESPACE>

```

**openedPorts** - returns information on open ports. [example](README.md#openedports)
```
kubectl curl "http://<PODNAME>:7888/openedPorts" -n <NAMESPACE>

```

**version** - returns the build version of the host-scanner.
```
kubectl curl "http://<PODNAME>:7888/version" -n <NAMESPACE>

```



## Build & test Host-Scanner on local environment

### 1. Pre-requisites

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
>For build it:
>```
> python3 ./scripts/build-host-scanner-local.py --build
>```
>For revert it:
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
>          procMount: Unmasked
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
kubectl apply -f deployment/k8s-deployment.yaml 
```

Verify pod is launched successfully
```
kubectl get pods -A
```

Test an API
```
kubectl curl "http://<podName>:7888/<APIName>" --namespace <namespace>
```

View pod logs

On a new terminal, view pod logs:
```
kubectl logs <host-scanner-pod-name> --namespace <namespace> -f
```

## APIS Response examples

### ControlPlaneInfo

```json
{
	"APIServerInfo": {
		"specsFile": {
			"ownership": {
				"uid": 0,
				"gid": 0,
				"username": "root",
				"groupname": "root"
			},
			"path": "/etc/kubernetes/manifests/kube-apiserver.yaml",
			"permissions": 384
		},
		"cmdLine": "kube-apiserver --advertise-address=192.168.49.2 --allow-privileged=true --authorization-mode=Node,RBAC --client-ca-file=/var/lib/minikube/certs/ca.crt --enable-admission-plugins=NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota --enable-bootstrap-token-auth=true --etcd-cafile=/var/lib/minikube/certs/etcd/ca.crt --etcd-certfile=/var/lib/minikube/certs/apiserver-etcd-client.crt --etcd-keyfile=/var/lib/minikube/certs/apiserver-etcd-client.key --etcd-servers=https://127.0.0.1:2379 --kubelet-client-certificate=/var/lib/minikube/certs/apiserver-kubelet-client.crt --kubelet-client-key=/var/lib/minikube/certs/apiserver-kubelet-client.key --kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname --proxy-client-cert-file=/var/lib/minikube/certs/front-proxy-client.crt --proxy-client-key-file=/var/lib/minikube/certs/front-proxy-client.key --requestheader-allowed-names=front-proxy-client --requestheader-client-ca-file=/var/lib/minikube/certs/front-proxy-ca.crt --requestheader-extra-headers-prefix=X-Remote-Extra- --requestheader-group-headers=X-Remote-Group --requestheader-username-headers=X-Remote-User --secure-port=8443 --service-account-issuer=https://kubernetes.default.svc.cluster.local --service-account-key-file=/var/lib/minikube/certs/sa.pub --service-account-signing-key-file=/var/lib/minikube/certs/sa.key --service-cluster-ip-range=10.96.0.0/12 --tls-cert-file=/var/lib/minikube/certs/apiserver.crt --tls-private-key-file=/var/lib/minikube/certs/apiserver.key "
	},
	"controllerManagerInfo": {
		"specsFile": {
			"ownership": {
				"uid": 0,
				"gid": 0,
				"username": "root",
				"groupname": "root"
			},
			"path": "/etc/kubernetes/manifests/kube-controller-manager.yaml",
			"permissions": 384
		},
		"configFile": {
			"ownership": {
				"uid": 0,
				"gid": 0,
				"username": "root",
				"groupname": "root"
			},
			"path": "/etc/kubernetes/controller-manager.conf",
			"permissions": 384
		},
		"cmdLine": "kube-controller-manager --allocate-node-cidrs=true --authentication-kubeconfig=/etc/kubernetes/controller-manager.conf --authorization-kubeconfig=/etc/kubernetes/controller-manager.conf --bind-address=127.0.0.1 --client-ca-file=/var/lib/minikube/certs/ca.crt --cluster-cidr=10.244.0.0/16 --cluster-name=mk --cluster-signing-cert-file=/var/lib/minikube/certs/ca.crt --cluster-signing-key-file=/var/lib/minikube/certs/ca.key --controllers=*,bootstrapsigner,tokencleaner --kubeconfig=/etc/kubernetes/controller-manager.conf --leader-elect=false --requestheader-client-ca-file=/var/lib/minikube/certs/front-proxy-ca.crt --root-ca-file=/var/lib/minikube/certs/ca.crt --service-account-private-key-file=/var/lib/minikube/certs/sa.key --service-cluster-ip-range=10.96.0.0/12 --use-service-account-credentials=true "
	},
	"schedulerInfo": {
		"specsFile": {
			"ownership": {
				"uid": 0,
				"gid": 0,
				"username": "root",
				"groupname": "root"
			},
			"path": "/etc/kubernetes/manifests/kube-scheduler.yaml",
			"permissions": 384
		},
		"configFile": {
			"ownership": {
				"uid": 0,
				"gid": 0,
				"username": "root",
				"groupname": "root"
			},
			"path": "/etc/kubernetes/scheduler.conf",
			"permissions": 384
		},
		"cmdLine": "kube-scheduler --authentication-kubeconfig=/etc/kubernetes/scheduler.conf --authorization-kubeconfig=/etc/kubernetes/scheduler.conf --bind-address=127.0.0.1 --kubeconfig=/etc/kubernetes/scheduler.conf --leader-elect=false "
	},
	"etcdConfigFile": {
		"ownership": {
			"uid": 0,
			"gid": 0,
			"username": "root",
			"groupname": "root"
		},
		"path": "/etc/kubernetes/manifests/etcd.yaml",
		"permissions": 384
	},
	"etcdDataDir": {
		"ownership": {
			"uid": 0,
			"gid": 0,
			"username": "root",
			"groupname": "root"
		},
		"path": "/var/lib/minikube/etcd",
		"permissions": 448
	},
	"adminConfigFile": {
		"ownership": {
			"uid": 0,
			"gid": 0,
			"username": "root",
			"groupname": "root"
		},
		"path": "/etc/kubernetes/admin.conf",
		"permissions": 384
	}
}
```


### CNIInfo

```json
{ 
    "CNIConfigFiles": [
            {
                "ownership": {
                    "uid": 0,
                    "gid": 0,
                    "username": "root",
                    "groupname": "root"
                },
                "path": "/etc/cni/net.d/10-minikube.conflist",
                "permissions": 504
            }
        ],
    "CNINames":["Flannel","Calico"]
}
```



### kernelVersion

```
Linux version 5.15.0-56-generic (buildd@lcy02-amd64-004) (gcc (Ubuntu 11.3.0-1ubuntu1~22.04) 11.3.0, GNU ld (GNU Binutils for Ubuntu) 2.38) #62-Ubuntu SMP Tue Nov 22 19:54:14 UTC 2022
```


### KubeletInfo

```json
{
	"serviceFiles": [
		{
			"ownership": {
				"uid": 0,
				"gid": 0,
				"username": "root",
				"groupname": "root"
			},
			"path": "/etc/systemd/system/kubelet.service.d/10-kubeadm.conf",
			"permissions": 420
		}
	],
	"configFile": {
		"ownership": {
			"uid": 0,
			"gid": 0,
			"username": "root",
			"groupname": "root"
		},
		"path": "/var/lib/kubelet/config.yaml",
		"content": "YXBpVmVyc2lvbjoga3ViZWxldC5jb25maWcuazhzLmlvL3YxYmV0YTEKYXV0aGVudGljYXRpb246CiAgYW5vbnltb3VzOgogICAgZW5hYmxlZDogZmFsc2UKICB3ZWJob29rOgogICAgY2FjaGVUVEw6IDBzCiAgICBlbmFibGVkOiB0cnVlCiAgeDUwOToKICAgIGNsaWVudENBRmlsZTogL3Zhci9saWIvbWluaWt1YmUvY2VydHMvY2EuY3J0CmF1dGhvcml6YXRpb246CiAgbW9kZTogV2ViaG9vawogIHdlYmhvb2s6CiAgICBjYWNoZUF1dGhvcml6ZWRUVEw6IDBzCiAgICBjYWNoZVVuYXV0aG9yaXplZFRUTDogMHMKY2dyb3VwRHJpdmVyOiBzeXN0ZW1kCmNsdXN0ZXJETlM6Ci0gMTAuOTYuMC4xMApjbHVzdGVyRG9tYWluOiBjbHVzdGVyLmxvY2FsCmNwdU1hbmFnZXJSZWNvbmNpbGVQZXJpb2Q6IDBzCmV2aWN0aW9uSGFyZDoKICBpbWFnZWZzLmF2YWlsYWJsZTogMCUKICBub2RlZnMuYXZhaWxhYmxlOiAwJQogIG5vZGVmcy5pbm9kZXNGcmVlOiAwJQpldmljdGlvblByZXNzdXJlVHJhbnNpdGlvblBlcmlvZDogMHMKZmFpbFN3YXBPbjogZmFsc2UKZmlsZUNoZWNrRnJlcXVlbmN5OiAwcwpoZWFsdGh6QmluZEFkZHJlc3M6IDEyNy4wLjAuMQpoZWFsdGh6UG9ydDogMTAyNDgKaHR0cENoZWNrRnJlcXVlbmN5OiAwcwppbWFnZUdDSGlnaFRocmVzaG9sZFBlcmNlbnQ6IDEwMAppbWFnZU1pbmltdW1HQ0FnZTogMHMKa2luZDogS3ViZWxldENvbmZpZ3VyYXRpb24KbG9nZ2luZzoKICBmbHVzaEZyZXF1ZW5jeTogMAogIG9wdGlvbnM6CiAgICBqc29uOgogICAgICBpbmZvQnVmZmVyU2l6ZTogIjAiCiAgdmVyYm9zaXR5OiAwCm1lbW9yeVN3YXA6IHt9Cm5vZGVTdGF0dXNSZXBvcnRGcmVxdWVuY3k6IDBzCm5vZGVTdGF0dXNVcGRhdGVGcmVxdWVuY3k6IDBzCnJvdGF0ZUNlcnRpZmljYXRlczogdHJ1ZQpydW50aW1lUmVxdWVzdFRpbWVvdXQ6IDBzCnNodXRkb3duR3JhY2VQZXJpb2Q6IDBzCnNodXRkb3duR3JhY2VQZXJpb2RDcml0aWNhbFBvZHM6IDBzCnN0YXRpY1BvZFBhdGg6IC9ldGMva3ViZXJuZXRlcy9tYW5pZmVzdHMKc3RyZWFtaW5nQ29ubmVjdGlvbklkbGVUaW1lb3V0OiAwcwpzeW5jRnJlcXVlbmN5OiAwcwp2b2x1bWVTdGF0c0FnZ1BlcmlvZDogMHMK",
		"permissions": 420
	},
	"kubeConfigFile": {
		"ownership": {
			"uid": 0,
			"gid": 0,
			"username": "root",
			"groupname": "root"
		},
		"path": "/etc/kubernetes/kubelet.conf",
		"permissions": 384
	},
	"clientCAFile": {
		"ownership": {
			"uid": 0,
			"gid": 0,
			"username": "root",
			"groupname": "root"
		},
		"path": "/var/lib/minikube/certs/ca.crt",
		"permissions": 420
	},
	"cmdLine": "/var/lib/minikube/binaries/v1.25.3/kubelet --bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --config=/var/lib/kubelet/config.yaml --container-runtime=remote --container-runtime-endpoint=/var/run/cri-dockerd.sock --hostname-override=minikube --image-service-endpoint=/var/run/cri-dockerd.sock --kubeconfig=/etc/kubernetes/kubelet.conf --node-ip=192.168.49.2 --runtime-request-timeout=15m "
}
```


### kubeProxyInfo

```json
{
	"cmdLine": "/usr/local/bin/kube-proxy --config=/var/lib/kube-proxy/config.conf --hostname-override=minikube "
}
```


### cloudProviderInfo

```json
{
    "providerMetaDataAPIAccess": true
}
```

### osRelease

```
NAME="Ubuntu"
VERSION="20.04.5 LTS (Focal Fossa)"
ID=ubuntu
ID_LIKE=debian
PRETTY_NAME="Ubuntu 20.04.5 LTS"
VERSION_ID="20.04"
HOME_URL="https://www.ubuntu.com/"
SUPPORT_URL="https://help.ubuntu.com/"
BUG_REPORT_URL="https://bugs.launchpad.net/ubuntu/"
PRIVACY_POLICY_URL="https://www.ubuntu.com/legal/terms-and-policies/privacy-policy"
VERSION_CODENAME=focal
UBUNTU_CODENAME=focal
```


### openedPorts

```json
{
	"tcpPorts": [
		{
			"Transport": "",
			"LocalAddress": "192.168.49.2",
			"LocalPort": 10259,
			"RemoteAddress": "0.0.0.0",
			"RemotePort": 0,
			"PID": 0,
			"Name": ""
		},
		{
			"Transport": "",
			"LocalAddress": "192.168.49.2",
			"LocalPort": 10257,
			"RemoteAddress": "0.0.0.0",
			"RemotePort": 0,
			"PID": 0,
			"Name": ""
		},
		{
			"Transport": "",
			"LocalAddress": "192.168.49.2",
			"LocalPort": 10248,
			"RemoteAddress": "0.0.0.0",
			"RemotePort": 0,
			"PID": 0,
			"Name": ""
		},
		{
			"Transport": "",
			"LocalAddress": "192.168.49.2",
			"LocalPort": 34961,
			"RemoteAddress": "0.0.0.0",
			"RemotePort": 0,
			"PID": 0,
			"Name": ""
		},
		{
			"Transport": "",
			"LocalAddress": "192.168.49.2",
			"LocalPort": 2379,
			"RemoteAddress": "0.0.0.0",
			"RemotePort": 0,
			"PID": 0,
			"Name": ""
		}
	],
	"udpPorts": [],
	"icmpPorts": []
}
```

## Contributions

Thanks to all our contributors! Check out our [CONTRIBUTING](https://github.com/kubescape/kubescape/blob/master/CONTRIBUTING.md) file to learn how to join them.

* Feel free to pick a task from the [issues](https://github.com/kubescape/host-scanner/issues?q=is%3Aissue+is%3Aopen+label%3A%22open+for+contribution%22), roadmap or suggest a feature of your own.
* [Open an issue](https://github.com/kubescape/host-scanner/issues/new/choose): we aim to respond to all issues within 48 hours.
* [Join the CNCF Slack](https://slack.cncf.io/) and then our [users](https://cloud-native.slack.com/archives/C04EY3ZF9GE) or [developers](https://cloud-native.slack.com/archives/C04GY6H082K) channel.