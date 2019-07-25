package vnicplug

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/futurewei-cloud/alktron/neutron"
	"github.com/futurewei-cloud/alktron/nsvtep"
	"github.com/futurewei-cloud/alktron/ovsplug"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"

	"github.com/futurewei-cloud/alktron/vnic"
)

// EPnic represents the physical endpoint NIC pluuged in netns
type EPnic struct {
	Name    string
	MAC     string
	IPv4Net *net.IPNet
	Gw      *net.IP
}

// PortGetBinder is the interface able to get and bind the neutron port
type PortGetBinder interface {
	GetPort(portID string) (*neutron.PortBindingDetail, error)
	BindPort(portID, hostID, devID string) (*neutron.PortBindingDetail, error)
}

// SubnetGetter is the interface able to get neutron subnet detail
type SubnetGetter interface {
	GetSubnet(subnetID string) (*subnets.Subnet, error)
}

// LocalHostPlugger is the interface which construct local ovs hybrid plug
type LocalHostPlugger interface {
	Plug() error
}

// DevNetnsManager is the interface which manages (attaches) endpoint device
// inside specified netns connecting to the local host bridge
type DevNetnsManager interface {
	Attach(dev string, mac net.HardwareAddr, ipnet *net.IPNet, gw *net.IP, prio int, hostBr string) error
}

// Plugger represents the vnic plugger which does all the steps to turn vnic into
// a set of veth pairs, linux bridge, and applicable ipset rules
type Plugger struct {
	PortGetBinder   PortGetBinder
	SubnetGetter    SubnetGetter
	HybridPlugGen   func(portID, mac, vm string) (ovsplug.LocalPlugger, error)
	DevNetnsPlugger DevNetnsManager
}

// NewPlugger creates the Plugger applicable to Neutron ML2 ovs_hybrid_plug env
func NewPlugger(neutronClient *neutron.Client, nspath string) *Plugger {
	return &Plugger{
		PortGetBinder:   neutronClient,
		SubnetGetter:    neutronClient,
		HybridPlugGen:   ovsplug.NewHybridPlug,
		DevNetnsPlugger: nsvtep.NewManager(nspath),
	}
}

// Plug plugs vnic and makes the endpoint present in the target netns
func (p Plugger) Plug(vnic *vnic.VNIC, devID, boundHost string, routePrio int) (*EPnic, error) {
	// todo: add proper cleanup code in case of error

	portID := vnic.PortID
	// todo: check port status to see if it is used already by other devID
	// todo: check port status to see if it is already ready for this devID

	bindingDetail, err := p.PortGetBinder.BindPort(portID, boundHost, devID)
	if err != nil {
		return nil, fmt.Errorf("failed to plug vnic on port binding: %v", err)
	}

	// ovs hybrid plug to construct qbr-qvb-qvo-brint
	mac, err := net.ParseMAC(bindingDetail.MACAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to plug vnic on port binding; port has invalid mac address: %v", err)
	}

	ovshybridplug, err := p.HybridPlugGen(portID, mac.String(), devID)
	if err != nil {
		return nil, fmt.Errorf("failed to plug vnic on ovs hybrid creation: %v", err)
	}
	if err = ovshybridplug.Plug(); err != nil {
		return nil, fmt.Errorf("failed to plug vnic on ovs hybrid plug: %v", err)
	}

	// todo: more stringent process of FixedIPs array
	// assuming the default element (index 0) always exist after successful binding
	ipnet, gw, err := p.getIPNetAndGw(bindingDetail.FixedIPs[0].IPAddress, bindingDetail.FixedIPs[0].SubnetID)
	if err != nil {
		return nil, fmt.Errorf("failed to plug vnic on getting ipnet: %v", err)
	}

	// make the endpoint nic inside netns, and add it to qbr
	if err = p.DevNetnsPlugger.Attach(vnic.Name, mac, ipnet, gw, routePrio, ovshybridplug.GetLocalBridge()); err != nil {
		return nil, err
	}

	if err = p.ensureStatusActive(portID); err != nil {
		return nil, err
	}

	return &EPnic{
		Name:    vnic.Name,
		MAC:     mac.String(),
		IPv4Net: ipnet,
		Gw:      gw,
	}, nil
}

func (p Plugger) ensureStatusActive(portID string) error {
	// todo: add time out to avoid live lock
	portDetail, err := p.PortGetBinder.GetPort(portID)
	for {
		if err != nil {
			return fmt.Errorf("failed to plug vnic on verifying port status: %v", err)
		}

		if strings.EqualFold("active", portDetail.Status) {
			break
		}

		// todo: add period time config to avoid hardcoded value
		<-time.After(1 * time.Second)
		portDetail, err = p.PortGetBinder.GetPort(portID)
	}

	return nil
}

func (p Plugger) getIPNetAndGw(ip, subnetID string) (*net.IPNet, *net.IP, error) {
	subnet, err := p.SubnetGetter.GetSubnet(subnetID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get subnet: %v", err)
	}
	cidr := subnet.CIDR
	_, ipv4Net, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid CIDR %q: %v", cidr, err)
	}

	ipGw := net.ParseIP(subnet.GatewayIP)

	return &net.IPNet{
		IP:   net.ParseIP(ip),
		Mask: ipv4Net.Mask,
	}, &ipGw, nil
}
