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

func TestAttachMultipleVNICsGetIndexFrom0(t *testing.T) {
	vn0 := vnic.VNIC{}
	gw0 := net.ParseIP("10.0.36.1")
	pn0 := &vnic.EPnic{
		IPv4Net: &net.IPNet{IP: net.ParseIP("10.0.36.8"), Mask: net.CIDRMask(24, 32)},
		Gw:      &gw0,
	}

	vn1 := vnic.VNIC{}
	gw1 := net.ParseIP("10.0.37.1")
	pn1 := &vnic.EPnic{
		IPv4Net: &net.IPNet{IP: net.ParseIP("10.0.37.8"), Mask: net.CIDRMask(24, 32)},
		Gw:      &gw1,
	}

	mockPlugger := &mockPlugger{}
	mockPlugger.On("Plug", &vn0).Return(pn0, nil)
	mockPlugger.On("Plug", &vn1).Return(pn1, nil)

	r, err := attachVNICs(mockPlugger, []vnic.VNIC{vn0, vn1}, "dummy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Logf("result detail: %v", r)

	if len(r.IPs) != 2 {
		t.Fatalf("expecting 2 entries; got %d", len(r.IPs))
	}

	if *r.IPs[0].Interface != 0 {
		t.Fatalf("expecting starting from 0; got %d", *r.IPs[0].Interface)
	}

	if *r.IPs[1].Interface != 1 {
		t.Fatalf("expecting next is 1; got %d", *r.IPs[1].Interface)
	}
}

type mockUnplugger struct {
	mock.Mock
}

func (m *mockUnplugger) Unplug(vnic *vnic.VNIC) error {
	args := m.Called(vnic)
	return args.Error(0)
}

func TestDeleteVNICs(t *testing.T) {
	vn := vnic.VNIC{}

	mockUnplugger := &mockUnplugger{}
	mockUnplugger.On("Unplug", &vn).Return(nil)

	if err := detachVNICs(mockUnplugger, []vnic.VNIC{vn}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	mockUnplugger.AssertExpectations(t)
}
