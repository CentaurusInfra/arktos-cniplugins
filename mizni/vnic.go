package main

import (
	"fmt"

	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/futurewei-cloud/cniplugins/vnic"
)

type plugger interface {
	Plug(vnic *vnic.VNIC) (*vnic.EPnic, error)
}

func attachVNICs(plugger plugger, vns []vnic.VNIC, sandbox string) (*current.Result, error) {
	r := &current.Result{}

	for i, vn := range vns {
		nic, err := plugger.Plug(&vn)
		if err != nil {
			return nil, fmt.Errorf("failed to plug vnic %v: %v", vn, err)
		}

		intf := &current.Interface{
			Name:    nic.Name,
			Mac:     nic.MAC,
			Sandbox: sandbox,
		}
		r.Interfaces = append(r.Interfaces, intf)

		ip := &current.IPConfig{
			Version:   "4", // we only care about ipv4 for now
			Interface: &i,
			Address:   *nic.IPv4Net,
			Gateway:   *nic.Gw,
		}
		r.IPs = append(r.IPs, ip)
	}

	return r, nil
}
