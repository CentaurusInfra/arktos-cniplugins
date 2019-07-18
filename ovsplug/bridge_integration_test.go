// +build integration

package ovsplug_test

import (
	"os"
	"testing"

	"github.com/futurewei-cloud/alktron/ovsplug"
)

// sudo -E go test -tags=integration ./... -v to run these integration tests
// need to provide env var TEST_XXX, otherwise may be skipped

func TestBridgeIsCreated(t *testing.T) {
	brName := os.Getenv("TEST_BRIDGE")
	if brName == "" {
		t.Skipf("Skipping due to lack of TEST_BRIDGE env var")
	}

	br := ovsplug.LinuxBridge{Name: brName}
	created, err := br.IsCreated()
	if err != nil {
		t.Fatalf("failed to get bridge %q: %v", brName, err)
	}

	if created {
		t.Errorf("expecting no bridge %q; got yes", brName)
	}
}

func TestBridgeCreate(t *testing.T) {
	brName := os.Getenv("TEST_BRIDGE")
	if brName == "" {
		t.Skipf("Skipping due to lack of TEST_BRIDGE env var")
	}

	br := ovsplug.LinuxBridge{Name: brName}
	if err := br.Create(); err != nil {
		t.Errorf("failed to create bridge %q: %v", brName, err)
	}

	created, _ := br.IsCreated()
	if !created {
		t.Errorf("bridge %q should have been created; found not yet", br.Name)
	}
}
