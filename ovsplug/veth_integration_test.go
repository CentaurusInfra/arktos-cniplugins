// +build integration

package ovsplug_test

import (
	"os"
	"testing"

	"github.com/futurewei-cloud/alktron/ovsplug"
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
