# cni simulated integration without backend Miza/Alcor

Though mizni is developed to work for Mizar/Alcor backedn, it is actually able to run without them in certian testing circumstance, given the proper _simulated network setup_ in place. This is to verify the basic operations of mizni in a minimal env.

Except for mizni build, all steps should be run as privileged user.

## test env requirements
1. Linux machine (verified on ubuntu 18.04 LTS)
1. ethtool installed

## Simulated network setup
```bash
ip netns add vpc-nsdemo
ip netns exec vpc-nsdemo ip link set dev lo up
ip netns exec vpc-nsdemo ip tuntap add dev vethf8a6471c-24 mode tap
ip netns exec vpc-nsdemo ip link set dev vethf8a6471c-24 up
ip netns exec vpc-nsdemo ip addr add 10.100.200.18/24 dev vethf8a6471c-24
ip netns exec vpc-nsdemo ip route add default via 10.100.200.1

```

## cni related setup
* build mizni
```bash
cd mizni
go build .
```

* copy mizni to the test machine under /opt/cni/bin/

* ensure cni netconf file exists as /etc/cni/net.d/mizni.conf
```json
{
  "cniVersion":"0.3.1",
  "name": "mynet",
  "type": "mizni"
}
```

## Add op simulated test
* prepare cni netns
```bash
ip netns add x
```

* set up cni envs
```bash
export CNI_COMMAND=ADD
export CNI_ARGS='VPC=demo;NICs=[{"portid":"f8a6471c-249a-4cd4-ad49-914bfdd95da1"}]'
export CNI_CONTAINERID=cafe123456
export CNI_NETNS=/run/netns/x	
export CNI_IFNAME=eth9
export CNI_PATH=/opt/cni/bin/
```

* run cni command
```bash
cat /etc/cni/net.d/mizni.conf | /opt/cni/bin/mizni
```

* you should be able to see result like below:
```json
{
    "cniVersion": "0.3.1",
    "interfaces": [
        {
            "name": "eth0",
            "mac": "c6:6d:29:34:87:ec",
            "sandbox": "cafe123456"
        }
    ],
    "ips": [
        {
            "version": "4",
            "interface": 0,
            "address": "10.100.200.18/24",
            "gateway": "10.100.200.1"
        }
    ],
    "dns": {}
}

```

## Del op simulated test
after the Add op testing, continue with following commands

```bash
export CNI_COMMAND=DEL
cat /etc/cni/net.d/mizni.conf | /opt/cni/bin/mizni
```

and verify that the veth device showed up in vpc-nsdemo namespace.
