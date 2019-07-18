package ovsplug

import (
	"fmt"

	"github.com/vishvananda/netlink"
)

// VethEP respresents one endpoint of veth pair
type VethEP struct {
	BridgePort
	Name string
}

// Veth represents a veth pair
type Veth struct {
	EP     *VethEP
	PeerEP *VethEP
}

// Create creates veth pair having specific names
func (v *Veth) Create(name, peerName string) error {
	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{Name: name},
		PeerName:  peerName,
	}

	if err := netlink.LinkAdd(veth); err != nil {
		return fmt.Errorf("failed to create veth pair (%q, %q): %v", name, peerName, err)
	}

	v.EP = getVethEP(name)
	v.PeerEP = getVethEP(peerName)
	if v.EP == nil && v.PeerEP == nil {
		if v.EP != nil {
			netlink.LinkDel(*v.EP.BridgePort.NetlinkDev)
		} else if v.PeerEP != nil {
			netlink.LinkDel(*v.PeerEP.BridgePort.NetlinkDev)
		}
		return fmt.Errorf("post-create failure on creating veth pair (%q, %q): unable to retrieve endpoints", name, peerName)
	}

	return nil
}

// todo: add Remove method

func getVethEP(name string) *VethEP {
	dev, err := netlink.LinkByName(name)
	if err != nil {
		return nil
	}

	// todo: add more stringent check against link type etc

	return &VethEP{
		Name:       name,
		BridgePort: BridgePort{NetlinkDev: &dev},
	}
}
