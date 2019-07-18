package ovsplug

import (
	"fmt"

	"github.com/vishvananda/netlink"
)

// LinuxBridge encapsulates brctl related ops
type LinuxBridge struct {
	Name    string
	bridge  *netlink.Bridge
	linkDev *netlink.Link
}

// IsCreated checks whether the named bridge already exists
func (br *LinuxBridge) IsCreated() (bool, error) {
	dev, err := netlink.LinkByName(br.Name)
	if err != nil {
		if _, ok := err.(netlink.LinkNotFoundError); ok {
			return false, nil
		}

		return false, err
	}

	if br.bridge == nil {
		if dev.Type() != "bridge" {
			return false, fmt.Errorf("name conflicting: %q had been used by link type %s", br.Name, dev.Type())
		}
		br.bridge = &netlink.Bridge{LinkAttrs: *dev.Attrs()}
	}

	return true, nil
}

// Create creates local bridge, like "brctl addbr foo"
func (br *LinuxBridge) Create() error {
	la := netlink.NewLinkAttrs()
	la.Name = br.Name
	newBr := &netlink.Bridge{LinkAttrs: la}
	if err := netlink.LinkAdd(newBr); err != nil {
		return fmt.Errorf("failed to create br %q: %v", br.Name, err)
	}

	dev, err := netlink.LinkByName(br.Name)
	if err != nil {
		return fmt.Errorf("post-create failure on creating bridge %q: %v", br.Name, err)
	}

	br.bridge = newBr
	br.linkDev = &dev
	return nil
}

// SetUp enables the link device
func (br *LinuxBridge) SetUp() error {
	return netlink.LinkSetUp(*br.linkDev)
}

// todo: add Remove method
// todo: add SetDown method if needed
