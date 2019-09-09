package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/futurewei-cloud/cniplugins/alktron/neutron"
	"github.com/tkanos/gonfig"
)

// NeutronConfig keeps the config settings for neutron access
type NeutronConfig struct {
	User                        string `json:"user" env:"ALKTRON_USER"`
	Password                    string `json:"password" env:"ALKTRON_PASSWORD"`
	IdentityURL                 string `json:"identity_url" env:"ALKTRON_IDENTITYURL"`
	Host                        string `json:"host" env:"ALKTRON_HOST"`
	ProbeIntervalInMilliseconds uint32 `json:"interval_in_ms" env:"ALKTRON_PROBEINTERVALINMS"`
	ProbeTimeoutInSeconds       uint32 `json:"timeout_in_sec" env:"ALKTRON_PROBETIMEOUTINSEC"`
	Region                      string `json:"region" env:"ALKTRON_REGION"`
}

const (
	defaultNeutronConfPath = "/etc/alktron/neutron.json"
	defaultOpenStackRegion = "RegionOne"
)

func loadNeutronConfig() (*NeutronConfig, error) {
	neutronConfPath := os.Getenv("ALKTRON_NEUTRONCONF_PATH")
	if neutronConfPath == "" {
		neutronConfPath = defaultNeutronConfPath
	}

	c := &NeutronConfig{}
	if err := gonfig.GetConf(neutronConfPath, c); err != nil {
		return nil, fmt.Errorf("failed to load neutron conf: %v", err)
	}

	if strings.TrimSpace(c.Region) == "" {
		c.Region = defaultOpenStackRegion
	}

	return c, nil
}

func (c NeutronConfig) getNeutronClient(domain, vpc string) (*neutron.Client, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}

	return neutron.New(c.User, c.Password, c.Region, domain, vpc, c.IdentityURL)
}

func (c NeutronConfig) validate() error {
	if c.User == "" {
		return fmt.Errorf("invalid neutron config: User not specified")
	}

	if c.Password == "" {
		return fmt.Errorf("invalid neutron config: Password not specified")
	}

	if c.IdentityURL == "" {
		return fmt.Errorf("invalid neutron config: IdentityURL not specified")
	}

	return nil
}
