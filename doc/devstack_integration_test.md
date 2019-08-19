# Testing on Devstack

This is system level acceptance testing integrating with devstack only. It should also work on a full fledged multi-node OpenStack system, if one is available.

## Prerequisites
* DevStack

| property | value | 
| -------- | ----- |
| version | Queen (should also work with later versions) |
| core plugin | ml2 |
| mechanism driver | openvswicth, linuxbridge (a.k.a. hybrid ovs) |
| firewall driver | iptables_hybrid |

prepare the test port
```bash
# given 7f4a0355-cfab-4c40-9be0-d58f7876ec0c is the network id of interest
openstack port create x --network 7f4a0355-cfab-4c40-9be0-d58f7876ec0c
# identify the port id generated, e.g. f8a6471c-249a-4cd4-ad49-914bfdd95da1
```
keep the port id for later use of cni add/del op

## Configuration & Binary files
* /opt/cni/binary/alktron

copy alktron binary to /opt/cni/bin/ on the devstack machine
```bash
# build using go tool (>=1.12) inside of project
go build ./
```

* /etc/cni/net.d/mynet.conf
```json
{
    "cniVersion":"0.3.1",
	"name": "mynet",
	"type": "alktron"
}
```
* /etc/alktron/neutron.json

given \<hostname\> is the hostname of devstack machine
```json
{
  "user": "admin",
  "password": "secret",
  "identity_url":"http://127.0.0.1/identity",
  "host":"<hostname>",
}
```

## Version op
```bash
export CNI_COMMAND=VERSION
/opt/cni/bin/alktron
{"cniVersion":"0.4.0","supportedVersions":["0.1.0","0.2.0","0.3.0","0.3.1"]}
```

## Add op
```bash
ip netns add x

export CNI_COMMAND=ADD
export CNI_ARGS='VPC=demo;NICs=[{"portid":"f8a6471c-249a-4cd4-ad49-914bfdd95da1"}]'
export CNI_CONTAINERID=cafe123456
export CNI_NETNS=/var/run/netns/x	
export CNI_IFNAME=eth9
export CNI_PATH=/opt/cni/bin/

cat /etc/cni/net.d/mynet.conf | /opt/cni/bin/alktron
```

Successful running should yield 0 exit code, and well formatted result like below:
```json
{
    "cniVersion": "0.3.1",
    "interfaces": [
        {
            "name": "eth0",
            "mac": "fa:16:3e:e7:16:7f",
            "sandbox": "cafe123456"
        }
    ],
    "ips": [
        {
            "version": "4",
            "interface": 0,
            "address": "10.0.0.9/26",
            "gateway": "10.0.0.1"
        }
    ],
    "dns": {}
}
```

look further inside netns, eth0 are up and mac/ip add properly assigned 
```bash
ip netns exec x ip a
```
it would display network information like below
```text
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
    inet6 ::1/128 scope host 
       valid_lft forever preferred_lft forever
33: eth0@if32: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default 
    link/ether fa:16:3e:e7:16:7f brd ff:ff:ff:ff:ff:ff link-netnsid 0
    inet 10.0.0.9/26 brd 10.0.0.63 scope global eth0
       valid_lft forever preferred_lft forever
    inet6 fd76:3b87:1966:0:f816:3eff:fee7:167f/64 scope global mngtmpaddr dynamic 
       valid_lft 86379sec preferred_lft 14379sec
    inet6 fe80::f816:3eff:fee7:167f/64 scope link 
       valid_lft forever preferred_lft forever
```

## Del op

Continue w/ the above Add op, do following:
```bash
export CNI_COMMAND=DEL
cat /etc/cni/net.d/mynet.conf | /opt/cni/bin/alktron
ip netns del x
```
