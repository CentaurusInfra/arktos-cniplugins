/*
Copyright 2019 The Alkaid Authors.

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
package vnicmanager

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"syscall"

	"github.com/containernetworking/plugins/pkg/ns"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

type nsdev struct{}

func (n nsdev) GetDevNetConf(name, nsPath string) (ipnet *net.IPNet, gw *net.IP, metric int, mac string, mtu int, err error) {
	err = ns.WithNetNSPath(nsPath, func(nsOrig ns.NetNS) error {
		itf, err := net.InterfaceByName(name)
		if err != nil {
			return fmt.Errorf("unable to find nic %s: %v", name, err)
		}

		mtu = itf.MTU
		mac = itf.HardwareAddr.String()

		// todo: consider support of multiple ip addresses
		if ipnet, err = getFirstIPNetV4(itf); err != nil {
			return err
		}

		if gw, metric, err = getV4Gateway(name); err != nil {
			return err
		}

		return nil
	})

	return
}

func (n nsdev) Migrate(nameFrom, nsPathFrom, nameTo, nsPathTo string, ipnet *net.IPNet, gw *net.IP, metric, mtu int) error {
	if err := moveDev(nameFrom, nsPathFrom, nsPathTo); err != nil {
		return fmt.Errorf("failed to move to target netns: %v", err)
	}

	if err := configDev(nameFrom, nameTo, nsPathTo, ipnet, mtu); err != nil {
		return fmt.Errorf("failed to config dev %s in ns %s: %v", nameTo, nsPathTo, err)
	}

	if err := confNet(nameTo, nsPathTo, gw, metric); err != nil {
		return fmt.Errorf("failed to config extra settings: %v", err)
	}

	ns, err := getNetns(nsPathTo)
	if err != nil {
		return fmt.Errorf("invalid ns path: %v", err)
	}
	if err := callMizarRequesites(nameTo, ns); err != nil {
		return fmt.Errorf("failed to run mizar requestite commands: %v", err)
	}

	return nil
}

func getFirstIPNetV4(itf *net.Interface) (*net.IPNet, error) {
	addrs, err := itf.Addrs()
	if err != nil {
		return nil, fmt.Errorf("unable to get ip address: %v", err)
	}

	for _, addr := range addrs {
		if ip, netmask, err := net.ParseCIDR(addr.String()); err == nil {
			if ip.To4() == nil {
				continue
			}

			return &net.IPNet{IP: ip, Mask: netmask.Mask}, nil
		}
	}

	return nil, fmt.Errorf("no ipv4 address found")
}

func getV4Gateway(device string) (*net.IP, int, error) {
	link, _ := netlink.LinkByName(device)
	routes, err := netlink.RouteList(link, netlink.FAMILY_V4)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to get route list of dev %s: %v", device, err)
	}

	// try to get from the default route entry
	for _, r := range routes {
		if r.Dst == nil {
			return &r.Gw, r.Priority, nil
		}
	}

	// fall back to first non-default entry
	for _, r := range routes {
		if r.Src == nil && r.Gw != nil {
			return &r.Gw, r.Priority, nil
		}
	}

	// todo: consider allowing nil gw
	return nil, 0, fmt.Errorf("unable to identify default gateway of dev %s", device)
}

func moveDev(name, nsPathFrom, nsPathTo string) error {
	return ns.WithNetNSPath(nsPathFrom, func(nsOrig ns.NetNS) error {
		l, err := netlink.LinkByName(name)
		if err != nil {
			return fmt.Errorf("unable to get device %s: %v", name, err)
		}

		nsTo, err := ns.GetNS(nsPathTo)
		if err != nil {
			return fmt.Errorf("invalid netns path %q: %v", nsPathTo, err)
		}

		if err := netlink.LinkSetNsFd(l, int(nsTo.Fd())); err != nil {
			return fmt.Errorf("unable to set netns to %q: %v", nsPathTo, err)
		}

		return nil
	})
}

func configDev(oldName, newName, nspath string, ipnet *net.IPNet, mtu int) error {
	return ns.WithNetNSPath(nspath, func(nsOrig ns.NetNS) error {
		link, err := netlink.LinkByName(oldName)
		if err != nil {
			return fmt.Errorf("failed to identify dev %q: %v", oldName, err)
		}

		if err := netlink.LinkSetName(link, newName); err != nil {
			return fmt.Errorf("failed to rename de to %q: %v", newName, err)
		}

		if err := netlink.LinkSetMTU(link, mtu); err != nil {
			return fmt.Errorf("failed to set mtu: %v", err)
		}

		addr := &netlink.Addr{IPNet: ipnet}
		if err = netlink.AddrAdd(link, addr); err != nil {
			return fmt.Errorf("failed to set ip addr %q: %v", ipnet.String(), err)
		}

		if err = netlink.LinkSetUp(link); err != nil {
			return fmt.Errorf("failed to bring up dev %s: %v", newName, err)
		}

		return nil
	})
}

func confNet(dev, nspath string, gw *net.IP, metric int) error {
	return ns.WithNetNSPath(nspath, func(nsOrig ns.NetNS) error {
		if err := setLoUp(); err != nil {
			return fmt.Errorf("unable to bring up lo dev")
		}

		defRoute := &netlink.Route{
			Dst:      nil, // default route entry
			Gw:       *gw,
			Priority: metric,
		}
		if err := netlink.RouteAdd(defRoute); err != nil {
			// fine if the exact route entry already exists
			if isDuplicateRouteEntryError(err) {
				log.Infof("duplicate default routing entry %s", defRoute.String())
				return nil
			}

			return fmt.Errorf("failed to configure nic, unable to add default route %q: %v", defRoute.String(), err)
		}

		return nil
	})
}

func setLoUp() error {
	lo, _ := netlink.LinkByName("lo")
	return netlink.LinkSetUp(lo)
}

func callMizarRequesites(dev, ns string) error {
	// per mizar requirements, extra commands need to be executed
	// ip netns exec {ep.ns} sysctl -w net.ipv4.tcp_mtu_probing=2
	// ip netns exec {ep.ns} ethtool -K {ep-dev} tso off gso off ufo off
	// ip netns exec {ep.ns} ethtool --offload {ep-dev} rx off tx off
	mizarCmds := fmt.Sprintf("sysctl -w net.ipv4.tcp_mtu_probing=2 && ethtool -K %s tso off gso off ufo off rx off tx off", dev)

	cmd := exec.Command("ip", "netns", "exec", ns, "sh", "-c", mizarCmds)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("exec got error: %v, detail output: %s", err, out)
	}

	return nil
}

func getNetns(nsPath string) (string, error) {
	const varPrefix = "/var"
	const varLen = len(varPrefix)

	if strings.HasPrefix(nsPath, varPrefix) {
		nsPath = nsPath[varLen:]
	}

	const nsPrefix = "/run/netns/"
	if !strings.HasPrefix(nsPath, nsPrefix) {
		return "", fmt.Errorf("unexpected ns path %s", nsPath)
	}

	return nsPath[len(nsPrefix):], nil
}

func isDuplicateRouteEntryError(err error) bool {
	syscallErr, ok := err.(syscall.Errno)
	if !ok {
		return false
	}

	return syscall.EEXIST == syscallErr
}
