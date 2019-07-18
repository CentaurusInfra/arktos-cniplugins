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

// NewLinuxBridge creates a new (or retrieve an existent) local Linux bridge device
func NewLinuxBridge(name string) (*LinuxBridge, error) {
	dev, err := netlink.LinkByName(name)
	if err != nil {
		if _, ok := err.(netlink.LinkNotFoundError); !ok {
			return nil, fmt.Errorf("failed to create bridge %v, cannot check link: %v", name, err)
		}

		// named link not found; let's create the bridge
		return createBridge(name)
	}

	// named link exists; we'll take it if type is bridge
	if dev.Type() != "bridge" {
		return nil, fmt.Errorf("name conflicting: %q had been used by link type %s", name, dev.Type())
	}

	return &LinuxBridge{
		Name:    name,
		bridge:  &netlink.Bridge{LinkAttrs: *dev.Attrs()},
		linkDev: &dev,
	}, nil
}

func createBridge(name string) (*LinuxBridge, error) {
	la := netlink.NewLinkAttrs()
	la.Name = name
	newBr := &netlink.Bridge{LinkAttrs: la}
	if err := netlink.LinkAdd(newBr); err != nil {
		return nil, fmt.Errorf("failed to create br %q: %v", name, err)
	}

	dev, err := netlink.LinkByName(name)
	if err != nil {
		return nil, fmt.Errorf("post-create failure on creating bridge %q: %v", name, err)
	}

	return &LinuxBridge{
		Name:    name,
		bridge:  newBr,
		linkDev: &dev,
	}, nil
}

// SetUp enables the link device
func (br *LinuxBridge) SetUp() error {
	return netlink.LinkSetUp(*br.linkDev)
}

// todo: add Remove method
// todo: add SetDown method
