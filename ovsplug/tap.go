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

// NewTap creates a tap device with specific name and mac address on the local host, ensures in up state
func NewTap(name string, mac *net.HardwareAddr) (*Tap, error) {
	// todo: cleanup - remove faulty tap dev
	la := netlink.LinkAttrs{Name: name}
	if mac != nil {
		la.HardwareAddr = *mac
	}

	tuntap := &netlink.Tuntap{
		LinkAttrs: la,
		Mode:      netlink.TUNTAP_MODE_TAP,
		Flags:     netlink.TUNTAP_VNET_HDR | netlink.TUNTAP_NO_PI | netlink.TUNTAP_DEFAULTS,
	}

	if err := netlink.LinkAdd(tuntap); err != nil {
		return nil, fmt.Errorf("failed to create tap %q: %v", name, err)
	}

	dev, err := netlink.LinkByName(name)
	if err != nil {
		return nil, fmt.Errorf("post-create failure on creating tap %q: %v", name, err)
	}

	// LinkAdd seems unable to set mac address properly; need separate mac update
	if mac != nil {
		if err = netlink.LinkSetHardwareAddr(dev, *mac); err != nil {
			// todo: cleanup - remove faulty tap dev
			return nil, fmt.Errorf("post-create failure on %q mac addr update: %v", name, err)
		}
	}

	dev, err = netlink.LinkByName(name)
	if err != nil {
		return nil, fmt.Errorf("post-create failure on getting tap %q link: %v", name, err)
	}

	tap := &Tap{
		Name:       name,
		BridgePort: BridgePort{NetlinkDev: &dev},
	}

	if err := tap.SetUp(); err != nil {
		return nil, fmt.Errorf("post-create failure on setting tap %q link up: %v", name, err)
	}

	return tap, nil
}

// todo: add Remove method
