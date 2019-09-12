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

// todo - replace EPnic defined in alktron package w/ this applicable to both alktron/mizni

// EPnic represents the physical endpoint NIC pluged in netns
// corresponding to a vnic
type EPnic struct {
	Name    string
	MAC     string
	IPv4Net *net.IPNet
	Gw      *net.IP
}
