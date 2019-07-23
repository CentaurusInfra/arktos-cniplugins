package vnicplug_test

import (
	"net"
	"reflect"
	"testing"

	"github.com/futurewei-cloud/alktron/neutron"
	"github.com/futurewei-cloud/alktron/ovsplug"
	"github.com/futurewei-cloud/alktron/vnic"
	"github.com/futurewei-cloud/alktron/vnicplug"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
	"github.com/stretchr/testify/mock"
)

type mockPortGetBinder struct {
	mock.Mock
}

func (m *mockPortGetBinder) GetPort(portID string) (*neutron.PortBindingDetail, error) {
	args := m.Called(portID)
	return args.Get(0).(*neutron.PortBindingDetail), args.Error(1)
}

func (m *mockPortGetBinder) BindPort(portID, hostID, devID string) (*neutron.PortBindingDetail, error) {
	args := m.Called(portID, hostID, devID)
	return args.Get(0).(*neutron.PortBindingDetail), args.Error(1)
}

type mockSubnetGetter struct {
	mock.Mock
}

func (m *mockSubnetGetter) GetSubnet(subnetID string) (*subnets.Subnet, error) {
	args := m.Called(subnetID)
	return args.Get(0).(*subnets.Subnet), args.Error(1)
}

type mockLocalPlugger struct {
	mock.Mock
}

func (m *mockLocalPlugger) Plug() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockLocalPlugger) GetLocalBridge() string {
	args := m.Called()
	return args.String(0)
}

type mockDevNetnsPlugger struct {
	mock.Mock
}

func (m *mockDevNetnsPlugger) Plug(name, mac string, cidr *net.IPNet, gw *net.IP, netns, hostBr string) error {
	args := m.Called(name, mac, cidr, gw, netns, hostBr)
	return args.Error(0)
}

func TestPlug(t *testing.T) {
	boundHost := "local-host"
	devID := "vm-id"

	qbr := "qbr12345678"
	netns := "dummy_ns"
	portID := "12345678"
	nicName := "tap-dev"
	subnetID := "subnet-id"
	mac := "11:22:33:44:55:66"
	gw, ipv4Net, _ := net.ParseCIDR("10.0.0.1/24")
	ipnet := &net.IPNet{
		IP:   net.ParseIP("10.0.0.4"),
		Mask: ipv4Net.Mask,
	}

	portBindingDetail := &neutron.PortBindingDetail{
		Port: ports.Port{
			MACAddress: mac,
			FixedIPs: []ports.IP{
				{SubnetID: subnetID, IPAddress: "10.0.0.4"},
			},
		},
	}

	portActiveDetail := &neutron.PortBindingDetail{
		Port: ports.Port{Status: "Active"},
	}

	subnetDetail := &subnets.Subnet{
		CIDR: "10.0.0.1/24",
	}

	mockPortGetBinder := &mockPortGetBinder{}
	mockPortGetBinder.On("BindPort", portID, boundHost, devID).Return(portBindingDetail, nil)
	mockPortGetBinder.On("GetPort", portID).Return(portActiveDetail, nil)

	mockSubnetGetter := &mockSubnetGetter{}
	mockSubnetGetter.On("GetSubnet", subnetID).Return(subnetDetail, nil)

	mockLocalPlugger := &mockLocalPlugger{}
	mockLocalPlugger.On("Plug").Return(nil)
	mockLocalPlugger.On("GetLocalBridge").Return(qbr)

	hybridPlugGen := func(portID, mac, vm string) (ovsplug.LocalPlugger, error) {
		return mockLocalPlugger, nil
	}

	mockDevNetnsPlugger := &mockDevNetnsPlugger{}
	mockDevNetnsPlugger.On("Plug", nicName, mac, ipnet, &gw, netns, qbr).Return(nil)

	plugger := vnicplug.Plugger{
		PortGetBinder:   mockPortGetBinder,
		SubnetGetter:    mockSubnetGetter,
		HybridPlugGen:   hybridPlugGen,
		DevNetnsPlugger: mockDevNetnsPlugger,
	}

	vnic := vnic.VNIC{
		PortID: portID,
		Name:   nicName,
	}

	expectedEPnic := &vnicplug.EPnic{
		Name:    nicName,
		MAC:     mac,
		Gw:      &gw,
		IPv4Net: ipnet,
	}

	epNIC, err := plugger.Plug(&vnic, devID, boundHost, netns)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	t.Logf("nic in netns is %v", *epNIC)
	if !reflect.DeepEqual(expectedEPnic, epNIC) {
		t.Errorf("expecting %v, got %v", expectedEPnic, epNIC)
	}

	mockSubnetGetter.AssertExpectations(t)
	mockLocalPlugger.AssertExpectations(t)
	mockPortGetBinder.AssertExpectations(t)
	mockDevNetnsPlugger.AssertExpectations(t)
}
