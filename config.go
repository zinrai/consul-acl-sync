package main

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadConfig loads and parses the YAML configuration file
func LoadConfig(path string) (*Config, error) {
	// Set default path if empty
	if path == "" {
		path = "consul-acl.yaml"
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate configuration
	if err := ValidateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// ValidateConfig validates the configuration
func ValidateConfig(cfg *Config) error {
	// Check for duplicate policy names
	if err := validatePolicyNames(cfg.Policies); err != nil {
		return err
	}

	// Validate tokens
	if err := validateTokens(cfg.Tokens, cfg.Policies); err != nil {
		return err
	}

	return nil
}

// validatePolicyNames checks for duplicate policy names and required fields
func validatePolicyNames(policies []Policy) error {
	policyNames := make(map[string]bool)
	for _, policy := range policies {
		if policy.Name == "" {
			return fmt.Errorf("policy name cannot be empty")
		}
		if policyNames[policy.Name] {
			return fmt.Errorf("duplicate policy name: %s", policy.Name)
		}
		policyNames[policy.Name] = true

		if strings.TrimSpace(policy.Rules) == "" {
			return fmt.Errorf("policy %s has empty rules", policy.Name)
		}
	}
	return nil
}

// validateTokens checks token configuration
func validateTokens(tokens []Token, policies []Policy) error {
	for i, token := range tokens {
		if len(token.Policies) == 0 {
			return fmt.Errorf("token #%d (%s) has no policies", i+1, token.Description)
		}

		// Check if referenced policies exist in the config
		validateTokenPolicies(token, policies)
	}
	return nil
}

// validateTokenPolicies checks if token references valid policies
func validateTokenPolicies(token Token, policies []Policy) {
	for _, policyLink := range token.Policies {
		// Only validate name references (not IDs)
		if policyLink.Name == "" {
			continue
		}

		if !policyExists(policyLink.Name, policies) {
			// It might be a reference to an existing policy in Consul
			// We'll check this during the diff phase
			fmt.Fprintf(os.Stderr, "Warning: Token references policy '%s' which is not defined in this config\n", policyLink.Name)
		}
	}
}

// policyExists checks if a policy with the given name exists
func policyExists(name string, policies []Policy) bool {
	for _, policy := range policies {
		if policy.Name == name {
			return true
		}
	}
	return false
}

// NormalizeRules removes extra whitespace and standardizes HCL rules
func NormalizeRules(rules string) string {
	// Trim leading/trailing whitespace
	rules = strings.TrimSpace(rules)

	// Normalize line endings
	rules = strings.ReplaceAll(rules, "\r\n", "\n")

	// Remove trailing whitespace from each line
	lines := strings.Split(rules, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}

	return strings.Join(lines, "\n")
}

// PoliciesEqual compares two policies for equality
func PoliciesEqual(a, b Policy) bool {
	// Compare names
	if a.Name != b.Name {
		return false
	}

	// Compare descriptions
	if a.Description != b.Description {
		return false
	}

	// Compare rules (normalized)
	if NormalizeRules(a.Rules) != NormalizeRules(b.Rules) {
		return false
	}

	// Compare datacenters
	if len(a.Datacenters) != len(b.Datacenters) {
		return false
	}

	// Create maps for comparison
	dcMapA := make(map[string]bool)
	for _, dc := range a.Datacenters {
		dcMapA[dc] = true
	}

	for _, dc := range b.Datacenters {
		if !dcMapA[dc] {
			return false
		}
	}

	return true
}

// TokensEqual compares two tokens for equality
func TokensEqual(a, b Token) bool {
	// Compare descriptions
	if a.Description != b.Description {
		return false
	}

	// Compare policies
	if len(a.Policies) != len(b.Policies) {
		return false
	}

	// Create maps for policy comparison
	policiesA := make(map[string]bool)
	for _, p := range a.Policies {
		key := p.Name
		if key == "" {
			key = p.ID
		}
		policiesA[key] = true
	}

	for _, p := range b.Policies {
		key := p.Name
		if key == "" {
			key = p.ID
		}
		if !policiesA[key] {
			return false
		}
	}

	return true
}
