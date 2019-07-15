package vnic

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/containernetworking/cni/pkg/types"
)

// Args represents CNI stdin args specifically for neutron integration
type Args struct {
	types.CommonArgs
	VPC  types.UnmarshallableString
	NICs types.UnmarshallableString
}

// VNIC contains individual vNic's port id and interface name
type VNIC struct {
	NIC    string `json:"nic,omitempty"`
	PortID string `json:"portid"`
}

// VNICs contains (the single) VPC info and all nic data
type VNICs struct {
	VPC  string
	NICs []VNIC
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

		if nic.NIC == "" {
			nics[i].NIC = "eth" + strconv.Itoa(i)
		}
	}

	// todo: add more to validate nic/portid

	return &VNICs{VPC: string(args.VPC), NICs: nics}, nil
}
