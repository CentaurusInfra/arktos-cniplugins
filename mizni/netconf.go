package main

import (
	"encoding/json"
	"fmt"

	"github.com/containernetworking/cni/pkg/types"
)

// netConf represents the customized netconf for mizni plugin
type netConf struct {
	types.NetConf
	ProbeTimeoutInMilliseconds uint32 `json:"probe_tm_ms"`
}

func loadNetConf(bytes []byte) (*netConf, error) {
	n := &netConf{}
	if err := json.Unmarshal(bytes, n); err != nil {
		return nil, fmt.Errorf("failed to load netconf: %v", err)
	}

	return n, nil
}
