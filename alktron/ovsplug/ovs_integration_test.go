// +build integration

package ovsplug_test

import (
	"os"
	"testing"

	"github.com/futurewei-cloud/plugins/alktron/ovsplug"
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
