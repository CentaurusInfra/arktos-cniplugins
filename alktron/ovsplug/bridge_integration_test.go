// +build integration

package ovsplug_test

import (
	"os"
	"testing"

	"github.com/futurewei-cloud/alktron/ovsplug"
)

// sudo -E go test -tags=integration ./... -v -run TestBridgeXXX to run specific case
// need to provide env var TEST_XXX, otherwise may be skipped

func TestBridgeNewAndUp(t *testing.T) {
	brName := os.Getenv("TEST_BRIDGE")
	if brName == "" {
		t.Skipf("Skipping due to lack of TEST_BRIDGE env var")
	}

	br := ovsplug.NewLinuxBridge(brName)

	if err := br.InitDevice(); err != nil {
		t.Errorf("unexpected error on CreateLinuxBridge: %v", err)
	}

	if err := br.SetUp(); err != nil {
		t.Errorf("unexpected error on Setup: %v", err)
	}
}

func TestBridgeDeleteBr(t *testing.T) {
	brName := os.Getenv("TEST_BRIDGE")
	if brName == "" {
		t.Skipf("Skipping due to lack of TEST_BRIDGE env var")
	}

	br := ovsplug.NewLinuxBridge(brName)
	if err := br.InitDevice(); err != nil {
		t.Skipf("skipping; test setup failed: %v", err)
	}

	if err := br.Delete(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBridgeDeletePort(t *testing.T) {
	brName := os.Getenv("TEST_BRIDGE")
	brPort := os.Getenv("TEST_BRIDGE_PORT")
	if brName == "" || brPort == "" {
		t.Skipf("Skipping due to lack of TEST_BRIDGE & TEST_BRIDGE_PORT env var")
	}

	br := ovsplug.NewLinuxBridge(brName)
	if err := br.InitDevice(); err != nil {
		t.Skipf("skipping; test setup failed: %v", err)
	}

	if err := br.DeletePort(brPort); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
