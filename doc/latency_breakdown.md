## Latency Breakdown

This doc is for information only; actual latency may fractuate a lot depending on the workloads, machine situation, netowrking conditions etc. 

Enclosed latency data was collected using alktron v0.1.1, on a DevStack onebox machine with no workload.

NIC add op user experience weights much more than del op. We only collect data for add op at this moment. A golang runtime trace instrumented binary was used for such purpose.

### Single NIC CNI ADD op latency
| subphase | duration (ms) |
| ---:      |  ---:     |
| port-bind | 1906 |
| ovs-hybrid-init | 3 |
| ovs-hybrid-vif-plug | 33 |
| ipaddr-gw-parse | 362 |
| attach-tap | 14 |
| wait-till-port-active | 1295 |
| overall wall-clock latency | 3982 |

