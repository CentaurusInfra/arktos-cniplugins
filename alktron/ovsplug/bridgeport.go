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
	"net"

	"github.com/vishvananda/netlink"
)

// BridgePort represents the base of network devices able to join a local bridge
type BridgePort struct {
	NetlinkDev *netlink.Link
}

// SetUp enables the link device
func (bp *BridgePort) SetUp() error {
	return netlink.LinkSetUp(*bp.NetlinkDev)
}

// ChangeMAC changes MAC address
func (bp *BridgePort) ChangeMAC(mac net.HardwareAddr) error {
	if err := netlink.LinkSetHardwareAddr(*bp.NetlinkDev, mac); err != nil {
		return fmt.Errorf("failed to change MAC address: %v", err)
	}

	devName := (*bp.NetlinkDev).Attrs().Name
	dev, err := netlink.LinkByName(devName)
	if err != nil {
		return fmt.Errorf("unexpected error, dev %q is corrupted: %v", devName, err)
	}

	bp.NetlinkDev = &dev
	return nil
}
