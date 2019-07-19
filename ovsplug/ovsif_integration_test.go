// +build integration

package ovsplug_test

import (
	"os"
	"testing"

	"github.com/futurewei-cloud/alktron/ovsplug"
)

// sudo -E go test ./... -tags=integration -v -run OVSIF to run this suite
// need to set TEST_OVS_XXX env var, otherwise skipped

func TestOVSIFSetExtRes(t *testing.T) {
	port := os.Getenv("TEST_OVS_PORT")
	if port == "" {
		t.Skipf("Skipping duc to lack of TEST_OVS_PORT env var")
	}

	ovsif := ovsplug.NewOVSInterface(port)
	out, err := ovsif.SetExternalResource("qvo1234", "active", "00:11:22:33:44:55", "vm-uuid")

	t.Logf("output: %s", out)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
