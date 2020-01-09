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
package main

import (
	"fmt"

	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/futurewei-cloud/cniplugins/vnic"
)

// route metric starts from 100, decreasing subsequentially
const initialRoutePrio = 100

// Plugger is the plugger oversees the whole process of attaching/detaching vnic(neutron port)
type Plugger interface {
	Plug(vnic *vnic.VNIC, devID, boundHost string, routePrio int) (*vnic.EPnic, error)
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

		nicIndex := i

		ip := &current.IPConfig{
			Version:   "4", // we only care about ipv4 for now
			Interface: &nicIndex,
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
