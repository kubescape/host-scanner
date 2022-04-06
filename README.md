# Kubescape host-sensor
## Description
This component is a data acquisition component in the Kubescape project. Its goal is to collect information about the Kubernetes node host for further security posture evaluation in Kubescape.

## Deployment
Host-sensor is deployed as a privileged Kubernetes DaemonSet in the cluster. It publishes an API for clients to read host infromation.

