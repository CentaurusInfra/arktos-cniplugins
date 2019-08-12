package ovsplug_test

import (
	"testing"

	"github.com/futurewei-cloud/alktron/ovsplug"
	"github.com/stretchr/testify/mock"
)

type mockBridge struct {
	mock.Mock
}

func (o *mockBridge) AddPort(port string) error {
	args := o.Called(port)
	return args.Error(0)
}

func (o *mockBridge) GetName() string {
	args := o.Called()
	return args.String(0)
}

func (o *mockBridge) DeletePort(port string) error {
	args := o.Called(port)
	return args.Error(0)
}

func (o *mockBridge) Delete() error {
	args := o.Called()
	return args.Error(0)
}

type mockExtResBridge struct {
	mock.Mock
}

func (m *mockExtResBridge) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockExtResBridge) AddPortAndSetExtResources(name, portID, status, mac, vm string) ([]byte, error) {
	args := m.Called(name, portID, status, mac, vm)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockExtResBridge) DeletePort(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

type mockVEP struct {
	mock.Mock
}

func (v *mockVEP) GetName() string {
	args := v.Called()
	return args.String(0)
}

func TestHybridPlug(t *testing.T) {
	portName := "qvo123456789"
	portID := "1234567890abcdef"
	mac := "aa:bb:cc:dd:ee:ff"
	vm := "libvirt-vm-id"

	mockOVSBr := &mockExtResBridge{}
	mockOVSBr.On("AddPortAndSetExtResources", portName, portID, "active", mac, vm).Return([]byte{}, nil)

	mockLxBr := &mockBridge{}
	mockLxBr.On("AddPort", "qvb123456789").Return(nil)

	mockqvb := &mockVEP{}
	mockqvb.On("GetName").Return("qvb123456789")
	mockqvo := &mockVEP{}
	mockqvo.On("GetName").Return("qvo123456789")

	h := ovsplug.HybridPlug{
		NeutronPortID: portID,
		MACAddr:       mac,
		VMID:          vm,

		OVSBridge:   mockOVSBr,
		LinuxBridge: mockLxBr,
		Qvb:         mockqvb,
		Qvo:         mockqvo,
	}

	if err := h.Plug(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	mockqvo.AssertExpectations(t)
	mockqvb.AssertExpectations(t)
	mockLxBr.AssertExpectations(t)
	mockOVSBr.AssertExpectations(t)
}

func TestHybridUnplug(t *testing.T) {
	qvoPort := "qvo123456789"
	qvbPort := "qvb123456789"
	portID := "1234567890abcdef"
	mac := "aa:bb:cc:dd:ee:ff"
	vm := "libvirt-vm-id"

	mockOVSBr := &mockExtResBridge{}
	mockOVSBr.On("DeletePort", qvoPort).Return(nil)

	mockLxBr := &mockBridge{}
	mockLxBr.On("DeletePort", qvbPort).Return(nil)
	mockLxBr.On("Delete").Return(nil)

	mockqvb := &mockVEP{}
	mockqvb.On("GetName").Return(qvbPort)
	mockqvo := &mockVEP{}
	mockqvo.On("GetName").Return(qvoPort)

	h := ovsplug.HybridPlug{
		NeutronPortID: portID,
		MACAddr:       mac,
		VMID:          vm,

		OVSBridge:   mockOVSBr,
		LinuxBridge: mockLxBr,
		Qvb:         mockqvb,
		Qvo:         mockqvo,
	}

	if err := h.Unplug(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	mockqvo.AssertExpectations(t)
	mockqvb.AssertExpectations(t)
	mockLxBr.AssertExpectations(t)
	mockOVSBr.AssertExpectations(t)
}
