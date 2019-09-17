package vnicmanager

import (
	"fmt"
	"net"

	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/vishvananda/netlink"
)

type nsdev struct{}

func (n nsdev) GetDevNetConf(name, nsPath string) (ipnet *net.IPNet, gw *net.IP, mac string, mtu int, err error) {
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

		if gw, err = getV4Gateway(name); err != nil {
			return err
		}

		return nil
	})

	return
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

func getV4Gateway(device string) (*net.IP, error) {
	link, _ := netlink.LinkByName(device)
	routes, err := netlink.RouteList(link, netlink.FAMILY_V4)
	if err != nil {
		return nil, fmt.Errorf("unable to get route list of dev %s: %v", device, err)
	}

	// try to get from the default route entry
	for _, r := range routes {
		if r.Dst == nil {
			return &r.Gw, nil
		}
	}

	// fall back to first non-default entry
	for _, r := range routes {
		if r.Src == nil && r.Gw != nil {
			return &r.Gw, nil
		}
	}

	// todo: consider allowing nil gw
	return nil, fmt.Errorf("unable to identify default gateway of dev %s", device)
}
