# Expected data passed from Kubelet through CRI to alktron

alktron, as a cni plugion, is expecting neutron port related information from the upstream caller (e.g. containerd, virtlet) in form of cni extra args, like below
```
...;VPC=demo;NICs=[{"portid":"93881c89-89ce-407e-a775-d8d3319431d5"}]
```

It is CRI runtime's responsibility to pass proper cni args to alktron. We have patched version of containerd & virtlet for end-to-end proof of concept purpose; they are able to accept vpc/portid from kubelet by the above protocol and pass further to alktron as expected. 

As the first link of such process chain, kubelet has to pass the needed data (namely neutron project & port id) to CRI runtime.

There are various ways for kubelet to pass data to CRI downstream, including in new fields by extending or breaking CRI mechanism. One easy (and recommended) approach, in comliance of CRI spec, is packing the infomation into Annotations (with key of VPC and NIVs, respectively) of PodSandboxConfig

``` golang
type PodSandboxConfig struct {
        // Pod name of the sandbox.
        Name string
        // Pod UID of the sandbox.
        Uid string
        // Pod namespace of the sandbox.
        Namespace string
        ...
        // Key-value pairs that may be used to scope and select individual resources.
        Labels map[string]string
        // Unstructured key-value map that may be set by the kubelet to store and
        // retrieve arbitrary metadata. This will include any annotations set on a
        // pod through the Kubernetes API.
        Annotations map[string]string
        // Optional configurations specific to Linux hosts.
        CgroupParent string
}
```

The official kubelet would put pod.metadata.annotations to PodSandboxConfig.Annotations; if alktron needed data were put in pod annotations, it would be passed to CRI runtime. One pod example is as below

```
$ cat pod-alktron-y.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  labels:
    name: nginx
  annotations:
    VPC: demo
    NICs: "[{\"portid\":\"1677d54c-5c67-46bf-aaae-309b929f6d3b\"}]"
spec:
  containers:
  - name: nginx
    image: nginx
    ports:
    - containerPort: 80
```
