package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/tldr-it-stepankutaj/openvpn-client/internal/api"
	"github.com/tldr-it-stepankutaj/openvpn-client/internal/config"
	"github.com/tldr-it-stepankutaj/openvpn-client/internal/logger"
	"github.com/tldr-it-stepankutaj/openvpn-client/internal/utils"
)

const programName = "openvpn-connect"

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "path to configuration file")
	flag.StringVar(&configPath, "c", "", "path to configuration file (shorthand)")
	flag.Parse()

	// Initialize logger
	log := logger.New(logger.Options{
		Level:   slog.LevelInfo,
		JSON:    true,
		Program: programName,
	})

	// Get OpenVPN config a file path from a positional argument
	args := flag.Args()
	if len(args) < 1 {
		log.Error("OpenVPN config file path not provided")
		os.Exit(1)
	}

	openvpnConfigFile := args[0]

	// Get environment variables from OpenVPN
	commonName := os.Getenv("common_name")
	if commonName == "" {
		log.Error("common_name environment variable not set")
		os.Exit(1)
	}

	trustedIP := os.Getenv("trusted_ip")
	trustedPort := os.Getenv("trusted_port")
	remoteIP := os.Getenv("ifconfig_pool_remote_ip")

	userLog := log.WithUser(commonName)

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		userLog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Create API client
	client := api.NewClient(&cfg.API)

	ctx := context.Background()

	// Authenticate if using a legacy service account
	if !cfg.API.UseToken() {
		if err := client.Authenticate(ctx, cfg.API.Username, cfg.API.Password); err != nil {
			userLog.Error("API authentication failed", "error", err)
			os.Exit(1)
		}
	}

	// Get user by username
	user, err := client.GetUserByUsername(ctx, commonName)
	if err != nil {
		userLog.Error("user not found", "error", err)
		os.Exit(1)
	}

	userLog = userLog.With("user_id", user.ID)

	// Build config file content
	var configContent strings.Builder

	// Set static VPN IP if configured
	vpnIP := remoteIP
	if user.VpnIP != "" {
		vpnIP = user.VpnIP
		configContent.WriteString(fmt.Sprintf("ifconfig-push %s 255.255.255.0\n", user.VpnIP))
	}

	// Get user's routes
	routes, err := client.GetUserRoutes(ctx, user.ID)
	if err != nil {
		userLog.Error("failed to get user routes", "error", err)
		os.Exit(1)
	}

	// Check for default route and collect networks
	hasDefaultRoute := false
	var networks []string

	for _, route := range routes {
		if utils.IsDefaultRoute(route.CIDR) {
			hasDefaultRoute = true
			continue
		}
		networks = append(networks, route.CIDR)
	}

	// Push routes
	if hasDefaultRoute {
		configContent.WriteString("push \"redirect-gateway def1\"\n")
	} else {
		for _, cidr := range networks {
			route, err := utils.CIDRToNetmask(cidr)
			if err != nil {
				userLog.Warn("invalid CIDR, skipping", "cidr", cidr, "error", err)
				continue
			}
			configContent.WriteString(fmt.Sprintf("push \"route %s\"\n", route))
		}
	}

	// Write a config file
	if err := os.WriteFile(openvpnConfigFile, []byte(configContent.String()), 0644); err != nil {
		userLog.Error("failed to write config file", "path", openvpnConfigFile, "error", err)
		os.Exit(1)
	}

	// Create VPN session
	session, err := client.CreateSession(ctx, user.ID, vpnIP, trustedIP)
	if err != nil {
		userLog.Warn("could not create session", "error", err)
	} else {
		// Save session ID for disconnect script
		sessionFile := filepath.Join(cfg.OpenVPN.SessionDir, fmt.Sprintf("session-%s", commonName))
		sessionData := fmt.Sprintf("%s\n%s\n%s", session.ID, trustedIP, trustedPort)
		if err := os.WriteFile(sessionFile, []byte(sessionData), 0600); err != nil {
			userLog.Warn("could not save session file", "path", sessionFile, "error", err)
		}
		userLog = userLog.WithSession(session.ID)
	}

	userLog.Info("client connected",
		"vpn_ip", vpnIP,
		"client_ip", trustedIP,
		"routes_count", len(networks),
		"default_route", hasDefaultRoute,
	)
	os.Exit(0)
}
