/*
Copyright 2019 The Arktos Authors.

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
package vnic_test

import (
	"reflect"
	"testing"

	"github.com/futurewei-cloud/cniplugins/vnic"
)

func TestLoadVNics(t *testing.T) {
	tcs := []struct {
		cniargs        string
		expected       *vnic.VNICs
		expectingError bool
	}{
		{`VPC=demo;NICs=[]`, nil, true},
		{`VPC=demo;NICs=[{}]`, nil, true},
		{`VPC=demo;NICs=[{"portid":"123456"}]`, &vnic.VNICs{Tenant: "default", VPC: "demo", NICs: []vnic.VNIC{{Name: "eth0", PortID: "123456"}}}, false},
		{`Tenant= ;VPC=demo;NICs=[{"portid":"123456"}]`, &vnic.VNICs{Tenant: "default", VPC: "demo", NICs: []vnic.VNIC{{Name: "eth0", PortID: "123456"}}}, false},
		{`Tenant=mydomain;VPC=demo;NICs=[{"portid":"123456"}]`, &vnic.VNICs{Tenant: "mydomain", VPC: "demo", NICs: []vnic.VNIC{{Name: "eth0", PortID: "123456"}}}, false},
		{`Tenant=mydomain;VPC=demo;NICs=[{"name":"eth1","portid":"123456"}]`, &vnic.VNICs{Tenant: "mydomain", VPC: "demo", NICs: []vnic.VNIC{{Name: "eth1", PortID: "123456"}}}, false},
	}

	for _, tc := range tcs {
		vnics, err := vnic.LoadVNICs(tc.cniargs)
		if !reflect.DeepEqual(tc.expected, vnics) {
			t.Errorf("input %q, expecting %v, got %v", tc.cniargs, tc.expected, vnics)
		}

		if tc.expectingError && err == nil {
			t.Errorf("input %q, expecting error, got nil", tc.cniargs)
		}

		if !tc.expectingError && err != nil {
			t.Errorf("input %q, expecting no error, got %v", tc.cniargs, err)
		}
	}
}
