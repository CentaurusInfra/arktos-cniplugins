package main

import (
	"errors"
	"fmt"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/futurewei-cloud/alktron/vnic"
)

func cmdAdd(args *skel.CmdArgs) error {
	// todo: load netconf, validate netns etc

	// to validate and parse vpc, portid from args.Args
	vnics, err := vnic.LoadVNICs(args.Args)
	if err != nil {
		return fmt.Errorf("ADD op failed to load cni args: %v", err)
	}

	// todo: to remove temporary go compiler tamer
	_ = vnics

	return errors.New("to be implemented")
}

func cmdDel(args *skel.CmdArgs) error {
	return errors.New("to be implemented")
}

func cmdCheck(args *skel.CmdArgs) error {
	return errors.New("to be implemented")
}

func main() {
	supportVersions := version.PluginSupports("0.1.0", "0.2.0", "0.3.0", "0.3.1")
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, supportVersions, "alktron")
}
