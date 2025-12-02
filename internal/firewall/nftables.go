package firewall

import (
	"fmt"
	"sort"
	"strings"

	"github.com/tldr-it-stepankutaj/openvpn-client/internal/config"
)

// NFTables implements Firewall interface for nftables
type NFTables struct {
	rulesFile     string
	reloadCommand string
}

// NewNFTables creates a new NFTables firewall generator
func NewNFTables(cfg *config.NFTablesConfig) *NFTables {
	return &NFTables{
		rulesFile:     cfg.RulesFile,
		reloadCommand: cfg.ReloadCommand,
	}
}

// GenerateRules generates nftables rules for the given users
func (n *NFTables) GenerateRules(users []UserWithNetworks) string {
	var rules strings.Builder
	rules.WriteString("# Auto-generated VPN user rules (nftables)\n")
	rules.WriteString("# Do not edit manually - changes will be overwritten\n\n")

	for _, user := range users {
		// Sort networks for a consistent output
		sort.Strings(user.Networks)

		rules.WriteString(fmt.Sprintf("# %s\n", user.Username))
		rules.WriteString(fmt.Sprintf("ip saddr %s ip daddr { %s } accept\n",
			user.VpnIP,
			strings.Join(user.Networks, ", ")))
	}

	return rules.String()
}

// GetRulesFile returns the path to the rule file
func (n *NFTables) GetRulesFile() string {
	return n.rulesFile
}

// GetReloadCommand returns the command to reload firewall rules
func (n *NFTables) GetReloadCommand() string {
	return n.reloadCommand
}
