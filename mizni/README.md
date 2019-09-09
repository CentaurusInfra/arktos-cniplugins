# mizni
[cni](https://github.com/containernetworking/cni/blob/master/SPEC.md) plugin integrating [mizar](https://github.com/futurewei-cloud/Mizar.git), [alioth](https://github.com/futurewei-cloud/AliothController.git) & its agent [alcor](https://github.com/futurewei-cloud/AlcorControlAgent.git) into [alkaid](https://github.com/futurewei-cloud/alkaid.git)

See [design spec](https://github.com/futurewei-cloud/alkaid/blob/master/docs/design-proposals/network/NICAndVPCSupportInAlkaid.md) for broader information of project background, design considerations, and impacts upon other components.

## target cni version
0.3.1

## interface manipulation
locate the net device (tap for now) provisioned by mizar/alioth agent, move it into cni netns with the proper ip address, route setting, etc.

## IPAM
no IPAM needed; mizar/alioth/alcore would assign proper ip address to the net device.
