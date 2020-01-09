/*
Copyright 2019 The Alkaid Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
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

// NewLinuxBridge creates a a local Linux bridge resource struct
// The unerlying network device is created by InitDevice method
func NewLinuxBridge(name string) *LinuxBridge {
	return &LinuxBridge{
		Name: name,
	}
}

// InitDevice initializes underlying device of the local Linux bridge, ensures in up state
func (br *LinuxBridge) InitDevice() error {
	dev, err := netlink.LinkByName(br.Name)
	if err != nil {
		if _, ok := err.(netlink.LinkNotFoundError); !ok {
			return fmt.Errorf("failed to create bridge %q, cannot check link: %v", br.Name, err)
		}

		// named link not found; let's create the bridge
		if dev, err = createBridgeLink(br.Name); err != nil {
			return fmt.Errorf("failed to create bridge %q: %v", br.Name, err)
		}
	}

	brDesc, err := createBridgeDesc(br.Name, dev)
	if err != nil {
		return fmt.Errorf("failed to create description of bridge %q: %v", br.Name, err)
	}

	br.linkDev = &dev
	br.bridge = brDesc

	if err := br.SetUp(); err != nil {
		return fmt.Errorf("failed to set bridge %q up: %v", br.Name, err)
	}

	return nil
}

func createBridgeLink(name string) (netlink.Link, error) {
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

	return dev, nil
}

func createBridgeDesc(name string, link netlink.Link) (*netlink.Bridge, error) {
	if link.Type() != "bridge" {
		return nil, fmt.Errorf("name conflicting: %q had been used by link type %s", name, link.Type())
	}

	la := netlink.NewLinkAttrs()
	la.Name = name
	brDesc := &netlink.Bridge{LinkAttrs: la}

	return brDesc, nil
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

// Delete deletes the Linux bridge (with its underlying network device)
func (br *LinuxBridge) Delete() error {
	if err := deleteDev(br.Name); err != nil {
		return fmt.Errorf("failed to delete bridge %q: %v", br.Name, err)
	}

	return nil
}

// DeletePort deletes port from local Linux bridge by deleting the associated network device
// If the port device is in veth pair, whole veth pair would be deleted too
func (br *LinuxBridge) DeletePort(port string) error {
	if err := deleteDev(port); err != nil {
		return fmt.Errorf("failed to delete port %q from bridge %q: %v", port, br.Name, err)
	}

	return nil
}

func deleteDev(devName string) error {
	dev, err := netlink.LinkByName(devName)
	if err != nil {
		if _, ok := err.(netlink.LinkNotFoundError); !ok {
			return fmt.Errorf("could not locate dev %q; %v", devName, err)
		}

		// named link not found; it is OK
		return nil
	}

	return netlink.LinkDel(dev)
}
