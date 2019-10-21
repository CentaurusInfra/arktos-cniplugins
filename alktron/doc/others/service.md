# Kubernetes Service Test Report

## Purpose
To track the verifications Kubernetes Service created based on Alktron (neutron cni plugin).

## Aimed Scenarios
See [integration-test](../integration_test.md) for general instructions to prepare the onebox test env.
Given Devstack has already created 3 port, namely x, y, z, which has IP address 10.0.0.14, 10.0.0.8, 10.0.0.18, respectively. x, y & z ports are configured inside VPC private-subnet 10.0.0.0/16.
### vm based service
a vm-pod is used as the single backend pod, with the custom label, using pod-vm.yaml
```yaml (partial)
apiVersion: v1
kind: Pod
metadata:
  name: pod-vm-z
  labels:
    mylabel: pod-vm
  annotations:
    VPC: demo
    NICs: "[{\"portid\":\"z\"}]"
...
```
a service is created using following service-vm.yaml
```yaml
kind: Service
apiVersion: v1
metadata:
  name: pod-vm-svc
spec:
  type: ClusterIP
  selector:
    mylabel: pod-vm
  ports:
  - name: ssh
    port: 22
```
### container-pod based service
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: pod-container-y
  labels:
    mylabel: pod-container
  annotations:
    VPC: demo
    NICs: "[{\"portid\":\"y\"}]"
spec:
  containers:
  - name: nginx
    image: nginx
    imagePullPolicy: IfNotPresent
```
service yaml is as below:
```yaml
kind: Service
apiVersion: v1
metadata:
  name: pod-container-svc
spec:
  type: ClusterIP
  selector:
    mylabel: pod-container
  ports:
  - name: http
    port: 80
```
### service consumer
a vm pod is created using pod-vm-client.yaml
```yaml(partial)
apiVersion: v1
kind: Pod
metadata:
  name: consumer-vm-x
  annotations:
    VPC: demo
    NICs: "[{\"portid\":\"x\"}]"
...
```
from inside the consumer vm, access the targeted pod via service cluster IP. If it does not work, record errors, and some troubleshooting data if relevant. 

## vm-pod service result
service and end points are created as expected
```bash
$ kuberctl get service
pod-vm-svc   ClusterIP   10.105.119.92   <none>        22/TCP    78m

$ kubectl get ep
pod-vm-svc   10.0.0.18:22           80m
```
consumer cannot use service IP to ssh into target vm
```bash
$ ssh 10.105.119.92
ssh: Connection to cirros@10.105.119.92:22 exited: Connect failed: No route to host
```
tcmdump traces at various interfaces
* qbr-{consumer}
may need to run sudo sysctl -w net.bridge.bridge-nf-call-iptables=0 to allow packet broadcast at qbr.
* qvo-{consumer}
```bash
# tcpdump -eni qbr63902e5f-8b
13:08:13.834385 fa:16:3e:47:44:73 > fa:16:3e:17:5f:34, ethertype IPv4 (0x0800), length 74: 10.0.0.14.47134 > 10.105.119.92.22: Flags [S], seq 743410733, win 29200, options [mss 1460,sackOK,TS val 958850269 ecr 0,nop,wscale 4], length 0
13:08:14.840180 fa:16:3e:47:44:73 > fa:16:3e:17:5f:34, ethertype IPv4 (0x0800), length 74: 10.0.0.14.47134 > 10.105.119.92.22: Flags [S], seq 743410733, win 29200, options [mss 1460,sackOK,TS val 958851275 ecr 0,nop,wscale 4], length 0
13:08:16.856259 fa:16:3e:47:44:73 > fa:16:3e:17:5f:34, ethertype IPv4 (0x0800), length 74: 10.0.0.14.47134 > 10.105.119.92.22: Flags [S], seq 743410733, win 29200, options [mss 1460,sackOK,TS val 958853291 ecr 0,nop,wscale 4], length 0
13:08:16.894642 fa:16:3e:17:5f:34 > fa:16:3e:47:44:73, ethertype IPv4 (0x0800), length 102: 172.24.4.27 > 10.0.0.14: ICMP host 10.105.119.92 unreachable, length 68
13:08:16.894673 fa:16:3e:17:5f:34 > fa:16:3e:47:44:73, ethertype IPv4 (0x0800), length 102: 172.24.4.27 > 10.0.0.14: ICMP host 10.105.119.92 unreachable, length 68
13:08:16.894682 fa:16:3e:17:5f:34 > fa:16:3e:47:44:73, ethertype IPv4 (0x0800), length 102: 172.24.4.27 > 10.0.0.14: ICMP host 10.105.119.92 unreachable, length 68
```
where 172.24.4.27 is the gateway configured in neutron for VPC public-subnet 10.24.4.0/24.
* qvo-{target-vm}
```bash
# tcpdump -eni qvo{target-pod-vm}
(nothing captured)
```

## container-pod service result
service and end point are created as expected:
```bash
$ kubectl get svc
pod-container-svc   ClusterIP   10.98.245.15    <none>        80/TCP    9s
$ kubectl get ep
pod-container-svc   10.0.0.8:80            13s
```
consumer cannot get back nginx response
```bash
$ curl 10.98.245.15:80
curl: (7) Failed to connect to 10.98.245.15 port 80: No route to host
```
tcpdump traces at qvo-{consumer}
```text
13:28:10.515481 fa:16:3e:47:44:73 > fa:16:3e:17:5f:34, ethertype IPv4 (0x0800), length 74: 10.0.0.14.35988 > 10.98.245.15.80: Flags [S], seq 48787203, win 29200, options [mss 1460,sackOK,TS val 3325461308 ecr 0,nop,wscale 4], length 0
13:28:11.544215 fa:16:3e:47:44:73 > fa:16:3e:17:5f:34, ethertype IPv4 (0x0800), length 74: 10.0.0.14.35988 > 10.98.245.15.80: Flags [S], seq 48787203, win 29200, options [mss 1460,sackOK,TS val 3325462337 ecr 0,nop,wscale 4], length 0
13:28:13.560084 fa:16:3e:47:44:73 > fa:16:3e:17:5f:34, ethertype IPv4 (0x0800), length 74: 10.0.0.14.35988 > 10.98.245.15.80: Flags [S], seq 48787203, win 29200, options [mss 1460,sackOK,TS val 3325464353 ecr 0,nop,wscale 4], length 0
13:28:13.566610 fa:16:3e:17:5f:34 > fa:16:3e:47:44:73, ethertype IPv4 (0x0800), length 102: 172.24.4.27 > 10.0.0.14: ICMP host 10.98.245.15 unreachable, length 68
13:28:13.566674 fa:16:3e:17:5f:34 > fa:16:3e:47:44:73, ethertype IPv4 (0x0800), length 102: 172.24.4.27 > 10.0.0.14: ICMP host 10.98.245.15 unreachable, length 68
13:28:13.566682 fa:16:3e:17:5f:34 > fa:16:3e:47:44:73, ethertype IPv4 (0x0800), length 102: 172.24.4.27 > 10.0.0.14: ICMP host 10.98.245.15 unreachable, length 68
``` 