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

// sudo -E go test ./... -tags=integration -v -run Veth to run this integration test set
// need to set TEST_VETH_XXX env var, otherwise skipped

func TestVethNewAndUp(t *testing.T) {
	vetha := os.Getenv("TEST_VETH_A")
	vethb := os.Getenv("TEST_VETH_B")
	if vetha == "" || vethb == "" {
		t.Skipf("Skipping due to lack of TEST_VETH_A & TEST_VETH_B env vars")
	}

	vethPair, err := ovsplug.NewVeth(vetha, vethb)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if vethPair.EP == nil {
		t.Errorf("expecting EP; got nil")
	}

	if vethPair.PeerEP == nil {
		t.Errorf("expecting PeerEP; got nil")
	}

	if err := vethPair.EP.SetUp(); err != nil {
		t.Errorf("unexpeted error setting EP up: %v", err)
	}
}
