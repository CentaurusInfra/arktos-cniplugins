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
