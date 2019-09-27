package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/futurewei-cloud/cniplugins/mizni/vnicmanager"
	"github.com/futurewei-cloud/cniplugins/vnic"
)

const capProbeTimeoutInMilliseconds = 1000 * 60 // 1 minute

func cmdAdd(args *skel.CmdArgs) error {
	vnics, err := vnic.LoadVNICs(args.Args)
	if err != nil {
		return fmt.Errorf("ADD op failed to load cni args: %v", err)
	}

	netconf, err := loadNetConf(args.StdinData)
	if err != nil {
		return fmt.Errorf("ADD op failed to load netconf: %v", err)
	}

	if netconf.ProbeTimeoutInMilliseconds > capProbeTimeoutInMilliseconds {
		return fmt.Errorf("Invalid netconf setting: prober timeout exceeds the cap of 60 seconds")
	}

	netns, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	}
	defer netns.Close()

	plugger := vnicmanager.New(vnics.VPC, args.Netns, time.Millisecond*time.Duration(netconf.ProbeTimeoutInMilliseconds))

	r, err := attachVNICs(plugger, vnics.NICs, args.ContainerID)
	if err != nil {
		return fmt.Errorf("ADD op failed to attach vnics: %v", err)
	}

	versionedResult, err := r.GetAsVersion(netconf.CNIVersion)
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

	unplugger := vnicmanager.New(vnics.VPC, args.Netns, 0) //probe timeout not used by unplugger; set 0 to satisfy the signiture
	return detachVNICs(unplugger, vnics.NICs)
}

func cmdCheck(args *skel.CmdArgs) error {
	return errors.New("not implemented")
}

func main() {
	supportVersions := version.PluginSupports("0.1.0", "0.2.0", "0.3.0", "0.3.1")
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, supportVersions, "mizni")
}
