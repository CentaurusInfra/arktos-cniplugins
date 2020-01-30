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
package main

import (
	"fmt"

	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/futurewei-cloud/cniplugins/vnic"
	log "github.com/sirupsen/logrus"
	"github.com/uber-go/multierr"
)

type plugger interface {
	Plug(vnic *vnic.VNIC) (*vnic.EPnic, error)
}

type unplugger interface {
	Unplug(vnic *vnic.VNIC) error
}

func attachVNICs(plugger plugger, vns []vnic.VNIC, sandbox string) (*current.Result, error) {
	r := &current.Result{}

	for i, vn := range vns {
		nic, err := plugger.Plug(&vn)
		if err != nil {
			if errCleanup := detachVNICs(plugger.(unplugger), vns[:i]); errCleanup != nil {
				log.Warnf("attach vnics aborted; cleanup had error: %v", errCleanup)
			}
			return nil, fmt.Errorf("failed to plug vnic %v: %v", vn, err)
		}

		intf := &current.Interface{
			Name:    nic.Name,
			Mac:     nic.MAC,
			Sandbox: sandbox,
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
	}

	return r, nil
}

func detachVNICs(unplugger unplugger, vns []vnic.VNIC) error {
	var combinedErrors error

	for _, vn := range vns {
		if err := unplugger.Unplug(&vn); err != nil {
			combinedErrors = multierr.Combine(combinedErrors, err)
		}
	}

	return combinedErrors
}
