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

// NewVeth creates a new veth pair having specific endpoint names, ensures its pairs in up state
func NewVeth(name, peerName string) (*Veth, error) {
	// todo: cleanup faulty veth pair in case of error
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

	v := &Veth{EP: ep, PeerEP: peer}
	if err := v.SetUp(); err != nil {
		return nil, fmt.Errorf("post-create failure on creating veth pair (%q, %q): %v", name, peerName, err)
	}

	return v, nil
}

// SetUp ensures endpoints of veth pair in up status
func (v *Veth) SetUp() error {
	if err := v.EP.SetUp(); err != nil {
		return fmt.Errorf("failed to set veth ep %q up: %v", v.EP.Name, err)
	}

	if err := v.PeerEP.SetUp(); err != nil {
		return fmt.Errorf("failed to set veth ep %q up: %v", v.PeerEP.Name, err)
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

// GetName gets name
func (ep VethEP) GetName() string {
	return ep.Name
}
