package ovsplug

import (
	"fmt"
	"os/exec"
)

// OVSInterface represents an interface record in OVSDB
type OVSInterface struct {
	Name     string
	resource *ExternalResource
}

// ExternalResource represents external resources of an ovsdb interface expected by neutron agent
type ExternalResource struct {
	IFName, Status, AttachedMAC, VMUUID string
}

// NewOVSInterface creates a named ovsdb interface record
func NewOVSInterface(name string) *OVSInterface {
	return &OVSInterface{Name: name}
}

func (r ExternalResource) toExternalIds() []string {
	result := make([]string, 0)
	if r.IFName != "" {
		result = append(result, fmt.Sprintf("external-ids:iface-id=%s", r.IFName))
	}
	if r.Status != "" {
		result = append(result, fmt.Sprintf("external-ids:iface-status=%s", r.Status))
	}
	if r.AttachedMAC != "" {
		result = append(result, fmt.Sprintf("external-ids:attached-mac=%s", r.AttachedMAC))
	}
	if r.VMUUID != "" {
		result = append(result, fmt.Sprintf("external-ids:vm-uuid=%s", r.VMUUID))
	}

	return result
}

// SetExternalResource sets the external resource of the interface and persists into ovsdb
func (oif *OVSInterface) SetExternalResource(name, status, mac, vm string) ([]byte, error) {
	if name == "" && status == "" && mac == "" && vm == "" {
		return nil, fmt.Errorf("invalid inputs, all empty")
	}

	oif.resource = &ExternalResource{
		IFName:      name,
		Status:      status,
		AttachedMAC: mac,
		VMUUID:      vm,
	}

	// go-openvswicth is short of set Interface external-ids support, hence we invent the wheel here
	// todo: either find a capable package or send PR to go-openvswicth adding such feature

	args := []string{"set", "Interface", oif.Name}
	args = append(args, oif.resource.toExternalIds()...)
	return exec.Command("ovs-vsctl", args...).CombinedOutput()
}
