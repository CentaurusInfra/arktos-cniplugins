// +build integration

package ovsplug_test

import (
	"os"
	"testing"

	"github.com/futurewei-cloud/alktron/ovsplug"
)

// go test ./... -tags=integration -v -run OVSBr to run this suite
// need to set TEST_OVS_XXX env vars, otherwise skipped

func TestOVSBrAddPort(t *testing.T) {
	ovsbr := os.Getenv("TEST_OVS_BR")
	port := os.Getenv("TEST_OVS_PORT")
	if ovsbr == "" || port == "" {
		t.Skipf("Skipping due to lack of TEST_OVS_BR & TEST_OVS_PORT env vars")
	}

	ob := ovsplug.NewOVSBridge(ovsbr)
	err := ob.AddPort(port)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
