## Latency Breakdown

This doc is for information only; actual latency may fractuate a lot depending on the workloads, machine situation, netowrking conditions etc. 

Enclosed latency data was collected using mizni v0.3.0, on an onebox machine simulating the network backend by a tap device; it has no workload.

NIC add op user experience weights much more than del op. We only collect data for add op at this moment. A golang runtime trace instrumented binary was used for such purpose.

### Single NIC CNI ADD op latency
| subphase | duration (ms) |
| ---:      |  ---:     |
| probe dev | 0.265 |
| get net setting | 0.323 |
| move dev to cni ns | 28.581 |
| reapplying settings to dev | 20.981 |
| mizar requisite commands | 7.789 |
| overall wall-clock latency | 58 |

