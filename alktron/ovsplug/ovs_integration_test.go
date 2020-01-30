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

package ovsplug_test

import (
	"os"
	"testing"

	"github.com/futurewei-cloud/cniplugins/alktron/ovsplug"
)

// sudo -E go test ./... -tags=integration -v -run TestOVSXXX to run specific case
// need to set TEST_OVS_XXX env vars, otherwise skipped

func TestOVSAddPortAndSetExtResources(t *testing.T) {
	ovsBr := os.Getenv("TEST_OVS_BR")
	port := os.Getenv("TEST_OVS_PORT")
	if ovsBr == "" || port == "" {
		t.Skipf("Skipping due to lack of TEST_OVS_BR & TEST_OVS_PORT env var")
	}
	br := ovsplug.NewOVSBridge(ovsBr)

	out, err := br.AddPortAndSetExtResources(port, "port-id", "active", "00:11:22:33:44:55", "vm-uuid")

	if err != nil {
		t.Errorf("unexpected error, out=%s: %v", string(out), err)
	}
}

func TestOVSDeletePort(t *testing.T) {
	ovsBr := os.Getenv("TEST_OVS_BR")
	port := os.Getenv("TEST_OVS_PORT")
	if ovsBr == "" || port == "" {
		t.Skipf("Skipping due to lack of TEST_OVS_BR & TEST_OVS_PORT env var")
	}
	br := ovsplug.NewOVSBridge(ovsBr)

	if err := br.DeletePort(port); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
