// +build integration

package main

import (
	"testing"
)

func TestLoadNeutronConfig(t *testing.T) {
	c, err := loadNeutronConfig()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf("config detail: %v", *c)

	if err := c.validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
