// +build integration

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
