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
package vnicplug

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/futurewei-cloud/cniplugins/alktron/neutron"
	"github.com/futurewei-cloud/cniplugins/alktron/nsvtep"
	"github.com/futurewei-cloud/cniplugins/alktron/ovsplug"
	"github.com/futurewei-cloud/cniplugins/vnic"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
	log "github.com/sirupsen/logrus"
	"github.com/uber-go/multierr"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	defaultProbeInterval = time.Millisecond * 500
	defaultProbeTimeout  = time.Second * 15
)

// PortGetBinder is the interface able to get and bind the neutron port
type PortGetBinder interface {
	GetPort(portID string) (*neutron.PortBindingDetail, error)
	BindPort(portID, hostID, devID string) (*neutron.PortBindingDetail, error)
	UnbindPort(portID string) (*neutron.PortBindingDetail, error)
}

// SubnetGetter is the interface able to get neutron subnet detail
type SubnetGetter interface {
	GetSubnet(subnetID string) (*subnets.Subnet, error)
}

// LocalHostPlugger is the interface which construct local ovs hybrid plug
type LocalHostPlugger interface {
	Plug() error
}

// DevNetnsManager is the interface which manages (attach/detach) endpoint device
// inside specified netns connecting to the local host bridge
type DevNetnsManager interface {
	Attach(dev string, mac net.HardwareAddr, ipnet *net.IPNet, gw *net.IP, prio int, hostBr string) error
	Detach(dev string, hostBr string) error
}

// Plugger represents the vnic plugger which does all the steps to turn vnic into
// a set of veth pairs, linux bridge, and applicable ipset rules
type Plugger struct {
	PortGetBinder   PortGetBinder
	SubnetGetter    SubnetGetter
	HybridPlugGen   func(portID string) ovsplug.LocalPlugger
	DevNetnsPlugger DevNetnsManager

	probeInterval time.Duration
	probeTimeout  time.Duration
}

// NewPlugger creates the Plugger applicable to Neutron ML2 ovs_hybrid_plug env
func NewPlugger(neutronClient *neutron.Client, nspath string) *Plugger {
	return &Plugger{
		PortGetBinder:   neutronClient,
		SubnetGetter:    neutronClient,
		HybridPlugGen:   ovsplug.NewHybridPlug,
		DevNetnsPlugger: nsvtep.NewManager(nspath),
		probeInterval:   defaultProbeInterval,
		probeTimeout:    defaultProbeTimeout,
	}
}

// SetProbeInterval sets the port status probe configuration of interval time
func (p Plugger) SetProbeInterval(interval time.Duration) {
	p.probeInterval = interval
}

// SetProbeTimeout sets the port status probe configuration of timeout value
func (p Plugger) SetProbeTimeout(timeout time.Duration) {
	p.probeTimeout = timeout
}

// Plug plugs vnic and makes the endpoint present in the target netns
func (p Plugger) Plug(vn *vnic.VNIC, devID, boundHost string, routePrio int) (*vnic.EPnic, error) {
	var err error
	defer func() {
		if err != nil {
			p.Unplug(vn)
		}
	}()

	portID := vn.PortID
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

	ovshybridplug := p.HybridPlugGen(portID)

	if err = ovshybridplug.InitDevices(); err != nil {
		return nil, fmt.Errorf("failed to plug vnic on ovs hybrid creation: %v", err)
	}

	if err = ovshybridplug.Plug(mac.String(), devID); err != nil {
		return nil, fmt.Errorf("failed to plug vnic on ovs hybrid plug: %v", err)
	}

	// todo: more stringent process of FixedIPs array
	// assuming the default element (index 0) always exist after successful binding
	ipnet, gw, err := p.getIPNetAndGw(bindingDetail.FixedIPs[0].IPAddress, bindingDetail.FixedIPs[0].SubnetID)
	if err != nil {
		return nil, fmt.Errorf("failed to plug vnic on getting ipnet: %v", err)
	}

	// make the endpoint nic inside netns, and add it to qbr
	if err = p.DevNetnsPlugger.Attach(vn.Name, mac, ipnet, gw, routePrio, ovshybridplug.GetLocalBridge()); err != nil {
		return nil, err
	}

	if err = p.ensureStatusActive(portID); err != nil {
		return nil, err
	}

	return &vnic.EPnic{
		Name:    vn.Name,
		MAC:     mac.String(),
		IPv4Net: ipnet,
		Gw:      gw,
	}, nil
}

// Unplug cleans up network resources allocated for the vnic
func (p Plugger) Unplug(vnic *vnic.VNIC) error {
	portID := vnic.PortID

	// sequence of unplug:
	// 1- delete veth pair across netns with tap vtep at root ns
	qbr := ovsplug.GetBridgeName(portID)
	err1 := p.DevNetnsPlugger.Detach(vnic.Name, qbr)

	// 2- unplug vif
	ovshybridplug := p.HybridPlugGen(portID)
	err2 := ovshybridplug.Unplug()

	// 3- neutron port unbind
	out, err3 := p.PortGetBinder.UnbindPort(portID)
	if err3 != nil {
		log.Warnf("unbind port got error: %v, returned port detail: %v", err3, out)
	}

	return multierr.Combine(err1, err2, err3)
}

func (p Plugger) ensureStatusActive(portID string) error {
	return wait.PollImmediate(p.probeInterval, p.probeTimeout, func() (bool, error) {
		portDetail, err := p.PortGetBinder.GetPort(portID)
		if err != nil {
			return false, fmt.Errorf("failed to plug vnic on verifying port status: %v", err)
		}

		return strings.EqualFold("active", portDetail.Status), nil
	})
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
