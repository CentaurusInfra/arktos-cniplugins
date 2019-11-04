# Integrating with Neutron
Alktron is the cni plugin integrating with OpenStack Neutron service. Alktron is responsible for integration with Neutron at node level, similar to what Nova does in OpenStack system.

In Kubernetes/Alkaid, some terms are significant for vnic, and of diffent names in Neutron, as below:

| pod.spec | Nuetron | notes |
| --- | --- | --- |
| tenant | domain | tenant of Alkaid (current only default tenant is support; multi tenancy support is on radar) |
| vpc | project | VPC of a tenant in Alkaid |
| nic subnet | subnet name | subnet of CIDR in VPC |
| nic portId | port id | physical interface assignment |
| nic name | |  interface name; should have eth0 (default) in pod |


## /etc/alktron/neutron.json
This json file specifies the connection configuration for Alktron to talk to Neutron.

| name | notes | sample value |
| --- | --- | --- |
| user | Neutron user name | "admin" |
| password | password of the user | "secret" |
| identity_url | OpenStack identity service URL | "http://127.0.0.1/identity" |
| host | node name as in Neutron | commonly hostname |
| interval_in_ms | interval to probe Neutron port status | default to 500 ms | 
| timeout_in_sec | timeout waiting for Neutron port being ready | default to 15 seconds | 
| region | region in OpenStack Neutron system | default to "ReegionOne" |
