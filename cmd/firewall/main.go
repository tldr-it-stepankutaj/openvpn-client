package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/exec"

	"github.com/tldr-it-stepankutaj/openvpn-client/internal/api"
	"github.com/tldr-it-stepankutaj/openvpn-client/internal/config"
	"github.com/tldr-it-stepankutaj/openvpn-client/internal/firewall"
	"github.com/tldr-it-stepankutaj/openvpn-client/internal/logger"
)

const programName = "openvpn-firewall"

func main() {
	var (
		configPath string
		dryRun     bool
	)
	flag.StringVar(&configPath, "config", "", "path to configuration file")
	flag.StringVar(&configPath, "c", "", "path to configuration file (shorthand)")
	flag.BoolVar(&dryRun, "dry-run", false, "print rules without applying")
	flag.BoolVar(&dryRun, "n", false, "print rules without applying (shorthand)")
	flag.Parse()

	// Initialize logger
	log := logger.New(logger.Options{
		Level:   slog.LevelInfo,
		JSON:    true,
		Program: programName,
	})

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Create API client
	client := api.NewClient(&cfg.API)

	ctx := context.Background()

	// Authenticate if using a legacy service account
	if !cfg.API.UseToken() {
		if err := client.Authenticate(ctx, cfg.API.Username, cfg.API.Password); err != nil {
			log.Error("API authentication failed", "error", err)
			os.Exit(1)
		}
	}

	// Get all active users
	users, err := client.GetAllActiveUsers(ctx)
	if err != nil {
		log.Error("failed to get users", "error", err)
		os.Exit(1)
	}

	log.Info("fetched active users", "count", len(users))

	// Collect networks for each user
	usersWithNetworks, err := firewall.CollectUserNetworks(ctx, client, users)
	if err != nil {
		log.Error("failed to collect networks", "error", err)
		os.Exit(1)
	}

	log.Info("collected user networks", "users_with_rules", len(usersWithNetworks))

	// Create a firewall generator
	fw := firewall.New(&cfg.Firewall)

	// Generate rules
	newRules := fw.GenerateRules(usersWithNetworks)

	// Dry run - just print rules
	if dryRun {
		log.Info("dry run mode - printing rules")
		_, err := os.Stdout.WriteString(newRules)
		if err != nil {
			return
		}
		os.Exit(0)
	}

	// Check if rules changed
	rulesFile := fw.GetRulesFile()
	oldRules, _ := os.ReadFile(rulesFile)
	if string(oldRules) == newRules {
		log.Info("firewall rules unchanged", "file", rulesFile)
		os.Exit(0)
	}

	// Write new rules
	if err := os.WriteFile(rulesFile, []byte(newRules), 0644); err != nil {
		log.Error("failed to write rules file", "file", rulesFile, "error", err)
		os.Exit(1)
	}

	log.Info("wrote firewall rules", "file", rulesFile)

	// Reload firewall
	reloadCmd := fw.GetReloadCommand()
	cmd := exec.Command("sh", "-c", reloadCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error("failed to reload firewall",
			"command", reloadCmd,
			"output", string(output),
			"error", err,
		)
		os.Exit(1)
	}

	log.Info("firewall rules updated",
		"users", len(usersWithNetworks),
		"type", cfg.Firewall.Type,
		"file", rulesFile,
	)
	os.Exit(0)
}
