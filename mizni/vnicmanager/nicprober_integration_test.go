// +build integration

// sudo -E go test ./... -v -tags=integration -run TestXXXX to run specific test case
// need to set env var TEST_XXXX_XXX, otherwise skipped

package vnicmanager

import (
	"os"
	"testing"
	"time"
)

func TestNSDeviceReady(t *testing.T) {
	nsPath := os.Getenv("TEST_DEVPROBER_NS_PATH")   //e.g. "/run/netns/x"
	devName := os.Getenv("TEST_DEVPROBER_DEV_NAME") //e.g. "veth123"
	if nsPath == "" || devName == "" {
		t.Skipf("Skipping due to lack of TEST_DEVPROBER_NS_PATH & TEST_DEVPROBER_DEV_NAME")
	}

	nicProber := &nicProberWithTimeout{timeout: time.Second * 15}

	if err := nicProber.DeviceReady(devName, nsPath); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
