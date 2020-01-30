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
// +build integration

package nsvtep_test

import (
	"net"
	"os"
	"testing"

	"github.com/futurewei-cloud/cniplugins/alktron/nsvtep"
)

// sudo -E go test ./... -v -tags=integration -run NSvtepAttach to run this suite
// need to set env var TEST_NSVTEP_XXX, otherwise skipped

func TestNSvtepAttach(t *testing.T) {
	hostBr := os.Getenv("TEST_NSVTEP_HOST_BR")
	nspath := os.Getenv("TEST_NSVTEP_NETNS_PATH")
	if hostBr == "" || nspath == "" {
		t.Skipf("Skipping due to lack of TEST_NSVTEP_HOST_BR & TEST_NSVTEP_NETNS_PATH")
	}

	dev := "mynic"
	mac, _ := net.ParseMAC("de:ad:be:ef:a7:a7")
	ipnet := net.IPNet{
		IP:   net.ParseIP("10.0.0.4"),
		Mask: net.CIDRMask(24, 32),
	}
	gw := net.IPv4(10, 0, 0, 2)

	nsvtepManager := nsvtep.Manager{NSPath: nspath}
	if err := nsvtepManager.Attach(dev, mac, &ipnet, &gw, 88, hostBr); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNSvtepDetach(t *testing.T) {
	nspath := os.Getenv("TEST_NSVTEP_NETNS_PATH")
	hostBr := os.Getenv("TEST_NSVTEP_HOST_BR")
	if hostBr == "" || nspath == "" {
		t.Skipf("Skipping due to lack of TEST_NSVTEP_HOST_BR & TEST_NSVTEP_NETNS_PATH")
	}

	nsvtepManager := nsvtep.Manager{NSPath: nspath}
	if err := nsvtepManager.Detach("dummy", hostBr); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
