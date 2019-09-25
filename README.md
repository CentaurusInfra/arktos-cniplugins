# cni plugins
This project hosts two cni plugins developed for [alkaid](https://github.com/futurewei-cloud/alkaid.git)
1. [alktron](alktron) - cni plugin to integrate with OpenStack Neutron
1. [mizni](mizni) - cni plugin to integrate with the high-performance high-scale network backend including control plane [alcor](https://github.com/futurewei-cloud/AlcorControlAgent.git) and data plane [mizar](https://github.com/futurewei-cloud/Mizar.git). 

# build instructions
Please use golang 1.12 or later

The recommended way is building from inside of the sub folder, e.g.
```bash
$ cd mizni
$ go build .
```
