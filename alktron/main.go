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
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/futurewei-cloud/cniplugins/alktron/vnicplug"
	"github.com/futurewei-cloud/cniplugins/vnic"
)

func cmdAdd(args *skel.CmdArgs) error {
	// to validate and parse vpc, portid from args.Args
	vnics, err := vnic.LoadVNICs(args.Args)
	if err != nil {
		return fmt.Errorf("ADD op failed to load cni args: %v", err)
	}

	cniVersion, err := getCNIVerInNetConf(args.StdinData)
	if err != nil {
		return fmt.Errorf("ADD op failed to load netconf: %v", err)
	}

	netns, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	}
	defer netns.Close()

	nc, err := loadNeutronConfig()
	if err != nil {
		return fmt.Errorf("failed to load neutron config: %v", err)
	}

	neutronClient, err := nc.getNeutronClient(vnics.Tenant, vnics.VPC)
	if err != nil {
		return fmt.Errorf("failed to get neutron client: %v", err)
	}

	hostBound := nc.Host
	if hostBound == "" {
		// todo: use localhost as default
		return fmt.Errorf("invalid config: Host not specified")
	}

	plugger := vnicplug.NewPlugger(neutronClient, args.Netns)

	if nc.ProbeIntervalInMilliseconds != 0 {
		plugger.SetProbeInterval(time.Millisecond * time.Duration(nc.ProbeIntervalInMilliseconds))
	}

	if nc.ProbeTimeoutInSeconds != 0 {
		plugger.SetProbeTimeout(time.Second * time.Duration(nc.ProbeTimeoutInSeconds))
	}

	// todo: consider a better devID (like pod id?) than args.ContainerID
	r, err := attachVNICs(plugger, vnics.NICs, args.ContainerID, hostBound)
	if err != nil {
		return fmt.Errorf("ADD op failed to attach vnics: %v", err)
	}

	versionedResult, err := r.GetAsVersion(cniVersion)
	return versionedResult.Print()
}

func cmdDel(args *skel.CmdArgs) error {
	vnics, err := vnic.LoadVNICs(args.Args)
	if err != nil {
		return fmt.Errorf("DEL op failed to load cni args: %v", err)
	}

	netns, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	}
	defer netns.Close()

	nc, err := loadNeutronConfig()
	if err != nil {
		return fmt.Errorf("failed to load neutron config: %v", err)
	}

	neutronClient, err := nc.getNeutronClient(vnics.Tenant, vnics.VPC)
	if err != nil {
		return fmt.Errorf("failed to get neutron client: %v", err)
	}

	plugger := vnicplug.NewPlugger(neutronClient, args.Netns)
	return detachVNICs(plugger, vnics.NICs)
}

func cmdCheck(args *skel.CmdArgs) error {
	return errors.New("to be implemented")
}

func main() {
	supportVersions := version.PluginSupports("0.1.0", "0.2.0", "0.3.0", "0.3.1")
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, supportVersions, "alktron")
}

func getCNIVerInNetConf(bytes []byte) (string, error) {
	n := &types.NetConf{}
	if err := json.Unmarshal(bytes, n); err != nil {
		return "", fmt.Errorf("failed to load netconf: %v", err)
	}
	return n.CNIVersion, nil
}
