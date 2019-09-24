package vnicmanager

import (
	"fmt"
	"net"

	"github.com/futurewei-cloud/cniplugins/vnic"
)

type devProber interface {
	DeviceReady(name, nsPath string) error
}

type devNetConfGetter interface {
	GetDevNetConf(name, nsPath string) (*net.IPNet, *net.IP, int, string, int, error)
}

type nsMigrator interface {
	Migrate(nameFrom, nsFrom, nameTo, nsTo string, ipnet *net.IPNet, gw *net.IP, metric, mtu int) error
}

// Manager represents the object in charge of plug single vnic
type Manager struct {
	VPC        string
	NScni      string
	DevProber  devProber
	ConfGetter devNetConfGetter
	NSMigrator nsMigrator
}

// New creates the VNIC Manager
func New(vpc, cniNS string) *Manager {
	nsdevManager := &nsdev{}

	return &Manager{
		VPC:        vpc,
		NScni:      cniNS,
		DevProber:  &nicProberWithTimeout{}, //todo: make timeout configurable based on cni netconf file
		ConfGetter: nsdevManager,
		NSMigrator: nsdevManager,
	}
}

// Plug plugs vnic
func (m Manager) Plug(vn *vnic.VNIC) (*vnic.EPnic, error) {
	// todo: proper cleanup on error

	alcorNSPath := getAlcorNSPath(m.VPC)
	dev := getDevName(vn.PortID)

	if err := m.DevProber.DeviceReady(dev, alcorNSPath); err != nil {
		return nil, fmt.Errorf("Plug vnic %q failed, dev %q not ready: %v", vn.PortID, dev, err)
	}

	ipNet, gw, metric, mac, mtu, err := m.ConfGetter.GetDevNetConf(dev, alcorNSPath)
	if err != nil {
		return nil, fmt.Errorf("Plug vnic %q failed, unable to get settings: %v", vn.PortID, err)
	}

	if err := m.NSMigrator.Migrate(dev, alcorNSPath, vn.Name, m.NScni, ipNet, gw, metric, mtu); err != nil {
		return nil, fmt.Errorf("Plug vnic %q failed, unable to migrate to cni-ns: %v", vn.PortID, err)
	}

	return &vnic.EPnic{
		Name:    vn.Name,
		MAC:     mac,
		IPv4Net: ipNet,
		Gw:      gw,
	}, nil
}

// Unplug unplugs vnic
func (m Manager) Unplug(vn *vnic.VNIC) error {
	ipNet, gw, metric, _, mtu, err := m.ConfGetter.GetDevNetConf(vn.Name, m.NScni)
	if err != nil {
		return fmt.Errorf("Unplug vnic %q failed, unable to get settings: %v", vn.PortID, err)
	}

	alcorNSPath := getAlcorNSPath(m.VPC)
	alcorDev := getDevName(vn.PortID)
	if err := m.NSMigrator.Migrate(vn.Name, m.NScni, alcorDev, alcorNSPath, ipNet, gw, metric, mtu); err != nil {
		return fmt.Errorf("Unplug vnic %q failed, unable to migrate to alcor-ns: %v", vn.PortID, err)
	}

	return nil
}

func getAlcorNSPath(vpc string) string {
	// alcor agreement of alcor ns name is vpc-ns{full-vpc-name}
	// the full nspath is /run/netns/ + alcor-ns-name
	const prefix = "/run/netns/vpc-ns"
	return prefix + vpc
}

func getDevName(portID string) string {
	// alcor agreement of device name composition: veth+{portid-11-char-prefix}
	const prefix = "veth"
	const lenPortIDPrefix = 11

	portPrefix := portID
	if len(portID) > lenPortIDPrefix {
		portPrefix = portID[:lenPortIDPrefix]
	}

	return prefix + portPrefix
}
