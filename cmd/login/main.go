package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"strings"

	"github.com/tldr-it-stepankutaj/openvpn-client/internal/api"
	"github.com/tldr-it-stepankutaj/openvpn-client/internal/config"
	"github.com/tldr-it-stepankutaj/openvpn-client/internal/logger"
)

const programName = "openvpn-login"

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

	// Get auth file path from positional argument
	args := flag.Args()
	if len(args) < 1 {
		log.Error("password file path not provided")
		os.Exit(1)
	}

	authFile := args[0]

	// Read the credential file
	data, err := os.ReadFile(authFile)
	if err != nil {
		log.Error("failed to read auth file", "path", authFile, "error", err)
		os.Exit(1)
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) < 2 {
		log.Error("invalid auth file format", "path", authFile)
		os.Exit(1)
	}

	username := strings.TrimSpace(lines[0])
	password := strings.TrimSpace(lines[1])

	userLog := log.WithUser(username)

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		userLog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Create API client
	client := api.NewClient(&cfg.API)

	ctx := context.Background()

	// Validate credentials
	authResp, err := client.ValidateVpnUser(ctx, username, password)
	if err != nil {
		userLog.Error("authentication error", "error", err)
		os.Exit(1)
	}

	if !authResp.Valid {
		if authResp.StatusCode == 429 {
			userLog.Warn("authentication rejected", "reason", "rate_limited_or_locked", "message", authResp.Message)
		} else {
			userLog.Warn("authentication failed", "message", authResp.Message)
		}
		os.Exit(1)
	}

	userLog.Info("user authenticated successfully", "user_id", authResp.User.ID)
	os.Exit(0)
}
