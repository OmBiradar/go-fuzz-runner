// examples/lnd/integration.go
package lnd

import (
	"github.com/OmBiradar/go-fuzz-runner/pkg/config"
)

// ConfigureForLND returns a configuration tailored for LND
func ConfigureForLND() *config.Config {
	cfg := config.Default()

	// Set LND-specific package patterns
	cfg.Packages = []string{
		"./lnwire/...",
		"./tlv/...",
		"./brontide/...",
		"./htlcswitch/...",
		"./zpay32/...",
		"./watchtower/wtwire/...",
	}

	// Allocate more time to critical packages
	cfg.TimeAllocation = map[string]float64{
		"lnwire":   0.4,  // 40% of time to critical wire protocol
		"brontide": 0.3,  // 30% to encryption
		"tlv":      0.1,  // 10% to TLV encoding
		"default":  0.05, // 5% to other packages
	}

	return cfg
}

// GetContinuousIntegrationConfig returns a configuration for CI runs
func GetContinuousIntegrationConfig() *config.Config {
	cfg := ConfigureForLND()

	// CI runs should be faster and only test changed packages
	cfg.ChangedOnly = true

	return cfg
}

// GetScheduledFuzzingConfig returns a configuration for scheduled fuzzing runs
func GetScheduledFuzzingConfig() *config.Config {
	cfg := ConfigureForLND()

	// Scheduled runs should be more thorough
	cfg.ChangedOnly = false
	cfg.FuzzTime = 30 * 60 * 1000000000 // 30 minutes per target

	return cfg
}
