package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/containernetworking/cni/pkg/types"

	"github.com/futurewei-cloud/alktron/neutron"
	"github.com/futurewei-cloud/alktron/vnicplug"

	"github.com/containernetworking/plugins/pkg/ns"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/futurewei-cloud/alktron/vnic"
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

	neutronClient, err := getNeutronClient(vnics.VPC)
	if err != nil {
		return fmt.Errorf("failed to get neutron client: %v", err)
	}

	hostBound, err := getEnvVarValue("ALKTRON_HOST")
	if err != nil {
		// todo: use localhost as default
		return err
	}

	plugger := vnicplug.NewPlugger(neutronClient, args.Netns)
	// todo: consider a better devID (like pod id?) than args.ContainerID
	r, err := attachVNICs(plugger, vnics.NICs, args.ContainerID, hostBound)
	if err != nil {
		return fmt.Errorf("ADD op failed to attach vnics: %v", err)
	}

	versionedResult, err := r.GetAsVersion(cniVersion)
	return versionedResult.Print()
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

func getCNIVerInNetConf(bytes []byte) (string, error) {
	n := &types.NetConf{}
	if err := json.Unmarshal(bytes, n); err != nil {
		return "", fmt.Errorf("failed to load netconf: %v", err)
	}
	return n.CNIVersion, nil
}

func getNeutronClient(vpc string) (*neutron.Client, error) {
	user, err := getEnvVarValue("ALKTRON_USER")
	if err != nil {
		return nil, err
	}

	password, err := getEnvVarValue("ALKTRON_PASSWORD")
	if err != nil {
		return nil, err
	}

	identityURL, err := getEnvVarValue("ALKTRON_IDENTITYURL")
	if err != nil {
		return nil, err
	}

	return neutron.New(user, password, vpc, identityURL)
}

func getEnvVarValue(name string) (string, error) {
	val := os.Getenv(name)
	if val == "" {
		return "", fmt.Errorf("invalid env var %q: empty not allowed", name)
	}

	return val, nil
}
