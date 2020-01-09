/*
Copyright 2019 The Alkaid Authors.

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
package vnicmanager

import (
	"fmt"
	"net"
	"time"

	"github.com/futurewei-cloud/cniplugins/vnic"
	log "github.com/sirupsen/logrus"
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
func New(vpc, cniNS string, probeTimeout time.Duration) *Manager {
	nsdevManager := &nsdev{}

	return &Manager{
		VPC:        vpc,
		NScni:      cniNS,
		DevProber:  &nicProberWithTimeout{timeout: probeTimeout},
		ConfGetter: nsdevManager,
		NSMigrator: nsdevManager,
	}
}

// Plug plugs vnic
func (m Manager) Plug(vn *vnic.VNIC) (*vnic.EPnic, error) {
	alcorNSPath := getAlcorNSPath(m.VPC)
	alcorDev := getDevName(vn.PortID)

	if err := m.DevProber.DeviceReady(alcorDev, alcorNSPath); err != nil {
		return nil, fmt.Errorf("Plug vnic %q failed, dev %q not ready: %v", vn.PortID, alcorDev, err)
	}

	epnic, err := m.plug(alcorDev, alcorNSPath, vn.Name, m.NScni)
	if err != nil {
		if errCleanup := m.Unplug(vn); errCleanup != nil {
			log.Warnf("Plug vnic %q failed; cleanup had error: %v", vn.PortID, errCleanup)
		}
	}
	return epnic, err
}

func (m Manager) plug(devFrom, nsPathFrom, devTo, nsPathTo string) (*vnic.EPnic, error) {
	ipNet, gw, metric, mac, mtu, err := m.ConfGetter.GetDevNetConf(devFrom, nsPathFrom)
	if err != nil {
		return nil, fmt.Errorf("Plug dev %q failed; unable to get settings: %v", devFrom, err)
	}

	if err := m.NSMigrator.Migrate(devFrom, nsPathFrom, devTo, nsPathTo, ipNet, gw, metric, mtu); err != nil {
		return nil, fmt.Errorf("Plug dev %q failed, unable to migrate to cni-ns: %v", devFrom, err)
	}

	return &vnic.EPnic{
		Name:    devTo,
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
