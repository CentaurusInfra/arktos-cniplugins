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
package vnicmanager_test

import (
	"errors"
	"net"
	"reflect"
	"strings"
	"testing"

	"github.com/futurewei-cloud/cniplugins/mizni/vnicmanager"
	"github.com/futurewei-cloud/cniplugins/vnic"
	"github.com/stretchr/testify/mock"
)

type mockDevNetConfGetter struct {
	mock.Mock
}

func (m *mockDevNetConfGetter) GetDevNetConf(name, nsPath string) (*net.IPNet, *net.IP, int, string, int, error) {
	args := m.Called(name, nsPath)
	return args.Get(0).(*net.IPNet), args.Get(1).(*net.IP), args.Int(2), args.String(3), args.Int(4), args.Error(5)
}

type mockDevProber struct {
	mock.Mock
}

func (m *mockDevProber) DeviceReady(name, nsPath string) error {
	args := m.Called(name, nsPath)
	return args.Error(0)
}

type mockNSMigrator struct {
	mock.Mock
}

func (m *mockNSMigrator) Migrate(nameFrom, nsFrom, nameTo, nsTo string, ipnet *net.IPNet, gw *net.IP, metric, mtu int) error {
	args := m.Called(nameFrom, nsFrom, nameTo, nsTo, ipnet, gw, metric, mtu)
	return args.Error(0)
}

func TestPlug(t *testing.T) {
	vpc := "88776655-deadbeef-0102"
	nsCNI := "nsDummy"

	vn := &vnic.VNIC{
		Name:   "dummy",
		PortID: "12345678-ABCDEF",
	}

	nsAlcor := "/run/netns/vpc-ns" + vpc
	devName := "veth12345678-AB"

	ipnet := &net.IPNet{IP: net.ParseIP("10.0.36.8"), Mask: net.CIDRMask(16, 32)}
	gw := net.ParseIP("10.0.0.1")
	metric := 100
	mac := "3e:36:8d:75:7a:ac"
	mtu := 1448

	mockNetConfGetter := &mockDevNetConfGetter{}
	mockNetConfGetter.On("GetDevNetConf", devName, nsAlcor).Return(ipnet, &gw, metric, mac, mtu, nil)

	mockDevProber := &mockDevProber{}
	mockDevProber.On("DeviceReady", devName, nsAlcor).Return(nil)

	mockNSMigrator := &mockNSMigrator{}
	mockNSMigrator.On("Migrate", devName, nsAlcor, vn.Name, nsCNI, ipnet, &gw, metric, mtu).Return(nil)

	expectedEPnic := &vnic.EPnic{
		Name:    vn.Name,
		MAC:     mac,
		IPv4Net: ipnet,
		Gw:      &gw,
	}

	manager := &vnicmanager.Manager{
		VPC:        vpc,
		NScni:      nsCNI,
		DevProber:  mockDevProber,
		ConfGetter: mockNetConfGetter,
		NSMigrator: mockNSMigrator,
	}

	epnic, err := manager.Plug(vn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf("%v", epnic)
	if !reflect.DeepEqual(expectedEPnic, epnic) {
		t.Errorf("expecting %v, got %v", expectedEPnic, epnic)
	}

	mockDevProber.AssertExpectations(t)
	mockNetConfGetter.AssertExpectations(t)
	mockNSMigrator.AssertExpectations(t)
}

func TestUnplug(t *testing.T) {
	devName := "ens3p1"
	nsCNI := "cni-ns"
	vpc := "88776655-deadbeef-0102"
	nicAlcor := "vethcafebaba"
	nsAlcor := "/run/netns/vpc-ns" + vpc

	vn := &vnic.VNIC{
		Name:   devName,
		PortID: "cafebaba",
	}

	ipnet := &net.IPNet{IP: net.ParseIP("10.0.36.8"), Mask: net.CIDRMask(16, 32)}
	gw := net.ParseIP("10.0.0.1")
	metric := 100
	mac := "3e:36:8d:75:7a:ac"
	mtu := 1448

	mockNetConfGetter := &mockDevNetConfGetter{}
	mockNetConfGetter.On("GetDevNetConf", devName, nsCNI).Return(ipnet, &gw, metric, mac, mtu, nil)
	_ = mac // not used in subsequent calls

	mockNSMigrator := &mockNSMigrator{}
	mockNSMigrator.On("Migrate", devName, nsCNI, nicAlcor, nsAlcor, ipnet, &gw, metric, mtu).Return(nil)

	manager := &vnicmanager.Manager{
		VPC:        vpc,
		NScni:      nsCNI,
		DevProber:  nil,
		ConfGetter: mockNetConfGetter,
		NSMigrator: mockNSMigrator,
	}

	if err := manager.Unplug(vn); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	mockNetConfGetter.AssertExpectations(t)
	mockNSMigrator.AssertExpectations(t)
}

func TestCleanupOnPlugError(t *testing.T) {
	vpc := "88776655-deadbeef-0102"
	nsCNI := "nsDummy"

	vn := &vnic.VNIC{
		Name:   "dummy",
		PortID: "12345678-ABCDEF",
	}

	nsAlcor := "/run/netns/vpc-ns" + vpc
	devName := "veth12345678-AB"

	ipnet := &net.IPNet{IP: net.ParseIP("10.0.36.8"), Mask: net.CIDRMask(16, 32)}
	gw := net.ParseIP("10.0.0.1")
	metric := 100
	mac := "3e:36:8d:75:7a:ac"
	mtu := 1448

	mockNetConfGetter := &mockDevNetConfGetter{}
	mockNetConfGetter.On("GetDevNetConf", devName, nsAlcor).Return(ipnet, &gw, metric, mac, mtu, nil)
	mockNetConfGetter.On("GetDevNetConf", "dummy", nsCNI).Return(ipnet, &gw, metric, mac, mtu, nil)

	mockDevProber := &mockDevProber{}
	mockDevProber.On("DeviceReady", devName, nsAlcor).Return(nil)

	mockNSMigrator := &mockNSMigrator{}
	mockNSMigrator.On("Migrate", devName, nsAlcor, vn.Name, nsCNI, ipnet, &gw, metric, mtu).Return(errors.New("mocked error"))
	mockNSMigrator.On("Migrate", vn.Name, nsCNI, devName, nsAlcor, ipnet, &gw, metric, mtu).Return(errors.New("error turned in to log warning anyway if any"))

	manager := &vnicmanager.Manager{
		VPC:        vpc,
		NScni:      nsCNI,
		DevProber:  mockDevProber,
		ConfGetter: mockNetConfGetter,
		NSMigrator: mockNSMigrator,
	}

	_, err := manager.Plug(vn)
	t.Logf("got returned error: %v", err)
	if err == nil || !strings.Contains(err.Error(), "mocked error") {
		t.Fatalf("expecting mocked error; got %v", err)
	}

	mockDevProber.AssertExpectations(t)
	mockNetConfGetter.AssertExpectations(t)
	mockNSMigrator.AssertExpectations(t)
}
