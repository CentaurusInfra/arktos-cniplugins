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

// Veth represents a veth pair
type Veth struct {
	EP     *VethEP
	PeerEP *VethEP
}

// NewVeth creates a new veth pair having specific endpoint names, ensures its pairs in up state
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
