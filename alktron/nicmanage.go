package main

import (
	"fmt"

	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/futurewei-cloud/alktron/alktron/vnicplug"
	"github.com/futurewei-cloud/alktron/vnic"
)

// route metric starts from 100, decreasing subsequentially
const initialRoutePrio = 100

// Plugger is the plugger oversees the whole process of attaching/detaching vnic(neutron port)
type Plugger interface {
	Plug(vnic *vnic.VNIC, devID, boundHost string, routePrio int) (*vnicplug.EPnic, error)
	Unplug(vnic *vnic.VNIC) error
}

func attachVNICs(plugger Plugger, vns []vnic.VNIC, devID, host string) (*current.Result, error) {
	// todo: consider proper cleanup in case of error with multi nics
	routePrio := initialRoutePrio

	r := &current.Result{}

	for i, vn := range vns {
		nic, err := plugger.Plug(&vn, devID, host, routePrio)
		if err != nil {
			return nil, fmt.Errorf("failed to plug vnic %v: %v", vn, err)
		}

		intf := &current.Interface{
			Name:    nic.Name,
			Mac:     nic.MAC,
			Sandbox: devID,
		}
		r.Interfaces = append(r.Interfaces, intf)

		ip := &current.IPConfig{
			Version:   "4", // we only care about ipv4 for now
			Interface: &i,
			Address:   *nic.IPv4Net,
			Gateway:   *nic.Gw,
		}
		r.IPs = append(r.IPs, ip)

		if routePrio > 1 {
			routePrio--
		}
	}

	return r, nil
}

func detachVNICs(plugger Plugger, vns []vnic.VNIC) error {
	for _, vn := range vns {
		if err := plugger.Unplug(&vn); err != nil {
			return fmt.Errorf("failed to unplug vnic %v: %v", vn, err)
		}
	}

	return nil
}
