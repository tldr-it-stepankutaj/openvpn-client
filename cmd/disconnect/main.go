package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/tldr-it-stepankutaj/openvpn-client/internal/api"
	"github.com/tldr-it-stepankutaj/openvpn-client/internal/config"
	"github.com/tldr-it-stepankutaj/openvpn-client/internal/logger"
)

const programName = "openvpn-disconnect"

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

	// Get environment variables from OpenVPN
	commonName := os.Getenv("common_name")
	if commonName == "" {
		log.Error("common_name environment variable not set")
		os.Exit(1)
	}

	bytesReceived, _ := strconv.ParseInt(os.Getenv("bytes_received"), 10, 64)
	bytesSent, _ := strconv.ParseInt(os.Getenv("bytes_sent"), 10, 64)

	userLog := log.WithUser(commonName)

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		userLog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Read session file
	sessionFile := filepath.Join(cfg.OpenVPN.SessionDir, fmt.Sprintf("session-%s", commonName))
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		userLog.Warn("session file not found, nothing to disconnect", "path", sessionFile)
		os.Exit(0)
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) < 1 {
		userLog.Warn("invalid session file format")
		os.Exit(0)
	}

	sessionID := strings.TrimSpace(lines[0])
	userLog = userLog.WithSession(sessionID)

	// Create API client
	client := api.NewClient(&cfg.API)

	ctx := context.Background()

	// Authenticate if using legacy service account
	if !cfg.API.UseToken() {
		if err := client.Authenticate(ctx, cfg.API.Username, cfg.API.Password); err != nil {
			userLog.Error("API authentication failed", "error", err)
			os.Exit(0)
		}
	}

	// End session
	if err := client.DisconnectSession(ctx, sessionID, bytesReceived, bytesSent); err != nil {
		userLog.Warn("could not end session", "error", err)
	}

	// Remove session file
	if err := os.Remove(sessionFile); err != nil {
		userLog.Warn("could not remove session file", "path", sessionFile, "error", err)
	}

	userLog.Info("client disconnected",
		"bytes_received", bytesReceived,
		"bytes_sent", bytesSent,
	)
	os.Exit(0)
}
