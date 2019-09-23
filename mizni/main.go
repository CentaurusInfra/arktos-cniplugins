package main

import (
	"errors"
	"fmt"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/futurewei-cloud/cniplugins/mizni/vnicmanager"
	"github.com/futurewei-cloud/cniplugins/vnic"
)

func cmdAdd(args *skel.CmdArgs) error {
	vnics, err := vnic.LoadVNICs(args.Args)
	if err != nil {
		return fmt.Errorf("ADD op failed to load cni args: %v", err)
	}

	configDecoder := version.ConfigDecoder{}
	cniVersion, err := configDecoder.Decode(args.StdinData)
	if err != nil {
		return fmt.Errorf("ADD op failed to load netconf: %v", err)
	}

	netns, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	}
	defer netns.Close()

	plugger := vnicmanager.New(vnics.VPC, args.Netns)

	r, err := attachVNICs(plugger, vnics.NICs, args.ContainerID)
	if err != nil {
		return fmt.Errorf("ADD op failed to attach vnics: %v", err)
	}

	versionedResult, err := r.GetAsVersion(cniVersion)
	if err != nil {
		return fmt.Errorf("ADD op failed to convert result: %v", err)
	}

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

	// todo: stuff a concrete object of unplugger
	var unplugger unplugger
	return detachVNICs(unplugger, vnics.NICs)
}

func cmdCheck(args *skel.CmdArgs) error {
	return errors.New("not implemented")
}

func main() {
	supportVersions := version.PluginSupports("0.1.0", "0.2.0", "0.3.0", "0.3.1")
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, supportVersions, "mizni")
}
