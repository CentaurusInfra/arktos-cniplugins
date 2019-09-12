package main

import (
	"errors"
	"fmt"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/version"
)

func cmdAdd(args *skel.CmdArgs) error {
	return fmt.Errorf("to be implemeneted")
}

func cmdDel(args *skel.CmdArgs) error {
	return fmt.Errorf("to be implemeneted")
}

func cmdCheck(args *skel.CmdArgs) error {
	return errors.New("not implemented")
}

func main() {
	supportVersions := version.PluginSupports("0.1.0", "0.2.0", "0.3.0", "0.3.1")
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, supportVersions, "mizni")
}
