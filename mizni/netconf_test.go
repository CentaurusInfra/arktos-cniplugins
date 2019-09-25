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
