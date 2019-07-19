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

type mockOVSInterface struct {
	mock.Mock
}

func (m *mockOVSInterface) SetExternalResource(name, status, mac, vm string) ([]byte, error) {
	args := m.Called(name, status, mac, vm)
	return args.Get(0).([]byte), args.Error(1)
}

type mockVEP struct {
	mock.Mock
}

func (v *mockVEP) GetName() string {
	args := v.Called()
	return args.String(0)
}

func TestHybridPlug(t *testing.T) {
	mockOVSBr := &mockBridge{}
	mockOVSBr.On("AddPort", "qvo123456789").Return(nil)

	mockOVSIf := &mockOVSInterface{}
	mockOVSIf.On("SetExternalResource", "1234567890abcdef", "active", "aa:bb:cc:dd:ee:ff", "libvirt-vm-id").Return([]byte{}, nil)

	mockLxBr := &mockBridge{}
	mockLxBr.On("AddPort", "qvb123456789").Return(nil)

	mockqvb := &mockVEP{}
	mockqvb.On("GetName").Return("qvb123456789")
	mockqvo := &mockVEP{}
	mockqvo.On("GetName").Return("qvo123456789")

	h := ovsplug.HybridPlug{
		NeutronPortID: "1234567890abcdef",
		MACAddr:       "aa:bb:cc:dd:ee:ff",
		VMID:          "libvirt-vm-id",

		OVSBridge:    mockOVSBr,
		OVSInterface: mockOVSIf,
		LinuxBridge:  mockLxBr,
		Qvb:          mockqvb,
		Qvo:          mockqvo,
	}

	if err := h.Plug(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	mockqvo.AssertExpectations(t)
	mockqvb.AssertExpectations(t)
	mockLxBr.AssertExpectations(t)
	mockOVSIf.AssertExpectations(t)
	mockOVSBr.AssertExpectations(t)
}
