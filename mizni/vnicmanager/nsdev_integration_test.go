// +build integration

// sudo -E go test ./... -v -tags=integration -run TestGetDevNetConf to run this suite
// need to set env var TEST_XXXX_XXX, otherwise skipped

package vnicmanager

import (
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
	ipnet, gw, mac, mtu, err := devGetter.GetDevNetConf(devName, vpcNS)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf("ipnet=%s, gw=%s, mac=%s, mtu=%d", ipnet, gw, mac, mtu)
}
