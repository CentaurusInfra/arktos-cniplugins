// +build integration

package ovsplug_test

import (
	"os"
	"testing"

	"github.com/futurewei-cloud/alktron/ovsplug"
)

// sudo -E go test -tags=integration ./... -v -run Bridge to run these integration tests
// need to provide env var TEST_XXX, otherwise may be skipped

func TestBridgeNewAndUp(t *testing.T) {
	brName := os.Getenv("TEST_BRIDGE")
	if brName == "" {
		t.Skipf("Skipping due to lack of TEST_BRIDGE env var")
	}

	br, err := ovsplug.NewLinuxBridge(brName)

	if err != nil {
		t.Errorf("unexpected error on CreateLinuxBridge: %v", err)
	}

	if err = br.SetUp(); err != nil {
		t.Errorf("unexpected error on Setup: %v", err)
	}
}
