package ovsplug

import (
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

// AddPort adds a port with specified name to the ovs bridge
func (b OVSBridge) AddPort(port string) error {
	c := ovs.New()
	return c.VSwitch.AddPort(b.Name, port)
}

// GetName gets the ovs bridge name
func (b OVSBridge) GetName() string {
	return b.Name
}

// todo: add DelPort method
