package main

import (
	"net"
	"reflect"
	"testing"

	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/futurewei-cloud/cniplugins/vnic"
	"github.com/stretchr/testify/mock"
)

type mockPlugger struct {
	mock.Mock
}

func (m *mockPlugger) Plug(vn *vnic.VNIC) (*vnic.EPnic, error) {
	args := m.Called(vn)
	return args.Get(0).(*vnic.EPnic), args.Error(1)
}

func TestAttachVNICs(t *testing.T) {
	sandbox := "mysandbox"

	nicName0 := "nic0"
	mac0 := "72:4c:fc:2c:cd:b3"
	ipnet0 := &net.IPNet{IP: net.ParseIP("10.0.36.8"), Mask: net.CIDRMask(16, 32)}
	gw0 := net.ParseIP("10.0.0.1")

	vn0 := vnic.VNIC{
		Name:   nicName0,
		PortID: "123456-7890",
	}

	pn0 := &vnic.EPnic{
		Name:    nicName0,
		MAC:     mac0,
		IPv4Net: ipnet0,
		Gw:      &gw0,
	}

	mockPlugger := &mockPlugger{}
	mockPlugger.On("Plug", &vn0).Return(pn0, nil)

	intfExpected := &current.Interface{
		Name:    nicName0,
		Mac:     mac0,
		Sandbox: sandbox,
	}

	intfIndex := 0
	ipconfigExpected := &current.IPConfig{
		Version:   "4",
		Interface: &intfIndex,
		Address:   *ipnet0,
		Gateway:   gw0,
	}

	r, err := attachVNICs(mockPlugger, []vnic.VNIC{vn0}, sandbox)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Logf("result detail: %v", r)

	if len(r.Interfaces) != 1 || len(r.IPs) != 1 {
		t.Fatalf("unexpected elements returned")
	}

	if !reflect.DeepEqual(intfExpected, r.Interfaces[0]) {
		t.Errorf("Interface[0]: expecting %v; got %v", intfExpected, r.Interfaces[0])
	}

	if !reflect.DeepEqual(ipconfigExpected, r.IPs[0]) {
		t.Errorf("IPs[0]: expecting %v; got %v", ipconfigExpected, r.IPs[0])
	}

	mockPlugger.AssertExpectations(t)
}
