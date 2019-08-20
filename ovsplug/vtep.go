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

// GetName gets name
func (ep VethEP) GetName() string {
	return ep.Name
}

// RemoveVEP removes vtep device (effectively w/ its peer)
func RemoveVEP(name string) error {
	if err := deleteDev(name); err != nil {
		return fmt.Errorf("failed to delete vtep %q: %v", name, err)
	}

	return nil
}

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
