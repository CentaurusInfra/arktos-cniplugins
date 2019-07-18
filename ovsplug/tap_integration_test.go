// +build integration

package ovsplug_test

import (
	"net"
	"os"
	"strings"
	"testing"

	"github.com/futurewei-cloud/alktron/ovsplug"
)

// sudo -E go test ./... -tags=integration -v -run Tap to run this integration test set
// need to set TEST_TAP env var, otherwise skipped

func TestTapCreateAndChangeMAC(t *testing.T) {
	tapName := os.Getenv("TEST_TAP")
	if tapName == "" {
		t.Skipf("Skipping due to lcak of TEST_TAP env var")
	}

	mac, _ := net.ParseMAC("00:12:34:56:78:AB")

	tap := &ovsplug.Tap{Name: tapName}
	err := tap.Create(&mac)
	if err != nil {
		t.Fatalf("unpected error: %v", err)
	}

	t.Logf("tap detail after created: %v", *tap.BridgePort.NetlinkDev)
	if tap.BridgePort.NetlinkDev == nil {
		t.Errorf("expecting netlink dev, got nil")
	}

	macInNetlink := (*tap.BridgePort.NetlinkDev).Attrs().HardwareAddr.String()
	if !strings.Contains(macInNetlink, "12:34:56:78") {
		t.Errorf("expecting mac has 12:34:56:78, got %q", macInNetlink)
	}

	mac, _ = net.ParseMAC("88:77:66:55:44:33")
	if err := tap.ChangeMAC(mac); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	t.Logf("tap detail after mac change: %v", *tap.BridgePort.NetlinkDev)
	macInNetlink = (*tap.BridgePort.NetlinkDev).Attrs().HardwareAddr.String()
	if !strings.Contains(macInNetlink, "88:77:66:55") {
		t.Errorf("expecting mac has 88:77:66:55, got %q", macInNetlink)
	}
}
