// +build integration

package nsvtep_test

import (
	"net"
	"os"
	"testing"

	"github.com/futurewei-cloud/plugins/alktron/nsvtep"
)

// sudo -E go test ./... -v -tags=integration -run NSvtepAttach to run this suite
// need to set env var TEST_NSVTEP_XXX, otherwise skipped

func TestNSvtepAttach(t *testing.T) {
	hostBr := os.Getenv("TEST_NSVTEP_HOST_BR")
	nspath := os.Getenv("TEST_NSVTEP_NETNS_PATH")
	if hostBr == "" || nspath == "" {
		t.Skipf("Skipping due to lack of TEST_NSVTEP_HOST_BR & TEST_NSVTEP_NETNS_PATH")
	}

	dev := "mynic"
	mac, _ := net.ParseMAC("de:ad:be:ef:a7:a7")
	ipnet := net.IPNet{
		IP:   net.ParseIP("10.0.0.4"),
		Mask: net.CIDRMask(24, 32),
	}
	gw := net.IPv4(10, 0, 0, 2)

	nsvtepManager := nsvtep.Manager{NSPath: nspath}
	if err := nsvtepManager.Attach(dev, mac, &ipnet, &gw, 88, hostBr); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNSvtepDetach(t *testing.T) {
	nspath := os.Getenv("TEST_NSVTEP_NETNS_PATH")
	hostBr := os.Getenv("TEST_NSVTEP_HOST_BR")
	if hostBr == "" || nspath == "" {
		t.Skipf("Skipping due to lack of TEST_NSVTEP_HOST_BR & TEST_NSVTEP_NETNS_PATH")
	}

	nsvtepManager := nsvtep.Manager{NSPath: nspath}
	if err := nsvtepManager.Detach("dummy", hostBr); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
