# mizni
[cni](https://github.com/containernetworking/cni/blob/master/SPEC.md) plugin integrating the new network control plane [alcor](https://github.com/futurewei-cloud/AlcorControlAgent.git) and data plane [mizar](https://github.com/futurewei-cloud/Mizar.git) into [alkaid](https://github.com/futurewei-cloud/alkaid.git)

See [design spec](https://github.com/futurewei-cloud/alkaid/blob/master/docs/design-proposals/network/NICAndVPCSupportInAlkaid.md) for broader information of project background, design considerations, and impacts upon other components.

## target cni version
0.3.1

## interface manipulation
locate the net device (tap for now) provisioned by alcor/mizar, move it into cni netns with the proper ip address, route setting, etc.

## IPAM
no IPAM needed; alcor/mizar would assign proper ip address to the net device.
