# cni plugins
this project hosts two cni plugins developed for [alkaid](https://github.com/futurewei-cloud/alkaid.git)
1. alktron - cni plugin to integrate with OpenStack Neutron
1. mizni - cni plugin to integrate with the high-performance scalable netowk backend (control plane [alcor](https://github.com/futurewei-cloud/AlcorControlAgent.git) and data plane [mizar](https://github.com/futurewei-cloud/Mizar.git)) 

# build requirement
golang 1.12 or later

One way to build the binary is building from inside of the sub folder
```bash
$ cd mizni
$ go build .
```