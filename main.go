package main

import (
	"flag"
	"fmt"
	"os"
)

// Injected at build time by goreleaser via -ldflags -X.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "consul-acl-sync:", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		configPath  string
		consulAddr  string
		showVersion bool
	)
	flag.StringVar(&configPath, "config", "", "path to configuration file (required)")
	flag.StringVar(&consulAddr, "consul-addr", "http://127.0.0.1:8500", "Consul HTTP API address")
	flag.BoolVar(&showVersion, "version", false, "print version and exit")
	flag.Parse()

	if showVersion {
		fmt.Printf("consul-acl-sync %s (commit %s, built %s)\n", version, commit, date)
		return nil
	}

	if configPath == "" {
		return fmt.Errorf("-config is required")
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		return err
	}

	client := NewConsulClient(consulAddr, os.Getenv("CONSUL_HTTP_TOKEN"))
	plan, err := CalculatePlan(client, cfg)
	if err != nil {
		return err
	}

	if !plan.HasChanges() {
		fmt.Println("No changes. Consul is up to date.")
		return nil
	}

	if err := Apply(client, plan); err != nil {
		return err
	}

	fmt.Printf("\nApplied: policies %d created, %d updated; tokens %d created, %d updated.\n",
		len(plan.PoliciesToCreate), len(plan.PoliciesToUpdate),
		len(plan.TokensToCreate), len(plan.TokensToUpdate))
	return nil
}
