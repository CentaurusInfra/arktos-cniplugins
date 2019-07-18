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

// NewVeth creates a new veth pair having specific endpoint names
func NewVeth(name, peerName string) (*Veth, error) {
	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{Name: name},
		PeerName:  peerName,
	}

	if err := netlink.LinkAdd(veth); err != nil {
		return nil, fmt.Errorf("failed to create veth pair (%q, %q): %v", name, peerName, err)
	}

	ep := getVethEP(name)
	peer := getVethEP(peerName)
	if ep == nil || peer == nil {
		if ep != nil {
			netlink.LinkDel(*ep.BridgePort.NetlinkDev)
		} else if peer != nil {
			netlink.LinkDel(*peer.BridgePort.NetlinkDev)
		}
		return nil, fmt.Errorf("post-create failure on creating veth pair (%q, %q): unable to retrieve endpoints", name, peerName)
	}

	return &Veth{EP: ep, PeerEP: peer}, nil
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
