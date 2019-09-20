// +build integration

// sudo -E go test ./... -v -tags=integration -run TestXXXX to run specific test case
// need to set env var TEST_XXXX_XXX, otherwise skipped

package vnicmanager

import (
	"net"
	"os"
	"testing"
)

func TestGetDevNetConf(t *testing.T) {
	vpcNS := os.Getenv("TEST_ALCOR_NETNS_PATH") //e.g. "/run/netns/x"
	devName := os.Getenv("TEST_ALCOR_DEV_NAME") //e.g. "veth123"
	if vpcNS == "" || devName == "" {
		t.Skipf("Skipping due to lack of TEST_ALCOR_NETNS_PATH & TEST_ALCOR_DEV_NAME")
	}

	devGetter := nsdev{}
	ipnet, gw, metric, mac, mtu, err := devGetter.GetDevNetConf(devName, vpcNS)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf("ipnet=%s, gw=%s, metric=%d, mac=%s, mtu=%d", ipnet, gw, metric, mac, mtu)
}

func TestNSMigrate(t *testing.T) {
	nsFrom := os.Getenv("TEST_NSMIGRATE_NSPATH_FROM") //e.g. "/run/netns/x"
	devFrom := os.Getenv("TEST_NSMIGRATE_NAME_FROM")  //e.g. "veth123"
	nsTo := os.Getenv("TEST_NSMIGRATE_NSPATH_TO")     //e.g. "/run/netns/y"
	devTo := os.Getenv("TEST_NSMIGRATE_NAME_TO")      //e.g. "eth0"
	if nsFrom == "" || devFrom == "" || nsTo == "" || devTo == "" {
		t.Skipf("Skipping due to lack of TEST_NSMIGRATE_NSPATH_FROM & TEST_NSMIGRATE_NAME_FROM & TEST_NSMIGRATE_NSPATH_TO & TEST_NSMIGRATE_NAME_TO")
	}

	mover := &nsdev{}
	ipnet := &net.IPNet{IP: net.ParseIP("10.0.36.8"), Mask: net.CIDRMask(16, 32)}
	gw := net.ParseIP("10.0.0.1")
	metric := 101
	mtu := 1460

	if err := mover.Migrate(devFrom, nsFrom, devTo, nsTo, ipnet, &gw, metric, mtu); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
