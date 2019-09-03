## multi nic support

alktron supports multi nics through args parameters. Depending on the runtime and images in use, pods may or not be able to utilize the nics other than the promary one. For example, virtlet supports multi nics, however, if the vm image does not expect nics beyond eth0, the nics other than eth0 may not get ip addr, even not in up status, after the vm starts. This is not the limitation incurred by alktron; it is actually caused by vm start scripts - usually vm tries to get ip address lease via dhcp for the primary nic only, leaving other nics as is in down state.

Below is an experimental session that 2 nics were attached to a cirros vm: eth0 was up and got ip addr, eth1 down and none ip. By manually issuing dhcp client command, eth1 gets ip lease and is connected to network.

vm-w-2-nics.yaml is based on virlet sample cirros-vm.yaml 
```json
apiVersion: v1
kind: Pod
metadata:
  name: cirros-vm-alktron-yz
  annotations:
    VPC: demo
    NICs: "[{\"portid\":\"07591fbc-f7f7-4c0b-8108-a224f23a2862\"},{\"portid\":\"82bcd9b6-108d-4906-8ec6-08435746b4dd\"}]"
    # This tells CRI Proxy that this pod belongs to Virtlet runtime
    kubernetes.io/target-runtime: virtlet.cloud
    # CirrOS doesn't load nocloud data from SCSI CD-ROM for some reason
    VirtletDiskDriver: virtio
    # inject ssh keys via cloud-init
    VirtletSSHKeys: |
      ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCaJEcFDXEK2ZbX0ZLS1EIYFZRbDAcRfuVjpstSc0De8+sV1aiu+dePxdkuDRwqFtCyk6dEZkssjOkBXtri00MECLkir6FcH3kKOJtbJ6vy3uaJc9w1ERo+wyl6SkAh/+JTJkp7QRXj8oylW5E20LsbnA/dIwWzAF51PPwF7A7FtNg9DnwPqMkxFo1Th/buOMKbP5ZA1mmNNtmzbMpMfJATvVyiv3ccsSJKOiyQr6UG+j7sc/7jMVz5Xk34Vd0l8GwcB0334MchHckmqDB142h/NCWTr8oLakDNvkfC1YneAfAO41hDkUbxPtVBG5M/o7P4fxoqiHEX+ZLfRxDtHB53 me@localhost
    # set root volume size
    VirtletRootVolumeSize: 1Gi
spec:
  # This nodeSelector specification tells Kubernetes to run this
  # pod only on the nodes that have extraRuntime=virtlet label.
  # This label is used by Virtlet DaemonSet to select nodes
  # that must have Virtlet runtime
  nodeSelector:
    extraRuntime: virtlet

  containers:
  - name: cirros-vm
    # This specifies the image to use.
    # virtlet.cloud/ prefix is used by CRI proxy, the remaining part
    # of the image name is prepended with https:// and used to download the image
    image: virtlet.cloud/cirros
    imagePullPolicy: IfNotPresent
    # tty and stdin required for `kubectl attach -t` to work
    tty: true
    stdin: true
    resources:
      limits:
        # This memory limit is applied to the libvirt domain definition
        memory: 160Mi
```

```bash
kubectl apply -f vm-w-2-nics.yaml

kubectl attach -it cirros-vm-alktron-yz
> sudo cirros-dhcpc up eth1
```

If the vm image has the ability to bring up other nic like below, both eth0 and eth1 should be able to get proper ip addresses automatically:
```bash
cat /etc/network/interface
auto lo
iface lo inet loopback

# primary nic
auto eth0
iface eth0 inet dhcp

# secondary nic
auto eth1
iface eth1 inet dhcp
```