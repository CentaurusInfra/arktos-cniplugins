package nsvtep

import (
	"fmt"
	"net"
	"strings"

	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/futurewei-cloud/cniplugins/alktron/ovsplug"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

// Manager represens the manager for endpoint nic inside of the specific netns
type Manager struct {
	NSPath string
}

// NewManager creates a manager
func NewManager(nspath string) *Manager {
	return &Manager{NSPath: nspath}
}

// Attach creates and attaches the endpoint nic (inside netns) to a host bridge
func (p Manager) Attach(dev string, mac net.HardwareAddr, ipnet *net.IPNet, gw *net.IP, prio int, hostBr string) error {
	netns, err := ns.GetNS(p.NSPath)
	if err != nil {
		return fmt.Errorf("invalid netns path %q: %v", p.NSPath, err)
	}
	defer netns.Close()

	lxbr := ovsplug.NewLinuxBridge(hostBr)
	if err := lxbr.InitDevice(); err != nil {
		return fmt.Errorf("failed to create bridge %q: %v", hostBr, err)
	}

	vtepTemp, err := createBridgedVTEPInNs(lxbr, netns)
	if err != nil {
		return fmt.Errorf("failed to hook endpoint nic %q inside netns %q to host bridge %q: %v", dev, p.NSPath, hostBr, err)
	}

	return ns.WithNetNSPath(p.NSPath, func(nsOrig ns.NetNS) error {
		return configNet(vtepTemp, dev, mac, ipnet, gw, prio)
	})
}

// Detach deletes the veth pair of tap device (and the peer inside netns), removes from host bridge
func (p Manager) Detach(dev, hostBr string) error {
	epTap := "tap" + hostBr[3:]

	if err := ovsplug.RemoveVEP(epTap); err != nil {
		return fmt.Errorf("failed to detach vtep %q for vnic %q: %v", epTap, dev, err)
	}

	return nil
}

func createBridgedVTEPInNs(lxbr *ovsplug.LinuxBridge, netns ns.NetNS) (string, error) {
	// creates veth pair, one end connecting to host bridge, the other across netns
	ep := "qvn" + lxbr.Name[3:]
	// neutron agent is expecting endpoint named as "tap<port-id-prefix>" attached to local bridge
	epOnBr := "tap" + lxbr.Name[3:]
	if _, err := ovsplug.NewVeth(ep, epOnBr); err != nil {
		return "", fmt.Errorf("failed to create veth pair (%q, %q): %v", ep, epOnBr, err)
	}
	lxbr.AddPort(epOnBr)

	if err := setDevNetns(ep, netns); err != nil {
		if err := ovsplug.RemoveVEP(epOnBr); err != nil {
			log.Warnf("cleanup failed, netns vtep %q: %v", epOnBr, err)
		}
		return "", err
	}

	return ep, nil
}

func setDevNetns(dev string, netns ns.NetNS) error {
	l, err := netlink.LinkByName(dev)
	if err != nil {
		return err
	}

	return netlink.LinkSetNsFd(l, int(netns.Fd()))
}

func configNet(devTemp, dev string, mac net.HardwareAddr, ipnet *net.IPNet, gw *net.IP, prio int) error {
	if err := setLoUp(); err != nil {
		return fmt.Errorf("failed to set lo up: %v", err)
	}

	if err := renameDev(devTemp, dev); err != nil {
		return fmt.Errorf("failed to configure nic, unable to rename %q to %q: %v", devTemp, dev, err)
	}

	if err := configDev(dev, mac, ipnet); err != nil {
		return fmt.Errorf("failed to configure nic: %v", err)
	}

	defRoute := &netlink.Route{
		Dst:      nil, // default route entry
		Gw:       *gw,
		Priority: prio,
	}
	if err := netlink.RouteAdd(defRoute); err != nil {
		// fine if route entry already exists
		if !strings.Contains(err.Error(), "file exists") {
			return fmt.Errorf("failed to configure nic, unable to add default route %q: %v", defRoute.String(), err)
		}
	}

	return nil
}

func setLoUp() error {
	lo, _ := netlink.LinkByName("lo")
	return netlink.LinkSetUp(lo)
}

func renameDev(old, new string) error {
	link, err := netlink.LinkByName(old)
	if err != nil {
		return fmt.Errorf("failed to rename net device; %q not found: %v", old, err)
	}

	return netlink.LinkSetName(link, new)
}

func configDev(dev string, mac net.HardwareAddr, ipnet *net.IPNet) error {
	link, err := netlink.LinkByName(dev)
	if err != nil {
		return fmt.Errorf("failed to configure dev, unable to find %q: %v", dev, err)
	}

	if err = netlink.LinkSetHardwareAddr(link, mac); err != nil {
		return fmt.Errorf("failed to configure dev, unable to set %q mac address %q: %v", dev, mac.String(), err)
	}

	addr := &netlink.Addr{IPNet: ipnet}
	if err = netlink.AddrAdd(link, addr); err != nil {
		return fmt.Errorf("failed to configure dev, unable to set %q ip addr %q: %v", dev, ipnet.String(), err)
	}

	if err = netlink.LinkSetUp(link); err != nil {
		return fmt.Errorf("failed to configure dev, unable to set %q up: %v", dev, err)
	}

	return nil
}
