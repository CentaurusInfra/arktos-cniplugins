// +build integration

package neutron_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/futurewei-cloud/alktron/neutron"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
)

// Integration w/ available openstack setup
// use go test ./... -tags=integration [-v] to run integration test cases
// each case may require proper TEST_XXX env vars, otherwise may be skipped
var authOpts = gophercloud.AuthOptions{
	// IdentityEndpoint must be specified by env var TEST_OPENSTACK_URL of value like "http://127.0.0.1:5000/identity"
	Username:   "admin",
	Password:   "secret",
	TenantName: "demo",
	DomainName: "default",
}

func getNeutronClient(authOpts gophercloud.AuthOptions) (*neutron.Client, error) {
	identityEndpoint := os.Getenv("TEST_OPENSTACK_URL")
	if identityEndpoint == "" {
		return nil, fmt.Errorf("openstack server not specified by TEST_OPENSTACK_URL env var")
	}

	authOpts.IdentityEndpoint = identityEndpoint
	provider, err := openstack.AuthenticatedClient(authOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %v", err)
	}

	networkClient, err := openstack.NewNetworkV2(provider, gophercloud.EndpointOpts{
		Name:   "neutron",
		Region: "RegionOne",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get network service client: %v", err)
	}

	neutronClient := neutron.Client{ServiceClient: networkClient}
	return &neutronClient, nil
}

func TestGetPort(t *testing.T) {
	gwPortID := os.Getenv("TEST_GWPORT")
	if gwPortID == "" {
		t.Skipf("skipping due to lack of TEST_GWPORT env var")
	}

	neutronClient, err := getNeutronClient(authOpts)
	if err != nil {
		t.Fatalf("failed to get neutron client: %v", err)
	}

	portDetail, err := neutronClient.GetPort(gwPortID)
	if err != nil {
		t.Skipf("skipping (likely due to nonexistent portid): %v", err)
	}

	t.Logf("port detail %v", portDetail)
	if portDetail.PortsBindingExt.HostID == "" {
		t.Errorf("gw port expecting hostid, got empty")
	}
	if !portDetail.PortsBindingExt.VIFDetails["ovs_hybrid_plug"].(bool) {
		t.Errorf("expecting ovs_hybrid_plug true; got false")
	}
}

func TestBindPort(t *testing.T) {
	portID := os.Getenv("TEST_BINDPORT")
	hostID := os.Getenv("TEST_BINDHOST")
	if portID == "" || hostID == "" {
		t.Skipf("skipping due to lack of TEST_BINDPORT & TEST_BINDHOST env vars")
	}

	neutronClient, err := getNeutronClient(authOpts)
	if err != nil {
		t.Fatalf("failed to get neutron client: %v", err)
	}

	portDetail, err := neutronClient.BindPort(portID, hostID, "dev123")
	if err != nil {
		t.Errorf("failed to bind port %q to %q: %v", portID, hostID, err)
	}

	t.Logf("port detail: %v", portDetail)
	if hostID != portDetail.PortsBindingExt.HostID {
		t.Errorf("hostid expecting %s, got %s", hostID, portDetail.PortsBindingExt.HostID)
	}
	if !portDetail.PortsBindingExt.VIFDetails["ovs_hybrid_plug"].(bool) {
		t.Errorf("expecting ovs_hybrid_plug true; got false")
	}
}

func TestGetSubnet(t *testing.T) {
	subnetID := os.Getenv("TEST_SUBNET")
	if subnetID == "" {
		t.Skipf("skipping due to lack of TEST_SUBNET env var")
	}

	neutronClient, err := getNeutronClient(authOpts)
	if err != nil {
		t.Fatalf("failed to get neutron client: %v", err)
	}

	subnetDetail, err := neutronClient.GetSubnet(subnetID)
	if err != nil {
		t.Errorf("failed to get subnet %q: %v", subnetID, err)
	}

	t.Logf("port detail: %v", subnetDetail)
	if subnetID != subnetDetail.ID {
		t.Errorf("subnetID expecting %s, got %s", subnetID, subnetDetail.ID)
	}
}
