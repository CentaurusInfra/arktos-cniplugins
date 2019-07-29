package ovsplug

import "fmt"

const ovsbridge = "br-int"

// HybridPlug represents the typical ovs hybrid plug structure
// seen as qbr <--> qvb <--> qvo <--> br-int
type HybridPlug struct {
	NeutronPortID string
	MACAddr       string
	VMID          string

	OVSBridge    Bridge
	OVSInterface ExtResourceSetter
	LinuxBridge  Bridge
	Qvo, Qvb     NamedDevice

	// tap not created here as it should be in the desired CNI netns
	// todo: add tap related data need for tap creation
}

// LocalPlugger is the interface which construct local ovs hybrid plug
type LocalPlugger interface {
	Plug() error
	GetLocalBridge() string
}

// Bridge is the interface of device with ports attached
type Bridge interface {
	NamedDevice
	AddPort(port string) error
}

// ExtResourceSetter is the interface to set external resource
type ExtResourceSetter interface {
	SetExternalResource(name, status, mac, vm string) ([]byte, error)
}

// NamedDevice is the interface that has a name
type NamedDevice interface {
	GetName() string
}

// NewHybridPlug creates an ovs hybrid plug for the neutron port
func NewHybridPlug(portID, mac, vm string) (LocalPlugger, error) {
	// Openstack convention to pick the first 11 chars of port id
	portPrefix := portID
	if len(portID) > 11 {
		portPrefix = portID[:11]
	}

	lbr, err := NewLinuxBridge("qbr" + portPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to create ovs hybrid plug for port id %q: %v", portID, err)
	}

	veth, err := NewVeth("qvb"+portPrefix, "qvo"+portPrefix)
	if err != nil {
		// todo: cleanup linux bridge
		return nil, fmt.Errorf("failed to create ovs hybrid plug for port id %q: %v", portID, err)
	}

	return &HybridPlug{
		NeutronPortID: portID,
		MACAddr:       mac,
		VMID:          vm,
		OVSBridge:     NewOVSBridge(ovsbridge),
		OVSInterface:  NewOVSInterface("qvo" + portPrefix),
		LinuxBridge:   lbr,
		Qvb:           veth.EP,
		Qvo:           veth.PeerEP,
	}, nil
}

// Plug creates needed devices and connects them properly
func (h HybridPlug) Plug() error {
	h.OVSBridge.AddPort(h.Qvo.GetName())

	_, err := h.OVSInterface.SetExternalResource(h.NeutronPortID, "active", h.MACAddr, h.VMID)
	if err != nil {
		return fmt.Errorf("plug failed on setting external-ids: %v", err)
	}

	if err = h.LinuxBridge.AddPort(h.Qvb.GetName()); err != nil {
		return fmt.Errorf("plug failed on adding qvb to qbr: %v", err)
	}

	return nil
}

// GetLocalBridge gets the local Linuxbridge name
func (h HybridPlug) GetLocalBridge() string {
	return h.LinuxBridge.GetName()
}
