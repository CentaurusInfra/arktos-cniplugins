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

// NewLinuxBridge creates a new (or retrieve an existent) local Linux bridge device, and ensures it is in up state
func NewLinuxBridge(name string) (*LinuxBridge, error) {
	dev, err := netlink.LinkByName(name)
	if err != nil {
		if _, ok := err.(netlink.LinkNotFoundError); !ok {
			return nil, fmt.Errorf("failed to create bridge %v, cannot check link: %v", name, err)
		}

		// named link not found; let's create the bridge
		return createBridge(name)
	}

	// named link exists; we'll take it if type is bridge, and ensure it is up
	return createBridgeFromDev(name, dev)
}

func createBridge(name string) (*LinuxBridge, error) {
	// todo: cleanup in case of error
	la := netlink.NewLinkAttrs()
	la.Name = name
	newBr := &netlink.Bridge{LinkAttrs: la}
	if err := netlink.LinkAdd(newBr); err != nil {
		return nil, fmt.Errorf("failed to create br %q: %v", name, err)
	}

	dev, err := netlink.LinkByName(name)
	if err != nil {
		return nil, fmt.Errorf("post-create failure on retrieving bridge %q: %v", name, err)
	}

	br, err := createBridgeByLinkAndSetUp(name, newBr, &dev)
	if err != nil {
		return nil, fmt.Errorf("post-create failure: %v", err)
	}

	return br, nil
}

func createBridgeFromDev(name string, link netlink.Link) (*LinuxBridge, error) {
	if link.Type() != "bridge" {
		return nil, fmt.Errorf("name conflicting: %q had been used by link type %s", name, link.Type())
	}

	br, err := createBridgeByLinkAndSetUp(name, &netlink.Bridge{LinkAttrs: *link.Attrs()}, &link)
	if err != nil {
		return nil, fmt.Errorf("failed to get bridge %q: %v", name, err)
	}

	return br, nil
}

func createBridgeByLinkAndSetUp(name string, brMeta *netlink.Bridge, link *netlink.Link) (*LinuxBridge, error) {
	br := &LinuxBridge{
		Name:    name,
		bridge:  brMeta,
		linkDev: link,
	}

	if err := br.SetUp(); err != nil {
		return nil, fmt.Errorf("failed to set bridge %q up: %v", name, err)
	}

	return br, nil

}

// SetUp enables the link device
func (br *LinuxBridge) SetUp() error {
	return netlink.LinkSetUp(*br.linkDev)
}

// AddPort adds a port with specified name to the linux bridge (as brctl addif does)
func (br *LinuxBridge) AddPort(port string) error {
	portDev, err := netlink.LinkByName(port)
	if err != nil {
		return fmt.Errorf("failed with retrieval of %q: %v", port, err)
	}

	if err := netlink.LinkSetMaster(portDev, br.bridge); err != nil {
		return fmt.Errorf("bridge %q failed to add %q: %v", br.Name, port, err)
	}

	return nil
}

// GetName gets the local Linux bridge name
func (br *LinuxBridge) GetName() string {
	return br.Name
}

// todo: add Remove method
// todo: add SetDown method
