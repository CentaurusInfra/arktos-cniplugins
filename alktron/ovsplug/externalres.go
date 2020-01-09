/*
Copyright 2019 The Alkaid Authors.

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
package ovsplug

import (
	"fmt"
)

// ExternalResource represents external resources of an ovsdb interface expected by neutron agent
type ExternalResource struct {
	IFID, Status, AttachedMAC, VMUUID string
}

func (r ExternalResource) toExternalIds() []string {
	result := make([]string, 0)
	if r.IFID != "" {
		result = append(result, fmt.Sprintf("external-ids:iface-id=%s", r.IFID))
	}
	if r.Status != "" {
		result = append(result, fmt.Sprintf("external-ids:iface-status=%s", r.Status))
	}
	if r.AttachedMAC != "" {
		result = append(result, fmt.Sprintf("external-ids:attached-mac=%s", r.AttachedMAC))
	}
	if r.VMUUID != "" {
		result = append(result, fmt.Sprintf("external-ids:vm-uuid=%s", r.VMUUID))
	}

	return result
}
