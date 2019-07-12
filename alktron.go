package main

import (
	"errors"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/version"
)

func cmdAdd(args *skel.CmdArgs) error {
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
