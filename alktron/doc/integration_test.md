# How to run module level integration testing

<i>This doc covers only the __module level__ integration test, not the system acceptance integration. This is a task in pregress - let's start from documenting the typical cases, add those not easy to figure out along the way. Providing instructions covering all cases is not the priority, however.</i>

Generally, these integration test cases are expected to run manually (as a side note, you are welcome to to convert some into automation). For privileged integration test cases (e.g. those manipulating network stack and netns), testing should be run as root (or sudo -E) 
<p/>

Integration test case generally expects the testing env has made some setup properly before hand, in order for test case to work on top of. For instance, TestLoadNeutronConfig is expecting well-formatted config content in /etc/alktron/neutron.json, plus some env vars, as later illustrates.
<p/>

The test case setup is test case specific. Please look at test code to figure them out. Usually they are not hard, since they are mostly single module integration anyway.

## Integration of host file system & env vars
### TestLoadNeutronConfig
to verify combinatory load of configuration settings from /etc/alktron/neutron.json file and env vars

```
$ cat /etc/alktron/neutron.json 
{
  "user": "admin",
  "identity_url":"http://127.0.0.1/identity",
  "Host":"localhost",
  interval_in_ms": 123,
  "timeout_in_sec": 14,
}

$ export ALKTRON_PASSWORD="secret"

$ go test . -tags=integration -v -run TestLoadNeutronConfig 
=== RUN   TestLoadNeutronConfig
--- PASS: TestLoadNeutronConfig (0.00s)
    conf_integration_test.go:16: config detail: {admin secret http://127.0.0.1/identity localhost 123 14}
PASS
ok      github.com/futurewei-cloud/alktron      0.007s
```


## Integration testing of netns & network devices
### TestNSvtepDetach
to verify veth pair across the root/cni netns can be removed properly. Its setup can be done like
```
$ sudo ip link add dev tap12345678901 type veth peer name qvn12345678901
$ sudo ip net add cni-12345
$ sudo ip link set dev qvn12345678901 netns cni-12345
$ sudo ip link set dev tap12345678901 up
$ sudo brctl addbr qbr12345678901
$ sudo brctl addif qbr12345678901 tap12345678901
```
running integration test as following
```
$ export TEST_NSVTEP_NETNS_PATH=/run/netns/cni-12345
$ export TEST_NSVTEP_HOST_BR=qbr12345678901
$ sudo -E go test ./... -v -tags=integration -run TestNSvtepDetach
```
verifying 
* successful removal of tap12345678901/qvn12345678901 vath pair
* qbr12345678901 has no tap12345678901 interface attached
