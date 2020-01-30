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
	"os/exec"

	"github.com/digitalocean/go-openvswitch/ovs"
)

// OVSBridge represents a local ovs bridge
type OVSBridge struct {
	Name string
}

// NewOVSBridge creates an ovs bridge
func NewOVSBridge(name string) *OVSBridge {
	return &OVSBridge{Name: name}
}

// AddPortAndSetExtResources adds the port of specified name to the ovs bridge,
// and sets various external reources, in atomic fashion (otherwise neutron agent
// may get partial update and unable to populate flow table properly)
func (b OVSBridge) AddPortAndSetExtResources(name, portID, status, mac, vm string) ([]byte, error) {
	if portID == "" && status == "" && mac == "" && vm == "" {
		return nil, fmt.Errorf("invalid inputs, all empty")
	}

	resource := &ExternalResource{
		IFID:        portID,
		Status:      status,
		AttachedMAC: mac,
		VMUUID:      vm,
	}

	// todo: add ovs timeout setting to avoid hard code
	args := []string{"--timeout=120", "--", "--may-exist", "add-port", b.Name, name, "--", "set", "Interface", name}
	args = append(args, resource.toExternalIds()...)
	return exec.Command("ovs-vsctl", args...).CombinedOutput()
}

// GetName gets the ovs bridge name
func (b OVSBridge) GetName() string {
	return b.Name
}

// DeletePort deletes port from ovs bridge (and also implicitly removes flow table entries)
func (b OVSBridge) DeletePort(name string) error {
	return ovs.New().VSwitch.DeletePort(b.Name, name)
}
