/*
Copyright 2019 The Arktos Authors.

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
