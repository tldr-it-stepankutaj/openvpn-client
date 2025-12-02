package firewall

import (
	"context"

	"github.com/tldr-it-stepankutaj/openvpn-client/internal/api"
	"github.com/tldr-it-stepankutaj/openvpn-client/internal/config"
)

// UserWithNetworks represents a user with their allowed networks
type UserWithNetworks struct {
	Username string
	VpnIP    string
	Networks []string
}

// Firewall is the interface for firewall rule generators
type Firewall interface {
	// GenerateRules generates firewall rules for the given users
	GenerateRules(users []UserWithNetworks) string
	// GetRulesFile returns the path to the rules file
	GetRulesFile() string
	// GetReloadCommand returns the command to reload firewall rules
	GetReloadCommand() string
}

// New creates a new firewall based on configuration
func New(cfg *config.FirewallConfig) Firewall {
	switch cfg.Type {
	case "iptables":
		return NewIPTables(&cfg.IPTables)
	default:
		return NewNFTables(&cfg.NFTables)
	}
}

// CollectUserNetworks collects networks for all users from API
func CollectUserNetworks(ctx context.Context, client *api.Client, users []api.UserResponse) ([]UserWithNetworks, error) {
	var result []UserWithNetworks

	for _, user := range users {
		if user.VpnIP == "" {
			continue
		}

		routes, err := client.GetUserRoutes(ctx, user.ID)
		if err != nil {
			// Skip user on error, don't fail entire operation
			continue
		}

		networks := make([]string, 0)
		for _, route := range routes {
			// Skip default route for firewall rules
			if route.CIDR == "0.0.0.0/0" || route.CIDR == "0/0" {
				continue
			}
			networks = append(networks, route.CIDR)
		}

		if len(networks) > 0 {
			result = append(result, UserWithNetworks{
				Username: user.Username,
				VpnIP:    user.VpnIP,
				Networks: networks,
			})
		}
	}

	return result, nil
}
