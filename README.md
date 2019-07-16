# alktron
[cni](https://github.com/containernetworking/cni/blob/master/SPEC.md) plugin integrating [alkaid](https://github.com/futurewei-cloud/alkaid.git) with [neutron](https://github.com/openstack/neutron)

## target cni version
0.3.1

## interface creating
creates a tap dev hooked to backend like qbr-qvb-qvo for typical ovs hybrid bridge ML2 plugin. If possible, stives to support other ML2 plugins.

tap dev should have mac address as defined in neutron port.

## IPAM
no IPAM needed; DHCP of neutron ML2 plugin shall be used implicitly.
