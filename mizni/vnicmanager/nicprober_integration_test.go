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
// +build integration

// sudo -E go test ./... -v -tags=integration -run TestXXXX to run specific test case
// need to set env var TEST_XXXX_XXX, otherwise skipped

package vnicmanager

import (
	"os"
	"testing"
	"time"
)

func TestNSDeviceReady(t *testing.T) {
	nsPath := os.Getenv("TEST_DEVPROBER_NS_PATH")   //e.g. "/run/netns/x"
	devName := os.Getenv("TEST_DEVPROBER_DEV_NAME") //e.g. "veth123"
	if nsPath == "" || devName == "" {
		t.Skipf("Skipping due to lack of TEST_DEVPROBER_NS_PATH & TEST_DEVPROBER_DEV_NAME")
	}

	nicProber := &nicProberWithTimeout{timeout: time.Second * 15}

	if err := nicProber.DeviceReady(devName, nsPath); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
