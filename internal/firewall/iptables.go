package firewall

import (
	"fmt"
	"sort"
	"strings"

	"github.com/tldr-it-stepankutaj/openvpn-client/internal/config"
)

const defaultChainName = "VPN_USERS"

// IPTables implements Firewall interface for iptables
type IPTables struct {
	chainName     string
	rulesFile     string
	reloadCommand string
}

// NewIPTables creates a new IPTables firewall generator
func NewIPTables(cfg *config.IPTablesConfig) *IPTables {
	chainName := cfg.ChainName
	if chainName == "" {
		chainName = defaultChainName
	}
	return &IPTables{
		chainName:     chainName,
		rulesFile:     cfg.RulesFile,
		reloadCommand: cfg.ReloadCommand,
	}
}

// GenerateRules generates iptables rules for the given users
func (i *IPTables) GenerateRules(users []UserWithNetworks) string {
	var rules strings.Builder
	rules.WriteString("# Auto-generated VPN user rules (iptables)\n")
	rules.WriteString("# Do not edit manually - changes will be overwritten\n\n")
	rules.WriteString("*filter\n")

	// Create/flush chain
	rules.WriteString(fmt.Sprintf(":%s - [0:0]\n", i.chainName))
	rules.WriteString(fmt.Sprintf("-F %s\n", i.chainName))

	for _, user := range users {
		// Sort networks for a consistent output
		sort.Strings(user.Networks)

		rules.WriteString(fmt.Sprintf("# %s\n", user.Username))
		for _, network := range user.Networks {
			rules.WriteString(fmt.Sprintf("-A %s -s %s -d %s -j ACCEPT\n",
				i.chainName, user.VpnIP, network))
		}
	}

	rules.WriteString("COMMIT\n")
	return rules.String()
}

// GetRulesFile returns the path to the rule file
func (i *IPTables) GetRulesFile() string {
	return i.rulesFile
}

// GetReloadCommand returns the command to reload firewall rules
func (i *IPTables) GetReloadCommand() string {
	return i.reloadCommand
}
