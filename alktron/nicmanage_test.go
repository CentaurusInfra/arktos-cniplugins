package main

import (
	"net"
	"reflect"
	"testing"

	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/futurewei-cloud/alktron/alktron/vnicplug"
	"github.com/futurewei-cloud/alktron/vnic"
	"github.com/stretchr/testify/mock"
)

type mockPlugger struct {
	mock.Mock
}

func (m *mockPlugger) Plug(vnic *vnic.VNIC, devID, boundHost string, routePrio int) (*vnicplug.EPnic, error) {
	args := m.Called(vnic, devID, boundHost, routePrio)
	return args.Get(0).(*vnicplug.EPnic), args.Error(1)
}

func (m *mockPlugger) Unplug(vnic *vnic.VNIC) error {
	args := m.Called(vnic)
	return args.Error(0)
}

func TestAttachVNICs(t *testing.T) {
	devID := "mysandbox"
	host := "a01.b.c"

	// nic 0 test setting
	nicName0 := "eth0"
	mac0 := "ba:be:fa:ce:11:00"
	gw0 := net.ParseIP("10.0.0.1")
	ipnet0 := &net.IPNet{IP: net.ParseIP("10.0.36.8"), Mask: net.CIDRMask(16, 32)}

	vn0 := vnic.VNIC{
		Name:   nicName0,
		PortID: "123456-7890",
	}

	pn0 := &vnicplug.EPnic{
		Name:    nicName0,
		MAC:     mac0,
		IPv4Net: ipnet0,
		Gw:      &gw0,
	}

	mockPlugger := &mockPlugger{}
	mockPlugger.On("Plug", &vn0, devID, host, 100).Return(pn0, nil)

	// expected cni result elements
	intfExpected := &current.Interface{
		Name:    nicName0,
		Mac:     mac0,
		Sandbox: devID,
	}

	intfIndex := 0
	ipconfigExpected := &current.IPConfig{
		Version:   "4",
		Interface: &intfIndex,
		Address:   *ipnet0,
		Gateway:   gw0,
	}

	r, err := attachVNICs(mockPlugger, []vnic.VNIC{vn0}, devID, host)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf("result detail: %v", *r)

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

func TestDetachVNICs(t *testing.T) {
	vn := vnic.VNIC{}

	mockPlugger := &mockPlugger{}
	mockPlugger.On("Unplug", &vn).Return(nil)

	if err := detachVNICs(mockPlugger, []vnic.VNIC{vn}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	mockPlugger.AssertExpectations(t)
}
