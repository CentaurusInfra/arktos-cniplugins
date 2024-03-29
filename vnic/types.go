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
package vnic

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/containernetworking/cni/pkg/types"
)

const tenantDefault = "default"

// Args represents CNI stdin args specifically for neutron integration
type Args struct {
	types.CommonArgs
	Tenant types.UnmarshallableString
	VPC    types.UnmarshallableString
	NICs   types.UnmarshallableString
}

// VNIC contains individual vNic's port id and interface name
type VNIC struct {
	Name   string `json:"name,omitempty"`
	PortID string `json:"portid"`
}

// VNICs contains (the single) Tenant + VPC info and all nic data
type VNICs struct {
	Tenant string // correspending to openstack domain
	VPC    string // corresponding to openstack project ID
	NICs   []VNIC
}

// LoadVNICs extracts neutron related VNICs out of CNI args
func LoadVNICs(cniargs string) (*VNICs, error) {
	args := Args{CommonArgs: types.CommonArgs{IgnoreUnknown: true}}
	if err := types.LoadArgs(cniargs, &args); err != nil {
		return nil, fmt.Errorf("cannot load cni args %q: %s", cniargs, err)
	}

	if args.VPC == "" {
		return nil, fmt.Errorf("cannot load cni args %q: empty VPC", cniargs)
	}

	if strings.TrimSpace(string(args.Tenant)) == "" {
		args.Tenant = tenantDefault
	}

	nics := make([]VNIC, 0)
	if err := json.Unmarshal([]byte(args.NICs), &nics); err != nil {
		return nil, fmt.Errorf("cannot unmarshal nics text %q: %s", args.NICs, err)
	}

	if len(nics) == 0 {
		return nil, errors.New("empty nics definition")
	}

	for i, nic := range nics {
		if nic.PortID == "" {
			return nil, fmt.Errorf("invlid nic definition at index %d in nics %s: empty portid", i, args.NICs)
		}

		if nic.Name == "" {
			nics[i].Name = "eth" + strconv.Itoa(i)
		}
	}

	// todo: add more to validate nic/portid

	return &VNICs{
		Tenant: string(args.Tenant),
		VPC:    string(args.VPC),
		NICs:   nics}, nil
}

// EPnic represents the nic endpoint pluged in netns which corresponds to a vnic
type EPnic struct {
	Name    string
	MAC     string
	IPv4Net *net.IPNet
	Gw      *net.IP
}
