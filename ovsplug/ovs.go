package ovsplug

import (
	"fmt"
	"os/exec"
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

// todo: add DelPort method
