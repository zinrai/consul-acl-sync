package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

const (
	version = "0.1.0"
)

func main() {
	// Handle help and version flags first
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Check for help or version flags in the first argument
	firstArg := os.Args[1]
	if firstArg == "-h" || firstArg == "--help" || firstArg == "help" {
		printUsage()
		os.Exit(0)
	}

	if firstArg == "-v" || firstArg == "--version" || firstArg == "version" {
		fmt.Printf("consul-acl-sync version %s\n", version)
		os.Exit(0)
	}

	var (
		configPath string
		showHelp   bool
	)

	// Parse subcommand
	subcommand := os.Args[1]

	// Create flag set for subcommand
	flagSet := flag.NewFlagSet(subcommand, flag.ExitOnError)
	flagSet.StringVar(&configPath, "config", "consul-acl.yaml", "Path to configuration file")
	flagSet.BoolVar(&showHelp, "help", false, "Show help")

	// Custom usage for each subcommand
	flagSet.Usage = func() {
		switch subcommand {
		case "plan":
			fmt.Fprintf(os.Stderr, "Usage: consul-acl-sync plan [options]\n\n")
			fmt.Fprintf(os.Stderr, "Show what changes would be made without applying them.\n\n")
		case "apply":
			fmt.Fprintf(os.Stderr, "Usage: consul-acl-sync apply [options]\n\n")
			fmt.Fprintf(os.Stderr, "Apply the configuration changes to Consul.\n\n")
		default:
			printUsage()
			return
		}
		fmt.Fprintf(os.Stderr, "Options:\n")
		flagSet.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment variables:\n")
		fmt.Fprintf(os.Stderr, "  CONSUL_HTTP_ADDR    Consul server address (default: http://localhost:8500)\n")
		fmt.Fprintf(os.Stderr, "  CONSUL_HTTP_TOKEN   Consul ACL token for authentication\n")
	}

	// Parse flags for subcommand
	if err := flagSet.Parse(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	if showHelp {
		flagSet.Usage()
		os.Exit(0)
	}

	// Execute subcommand
	switch subcommand {
	case "plan", "apply":
		// Only check environment variables for actual operations
		consulAddr := os.Getenv("CONSUL_HTTP_ADDR")
		if consulAddr == "" {
			consulAddr = "http://localhost:8500"
		}

		consulToken := os.Getenv("CONSUL_HTTP_TOKEN")
		if consulToken == "" {
			fmt.Fprintf(os.Stderr, "Error: CONSUL_HTTP_TOKEN environment variable is required\n")
			fmt.Fprintf(os.Stderr, "Please set it to your Consul management token.\n")
			os.Exit(1)
		}

		// Load configuration
		config, err := LoadConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
			os.Exit(1)
		}

		// Create Consul client
		client := NewConsulClient(consulAddr, consulToken)

		// Execute the command
		if subcommand == "plan" {
			if err := runPlan(client, config); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		} else if subcommand == "apply" {
			if err := runApply(client, config); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", subcommand)
		printUsage()
		os.Exit(1)
	}
}

// runPlan executes the plan command
func runPlan(client *ConsulClient, config *Config) error {
	fmt.Printf("Loading configuration from: consul-acl.yaml\n")
	fmt.Printf("Consul server: %s\n", client.addr)
	fmt.Println()

	// Calculate differences
	diff, err := CalculateDiff(client, config)
	if err != nil {
		return fmt.Errorf("failed to calculate differences: %w", err)
	}

	// Print the plan
	PrintPlan(diff)

	return nil
}

// runApply executes the apply command
func runApply(client *ConsulClient, config *Config) error {
	fmt.Printf("Loading configuration from: consul-acl.yaml\n")
	fmt.Printf("Consul server: %s\n", client.addr)
	fmt.Println()

	// Calculate differences
	diff, err := CalculateDiff(client, config)
	if err != nil {
		return fmt.Errorf("failed to calculate differences: %w", err)
	}

	// Print the plan
	PrintPlan(diff)

	// Check if there are changes
	if !diff.HasChanges() {
		return nil
	}

	// Check for auto-approve flag
	autoApprove := false
	for _, arg := range os.Args {
		if arg == "-auto-approve" || arg == "--auto-approve" {
			autoApprove = true
			break
		}
	}

	// Ask for confirmation if not auto-approved
	if !autoApprove {
		if !ConfirmApply(diff) {
			fmt.Println("\nApply cancelled.")
			return nil
		}
	}

	// Apply the changes
	if err := Apply(client, diff); err != nil {
		return err
	}

	return nil
}

// printUsage prints the usage information
func printUsage() {
	usage := `consul-acl-sync - Synchronize Consul ACL configuration from YAML files

Usage:
  consul-acl-sync <command> [options]

Commands:
  plan     Show what changes would be made
  apply    Apply the configuration changes
  version  Show version information

Global Options:
  -config string    Path to configuration file (default: consul-acl.yaml)
  -help            Show help

Environment Variables:
  CONSUL_HTTP_ADDR    Consul server address (default: http://localhost:8500)
  CONSUL_HTTP_TOKEN   Consul ACL token for authentication (required)

Examples:
  # Show what changes would be made
  consul-acl-sync plan -config consul-acl.yaml

  # Apply configuration changes
  consul-acl-sync apply -config consul-acl.yaml

  # Apply changes without confirmation prompt
  consul-acl-sync apply -auto-approve

For more information, visit: https://github.com/zinrai/consul-acl-sync`

	fmt.Fprintln(os.Stderr, strings.TrimSpace(usage))
}
