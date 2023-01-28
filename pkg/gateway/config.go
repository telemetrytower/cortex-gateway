package gateway

import (
	"flag"
	"fmt"
	"strings"
)

// Config for a gateway
type Config struct {
	DistributorAddress   string
	QueryFrontendAddress string
	RulerAddress         string
	AlertmanagerAddress  string
}

// RegisterFlags adds the flags required to config this package's Config struct
func (cfg *Config) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&cfg.DistributorAddress, "gateway.distributor.address", "", "Upstream HTTP URL for Cortex Distributor")
	f.StringVar(&cfg.QueryFrontendAddress, "gateway.query-frontend.address", "", "Upstream HTTP URL for Cortex Query Frontend")
	f.StringVar(&cfg.RulerAddress, "gateway.ruler.address", "", "Upstream HTTP URL for Cortex Ruler")
	f.StringVar(&cfg.AlertmanagerAddress, "gateway.alertmanager.address", "", "Upstream HTTP URL for Cortex Alertmanager")
}

// Validate given config parameters. Returns nil if everything is fine
func (cfg *Config) Validate() error {
	if cfg.DistributorAddress == "" {
		return fmt.Errorf("you must set -gateway.distributor.address")
	}

	if !strings.HasPrefix(cfg.DistributorAddress, "http") {
		return fmt.Errorf("distributor address must start with a valid scheme (http/https). Given is '%v'", cfg.DistributorAddress)
	}

	if cfg.QueryFrontendAddress == "" {
		return fmt.Errorf("you must set -gateway.query-frontend.address")
	}

	if !strings.HasPrefix(cfg.QueryFrontendAddress, "http") {
		return fmt.Errorf("query frontend address must start with a valid scheme (http/https). Given is '%v'", cfg.QueryFrontendAddress)
	}

	// gateway.ruler.address is empty if ruler proxy disabled.
	if cfg.RulerAddress != "" && !strings.HasPrefix(cfg.RulerAddress, "http") {
		return fmt.Errorf("ruler address must start with a valid scheme (http/https). Given is '%v'", cfg.RulerAddress)
	}

	// gateway.alertmanager.address is empty if alertmanager proxy disabled.
	if cfg.AlertmanagerAddress != "" && !strings.HasPrefix(cfg.AlertmanagerAddress, "http") {
		return fmt.Errorf("alertmanager address must start with a valid scheme (http/https). Given is '%v'", cfg.AlertmanagerAddress)
	}

	return nil
}
