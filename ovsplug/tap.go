package ovsplug

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
)

// Tap represents a tap device
type Tap struct {
	BridgePort
	Name string
}

// Create creates a tap device with specific mac address on the local host
func (t *Tap) Create(mac *net.HardwareAddr) error {
	if t.BridgePort.NetlinkDev != nil {
		return fmt.Errorf("tap %q already created", t.Name)
	}

	la := netlink.LinkAttrs{Name: t.Name}
	if mac != nil {
		la.HardwareAddr = *mac
	}

	tuntap := &netlink.Tuntap{
		LinkAttrs: la,
		Mode:      netlink.TUNTAP_MODE_TAP,
		Flags:     netlink.TUNTAP_VNET_HDR | netlink.TUNTAP_NO_PI | netlink.TUNTAP_DEFAULTS,
	}

	if err := netlink.LinkAdd(tuntap); err != nil {
		return fmt.Errorf("failed to create tap %q: %v", t.Name, err)
	}

	dev, err := netlink.LinkByName(t.Name)
	if err != nil {
		return fmt.Errorf("post-create failure on creating tap %q: %v", t.Name, err)
	}

	// LinkAdd seems unable to set mac address properly; need separate mac update
	if mac != nil {
		if err = netlink.LinkSetHardwareAddr(dev, *mac); err != nil {
			// todo: cleanup - remove faulty tap dev
			return fmt.Errorf("post-create failure on %q mac addr update: %v", t.Name, err)
		}
	}

	dev, err = netlink.LinkByName(t.Name)
	if err != nil {
		// todo: cleanup - remove faulty tap dev
		return fmt.Errorf("post-create failure on getting tap %q link: %v", t.Name, err)
	}

	t.BridgePort = BridgePort{NetlinkDev: &dev}
	return nil
}

// todo: add Remove method
