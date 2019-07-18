package ovsplug

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
)

// BridgePort represents the base of network devices able to join a local bridge
type BridgePort struct {
	NetlinkDev *netlink.Link
}

// JoinLinuxBridge joins the local linux bridge
func (bp BridgePort) JoinLinuxBridge(br LinuxBridge) error {
	if err := netlink.LinkSetMaster(*bp.NetlinkDev, br.bridge); err != nil {
		return fmt.Errorf("failed to join bridge %q: %v", br.Name, err)
	}

	return nil
}

// SetUp enables the link device
func (bp *BridgePort) SetUp() error {
	return netlink.LinkSetUp(*bp.NetlinkDev)
}

// ChangeMAC changes MAC address
func (bp *BridgePort) ChangeMAC(mac net.HardwareAddr) error {
	if err := netlink.LinkSetHardwareAddr(*bp.NetlinkDev, mac); err != nil {
		return fmt.Errorf("failed to change MAC address: %v", err)
	}

	devName := (*bp.NetlinkDev).Attrs().Name
	dev, err := netlink.LinkByName(devName)
	if err != nil {
		return fmt.Errorf("unexpected error, dev %q is corrupted: %v", devName, err)
	}

	bp.NetlinkDev = &dev
	return nil
}
