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
package main

import (
	"reflect"
	"testing"

	"github.com/containernetworking/cni/pkg/types"
)

func TestLoadNetConf(t *testing.T) {
	tcs := []struct {
		name     string
		input    string
		expected *netConf
	}{
		{
			name: "probe-tm-notprovided",
			input: `{
				"cniVersion": "0.3.1",
				"name": "mizni-simu",
				"type": "mizni"
			  }
			`,
			expected: &netConf{
				NetConf: types.NetConf{
					CNIVersion: "0.3.1",
					Name:       "mizni-simu",
					Type:       "mizni",
				},
				ProbeTimeoutInMilliseconds: 0,
			},
		},
		{
			name: "probe-tm-125",
			input: `{
				"cniVersion": "0.3.1",
				"name": "mizni-simu",
				"type": "mizni",
				"probe_tm_ms": 125
			  }
			`,
			expected: &netConf{
				NetConf: types.NetConf{
					CNIVersion: "0.3.1",
					Name:       "mizni-simu",
					Type:       "mizni",
				},
				ProbeTimeoutInMilliseconds: 125,
			},
		},
	}

	for _, tc := range tcs {
		n, err := loadNetConf([]byte(tc.input))
		if err != nil {
			t.Fatalf("tc %s got unexpected error: %v", tc.name, err)
		}

		t.Logf("tc %s netconf detail: %v", tc.name, n)
		if !reflect.DeepEqual(tc.expected, n) {
			t.Errorf("tc %s expecting %v, got %v", tc.name, tc.expected, n)
		}
	}
}
