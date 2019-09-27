# multiple nics support

Mizni plugin is designed to support multiple nics in one vpc. Multiple nics across different vpcs are not in scope at this moment.

Below illustrates how we can verify its multi vnics feature.

## At standalone cni integration env
The cni integration env is set up as instructed in [simulated_integration](simulated_integration.md). Extra steps are listed as below:
1. run following commands inside netns vpc-nsdemo to provision the second port
```bash
ip tuntap add dev veth12345678-01 mode tap
ip link set dev veth12345678-01 up
ip addr add 10.100.201.32/24 dev veth12345678-01
ip route add default via 10.100.201.1 metric 101 
``` 
2. ensure netns x created
3. invoke cni plugin by 
```bash
export CNI_ARGS='VPC=demo;NICs=[{"portid":"f8a6471c-249a-4cd4-ad49-914bfdd95da1"},{"portid":"12345678-01aa-aaaa-aaaa-aaaaaaaaaaaa"}]'
cat /etc/cni/etc/net.d/mizni.conf | /opt/cni/bin/mizni
```
You should see result like
```json
{
    "cniVersion": "0.3.1",
    "interfaces": [
        {
            "name": "eth0",
            "mac": "c6:6d:29:34:87:ec",
            "sandbox": "cafe123456"
        },
        {
            "name": "eth1",
            "mac": "02:fb:76:06:15:8e",
            "sandbox": "cafe123456"
        }
    ],
    "ips": [
        {
            "version": "4",
            "interface": 0,
            "address": "10.100.200.18/24",
            "gateway": "10.100.200.1"
        },
        {
            "version": "4",
            "interface": 1,
            "address": "10.100.201.32/24",
            "gateway": "10.100.201.1"
        }
    ],
    "dns": {}
}
```
4. inside netns x, find eth1 besides eth0, and two default routing entries.

## In kubernetes cluster
Mizni by itself is able to hook up connectivities for multiple vnics. However, whether the workload is able to identify or utilize the nics other than the primary one(usually eth0) depends on other factors beyond this plugin binary, such as CRI runtime in use, workload settings. 

In general case of VM workloads, VM should be configured to get ip addresses via HDCP protocol on boot for all of its nics, otherwise you may see nics having no ip address assigned.