package ovsplug

import (
	"fmt"

	"github.com/uber-go/multierr"
)

const ovsbridge = "br-int"

// HybridPlug represents the typical ovs hybrid plug structure
// seen as qbr <--> qvb <--> qvo <--> br-int
type HybridPlug struct {
	NeutronPortID string

	// network device resources
	OVSBridge   ExtResBridge
	LinuxBridge Bridge
	Qvo, Qvb    string
}

// LocalPlugger is the interface which construct local ovs hybrid plug
type LocalPlugger interface {
	InitDevices() error
	Plug(mac, vm string) error
	Unplug() error
	GetLocalBridge() string
}

// Bridge is the interface of device with ports attached
type Bridge interface {
	NamedDevice
	InitDevice() error
	AddPort(port string) error
	DeletePort(port string) error
	Delete() error
}

// ExtResBridge is the interface of device with ports along with external resource seeting
// Adding port and setting properties is required in one transaction
type ExtResBridge interface {
	NamedDevice
	AddPortAndSetExtResources(name, portID, status, mac, vm string) ([]byte, error)
	DeletePort(name string) error
}

// NamedDevice is the interface that has a name
type NamedDevice interface {
	GetName() string
}

// NewHybridPlug creates an ovs hybrid plug for the neutron port
// Only informational data is populated in this new func;
// the underlying network devices will be created in separate method, InitDevices
func NewHybridPlug(portID string) LocalPlugger {
	brName := GetBridgeName(portID)
	qvb, qvo := getVTEPName(portID)

	return &HybridPlug{
		NeutronPortID: portID,
		OVSBridge:     NewOVSBridge(ovsbridge),
		LinuxBridge:   NewLinuxBridge(brName),
		Qvo:           qvo,
		Qvb:           qvb,
	}
}

// InitDevices creates underlying network devices (qbr, qvb/qvo veth pair) of hybrid plug
func (h *HybridPlug) InitDevices() error {
	if err := h.LinuxBridge.InitDevice(); err != nil {
		return fmt.Errorf("failed to create ovs hybrid plug for port id %q: %v", h.NeutronPortID, err)
	}

	qvb, qvo := getVTEPName(h.NeutronPortID)
	_, err := NewVeth(qvb, qvo)
	if err != nil {
		// todo: cleanup linux bridge
		return fmt.Errorf("failed to create ovs hybrid plug for port id %q: %v", h.NeutronPortID, err)
	}

	return nil
}

// Plug creates needed devices and connects them properly
func (h *HybridPlug) Plug(mac, vm string) error {
	out, err := h.OVSBridge.AddPortAndSetExtResources(h.Qvo, h.NeutronPortID, "active", mac, vm)
	if err != nil {
		return fmt.Errorf("plug failed on setting external-ids, %s: %v", string(out), err)
	}

	if err = h.LinuxBridge.AddPort(h.Qvb); err != nil {
		return fmt.Errorf("plug failed on adding qvb to qbr: %v", err)
	}

	return nil
}

// GetLocalBridge gets the local Linuxbridge name
func (h *HybridPlug) GetLocalBridge() string {
	return h.LinuxBridge.GetName()
}

// Unplug cleans up network devices allocated for this hybrid structure
func (h *HybridPlug) Unplug() error {
	return multierr.Combine(
		h.OVSBridge.DeletePort(h.Qvo),
		h.LinuxBridge.DeletePort(h.Qvb),
		h.LinuxBridge.Delete(),
	)
}

// GetBridgeName gets the linux bridge name qbrxxxx-xx based on openstack convention
func GetBridgeName(portID string) string {
	return "qbr" + getPortPrefix(portID)
}

func getVTEPName(portID string) (string, string) {
	portPrefix := getPortPrefix(portID)
	return "qvb" + portPrefix, "qvo" + portPrefix
}

func getPortPrefix(portID string) string {
	// Openstack convention to pick the first 11 chars of port id
	// see https://github.com/openstack/nova/blob/4e9d2244799fb285f75056f9120201aaa408a765/nova/network/os_vif_util.py#L57
	// & https://github.com/openstack/nova/blob/4e9d2244799fb285f75056f9120201aaa408a765/nova/network/model.py#L166
	const lenPortIDPrefix = 11
	portPrefix := portID
	if len(portID) > lenPortIDPrefix {
		portPrefix = portID[:lenPortIDPrefix]
	}

	return portPrefix
}
