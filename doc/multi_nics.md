## multi nic support

alktron supports multi nics for vm. However, the nics other than the primary one may not get ip addr, even not in up status, when the vm starts. This is not the problem caused by alktron; it is actually caused by vm - usually vm only triees to get ip address lease via dhcp for the primary nic, which lets other nics as is in down state.

Below is a session that 2 nics were attached to a cirros vm; only eth0 was up and got ip addr, eth1 none. By manually issuing dhcp client command, eth1 got ip lease and was connected to network.

```bash
cat vm-w-2-nics.yaml
apiVersion: v1
kind: Pod
metadata:
  name: cirros-vm-alktron-yz
  annotations:
    VPC: demo
    NICs: "[{\"portid\":\"07591fbc-f7f7-4c0b-8108-a224f23a2862\"},{\"portid\":\"82bcd9b6-108d-4906-8ec6-08435746b4dd\"}]"
    ...

kubectl apply -f vm-w-2-nics.yaml

kubectl attach -it cirros-vm-alktron-yz
> sudo cirros-dhcpc up eth1
```

If vm image has the ability to bring up other nic like below, nic should be able to get proper ip automatically:
```bash
cat /etc/network/interface
auto lo
iface lo inet loopback

# primary nic
auto eth0
iface eth0 inet dhcp

# second nic
auto eth1
iface eth1 inet dhcp
```